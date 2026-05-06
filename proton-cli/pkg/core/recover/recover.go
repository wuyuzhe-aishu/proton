package recover

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/backup"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/etcd"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/file"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/shellcommand"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/tar"

	"github.com/jhunters/goassist/arrayutil"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/exp/slices"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

const (
	RecoverDir      string = "/opt/proton-backup"
	RecoverLogDir   string = "/opt/proton-backup/recoverlogs"
	RecoverConfName string = "recover.json"
	//k8s 的etcd快照备份目录
	EtcdSnapshotRelPath string = "k8s/etcdSnapshot"
	//异步恢复数据服务，循环状态接口次数
	RecoverCycleCount int = 60

	//异步恢复数据服务恢复后，检测循环状态接口次数
	RecoverCheckCycleCount int = 2880
)

type RecoverOpts struct {
	Resource            []string
	Id                  string
	RecoverName         string
	FromBackup          string
	SkipRecoverResource []string
}
type RecoverConf struct {
	HostName         string        `json:"hostName"`
	RecoverDirectory string        `json:"recoverDirectory"`
	List             []RecoverInfo `json:"list"`
}
type RecoverInfo struct {
	Id                 string   `json:"id"`
	Name               string   `json:"name"`
	CreateTime         int64    `json:"createTime"`
	EndTime            int64    `json:"endTime"`
	RunTime            int64    `json:"runTime"`
	FromBackupId       string   `json:"fromBackupId"`
	FromBackupPath     string   `json:"fromBackupPath"`
	FromBackupResource []string `json:"fromBackupResource"`
	LogPath            string   `json:"logPath"`
	Status             bool     `json:"status"`
	Resource           []string `json:"resource"`
}

// mariadb 请求还原返回信息
type MariaDBRecoverResponse struct {
	Id          string `json:"id"`
	PackageName string `json:"package_name"`
	CreateTime  string `json:"create_time"`
	Status      string `json:"status"`
	StorageNode string `json:"storage_node"`
}

// mariadb 查看还原状态返回信息
type MariaDBRecoverStatusResponse struct {
	Msg    string `json:"msg"`
	Status string `json:"status"`
}

// MariaDB请求还原参数
type MariaDBRecoverRequest struct {
	File string `json:"file"`
}

// Mongo请求还原参数
type MongoDBRecoverRequest struct {
	Backup_dir string `json:"backup_dir"`
}

// MongoDB 请求还原返回信息
type MongoDBRecoverResponse struct {
	Cause string `json:"cause"`
	Code  string `json:"code"`
	Msg   string `json:"msg"`
}

// MongoDB 查看还原状态返回信息
type MongoDBRecoverStatusResponse struct {
	Cause  string `json:"cause"`
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
}

var (
	Recoverlog = logger.NewLogger()
	//需要还原的的资源集合
	RecoverResourceCollection = []backup.ResourceInfo{}
)

// 检测配置文件是否存在，不存在创建
func CheckRecoverConfig() error {
	exist, err := file.PathExists(RecoverLogDir)
	if err != nil {
		Recoverlog.Errorln("检测还原配置日志目录是否存在异常：", RecoverLogDir, err)
		return err
	}
	if !exist {
		err := os.MkdirAll(RecoverLogDir, os.ModePerm)
		if err != nil {
			Recoverlog.Errorln("创建还原配置目录异常：", RecoverLogDir, err)
			return err
		}
	}
	var recoverConf = filepath.Join(RecoverDir, RecoverConfName)
	exist, err = file.PathExists(recoverConf)
	if err != nil {
		Recoverlog.Errorln("检测还原配置文件是否存在异常：", recoverConf, err)
		return err
	}
	if !exist {
		f, err := os.Create(recoverConf)
		f.Close()
		if err != nil {
			Recoverlog.Errorln("创建还原配置文件异常：", recoverConf, err)
			return err
		}
	}
	return nil
}

// 获取proton的还原配置对象
func GetRecoverConf() (*RecoverConf, error) {
	err := CheckRecoverConfig()
	if err != nil {
		return nil, err
	}

	var backupConf = filepath.Join(RecoverDir, RecoverConfName)
	hostname, _ := os.Hostname()
	var conf = RecoverConf{}
	content, err := os.ReadFile(backupConf)
	if err != nil {
		Recoverlog.Errorln("打开还原配置文件异常：", backupConf, err)
		return nil, err
	}
	if len(content) == 0 {
		conf.HostName = hostname
		conf.RecoverDirectory = "/mnt/recover"
	} else {
		err = json.Unmarshal([]byte(string(content)), &conf)
		if err != nil {
			return nil, err
		}
	}
	exist, err := file.PathExists(conf.RecoverDirectory)
	if err != nil {
		Recoverlog.Errorln("检测还原配置解压目录是否存在异常：", conf.RecoverDirectory, err)
		return nil, err
	}
	if !exist {
		err := os.MkdirAll(conf.RecoverDirectory, os.ModePerm)
		if err != nil {
			Recoverlog.Errorln("创建还原配置解压目录异常：", conf.RecoverDirectory, err)
			return nil, err
		}
	}
	return &conf, nil
}

// 数组删除重复
func removeDuplicateElement(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// 创建还原
func CreateRecover(opt RecoverOpts) error {
	conf, err := GetRecoverConf()
	if err != nil {
		return err
	}
	var info = RecoverInfo{
		Id:         opt.Id,
		Name:       opt.RecoverName,
		CreateTime: time.Now().Unix(),
		Status:     false,
	}
	var existinfo = arrayutil.Filter(conf.List, func(s1 RecoverInfo) bool { return s1.Name != info.Name })
	if len(existinfo) > 0 {
		return errors.New("duplicate recover name:" + info.Name)
	}

	opt.Resource = removeDuplicateElement(opt.Resource)

	backupconf, err := backup.GetBackupConf()
	if err != nil {
		return err
	}
	var backupinfo = arrayutil.Filter(backupconf.List, func(s1 backup.BackupInfo) bool { return s1.Name != opt.FromBackup })
	if len(backupinfo) > 0 {
		info.FromBackupId = backupinfo[0].Id
		info.FromBackupPath = backupinfo[0].Path
		info.FromBackupResource = backupinfo[0].Resource
	} else {
		return errors.New("Invalid backup name specified for restore:" + opt.FromBackup)
	}

	var recoverError = RecoverResource(opt, conf, &info)
	info.EndTime = time.Now().Unix()
	info.RunTime = int64(time.Unix(info.EndTime, 0).Sub(time.Unix(info.CreateTime, 0)).Seconds())
	// info.FromBackupPath = filepath.Join(conf.RecoverDirectory, opt.Id+".tar.gz")
	info.LogPath = filepath.Join(RecoverLogDir, opt.Id+".log")
	// info.Resource = opt.Resource
	if recoverError != nil {
		info.Status = false
		return recoverError
	} else {
		info.Status = true
		Recoverlog.Info("还原成功")
	}
	conf.List = append(conf.List, info)
	jsonBytes, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(RecoverDir, RecoverConfName), jsonBytes, 0644)
	if err != nil {
		return err
	}
	Recoverlog.Info("保存还原记录")
	return nil

}

//还原当前节点的资源

func RecoverResource(opt RecoverOpts, conf *RecoverConf, fo *RecoverInfo) error {
	//过滤出需要还原的资源,如果是all，还原所有备份的资源
	if !backup.IsContain(opt.Resource, "all") {
		for _, col := range backup.ResourceCollection {
			if backup.IsContain(opt.Resource, col.Name) {
				RecoverResourceCollection = append(RecoverResourceCollection, col)
			}
		}
	} else {
		for _, col := range backup.ResourceCollection {
			if backup.IsContain(fo.FromBackupResource, col.Name) {
				RecoverResourceCollection = append(RecoverResourceCollection, col)
			}
		}
	}
	//删除需要跳过的还原资源
	if len(opt.SkipRecoverResource) > 0 {
		var newResourceCollection = []backup.ResourceInfo{}
		for _, col := range RecoverResourceCollection {
			if !backup.IsContain(opt.SkipRecoverResource, col.Name) {
				newResourceCollection = append(newResourceCollection, col)
			}
		}
		RecoverResourceCollection = newResourceCollection
	}
	if len(RecoverResourceCollection) == 0 {
		return errors.New("请求还原的资源无效:" + strings.Join(opt.Resource, ","))
	}

	//从备份记录中获取备份资源列表
	var BackupResourceCollection = []backup.ResourceInfo{}
	if len(fo.FromBackupResource) == 0 {
		return errors.New("There is no backup resource in the backup package:" + fo.FromBackupId)
	} else {
		//获取备份包中的资源列表
		if !backup.IsContain(fo.FromBackupResource, "all") {
			for _, col := range backup.ResourceCollection {
				if backup.IsContain(fo.FromBackupResource, col.Name) {
					BackupResourceCollection = append(BackupResourceCollection, col)
				}
			}
		} else {
			BackupResourceCollection = backup.ResourceCollection
		}
		for _, res := range RecoverResourceCollection {
			var backup_res = arrayutil.Filter(BackupResourceCollection, func(s1 backup.ResourceInfo) bool { return s1.Name != res.Name })
			if len(backup_res) == 0 {
				return errors.New(res.Name + " of the resource to be restored is not in the backup package:" + opt.FromBackup)
			}
		}
	}
	// 用id排序,数值小的排在前
	sort.SliceStable(RecoverResourceCollection, func(i, j int) bool {
		return RecoverResourceCollection[i].RestoreId < RecoverResourceCollection[j].RestoreId
	})
	Recoverlog.Info("解压备份包")
	if err := tar.DecompressTarball(fo.FromBackupPath, conf.RecoverDirectory); err != nil {
		return err
	}
	//备份资源目录
	var workDirectory = filepath.Join(conf.RecoverDirectory, fo.FromBackupId)

	var clusterConf *configuration.ClusterConfig
	var k *kubernetes.Clientset

	_, k = client.NewK8sClient()
	if k == nil {
		var configPath = filepath.Join(workDirectory, "k8s/kube/config")
		if exist, err := file.PathExists(configPath); err != nil {
			return err
		} else if !exist {
			return client.ErrKubernetesClientSetNil
		}
		config, err := clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return client.ErrKubernetesClientSetNil
		}
		k, _ = kubernetes.NewForConfig(config)
	}
	clusterConf, err := configuration.LoadFromFile(filepath.Join(workDirectory, backup.ProtonCLIConfName))
	if err != nil {
		return err
	}
	if clusterConf == nil {
		return fmt.Errorf("Unable to get cluster configuration file")
	}

	var NamespaceResource = configuration.GetProtonResourceNSFromFile()
	for _, info := range RecoverResourceCollection {
		Recoverlog.Info("还原资源：" + info.Name)
		var skipRecover = false
		if info.Name == "proton-mariadb" {
			for _, p := range info.PathList {
				if clusterConf == nil {
					return fmt.Errorf("Unable to get cluster configuration file")
				} else if clusterConf.Proton_mariadb == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-mariadb configuration")
					}
					Recoverlog.Info("Unable to get Proton-mariadb configuration,skip recover")
					skipRecover = true
					continue
				} else if clusterConf.Proton_mariadb.Hosts == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-mariadb Hosts configuration")
					}
					Recoverlog.Info("Unable to get Proton-mariadb Hosts configuration,skip recover")
					skipRecover = true
					continue
				} else if len(clusterConf.Proton_mariadb.Hosts) == 0 {
					if p.Require {
						return fmt.Errorf("Proton-mariadb the number of copies is %d, and mariadb data cannot be recover", len(clusterConf.Proton_mariadb.Hosts))
					}
					Recoverlog.Infof("Proton-mariadb the number of copies is %d, and mariadb data cannot be recover,skip recover", len(clusterConf.Proton_mariadb.Hosts))
					skipRecover = true
					continue
				} else if !slices.Contains(clusterConf.Proton_mariadb.Hosts, conf.HostName) {
					if p.Require {
						return fmt.Errorf("The current node not have proton-mariadb")
					}
					Recoverlog.Infoln("The current node not have proton-mariadb,skip recover")
					skipRecover = true
					continue
				} else if clusterConf.Proton_mariadb.Admin_user == "" || clusterConf.Proton_mariadb.Admin_passwd == "" {
					return fmt.Errorf("Proton-mariadb username or password cannot be empty")
				}

				// 需要恢复的数据组件-已正常启动和访问
				for count := 1; count <= RecoverCycleCount; count++ {
					if count == RecoverCycleCount {
						return fmt.Errorf("mariadb-mgmt service is not ready,checked %d count,timeout", count)
					}
					time.Sleep(time.Second * 10)

					// 检测mariadb-mgmt服务是否正常
					mgmt, err := k.AppsV1().Deployments(NamespaceResource).Get(context.Background(), "mariadb-mgmt", metav1.GetOptions{})
					if err != nil {
						Recoverlog.Infof("can not get mariadb-mgmt Deployments")
						continue
					}
					if mgmt == nil || *mgmt.Spec.Replicas != mgmt.Status.ReadyReplicas {
						Recoverlog.Infof("mariadb-mgmt pod is not ready")
						continue
					}
					svc, err := k.CoreV1().Services(NamespaceResource).Get(context.Background(), backup.ProtonMariaDBServiceName, metav1.GetOptions{})
					if err != nil {
						Recoverlog.Infof("can not get mariadb-mgmt-cluster Services")
						continue

					} else if svc == nil {
						Recoverlog.Infof("unable to get mariadb service:" + backup.ProtonMariaDBServiceName)
						continue
					}
					var ip = svc.Spec.ClusterIP
					if ip == "" {
						Recoverlog.Infof("Unable to obtain the management terminal address of mariadb")
						continue
					}
					headers := map[string]string{
						"admin-key": base64.StdEncoding.EncodeToString([]byte(clusterConf.Proton_mariadb.Admin_user + ":" + clusterConf.Proton_mariadb.Admin_passwd)),
					}
					var httpclient = client.NewHttpClient(30)
					if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:8888/api/proton-rds-mgmt/v2/backups", ip), headers); err != nil {
						Recoverlog.Infof("unable to get mariadb backup message, error: %s", err.Error())
					} else if s != http.StatusOK {
						Recoverlog.Infof("unable to get mariadb backup message, http status code: %d, response body: %v", s, b)
					} else {
						goto PROTON_MARIADB_READY
					}

				}
			PROTON_MARIADB_READY:
				Recoverlog.Info("mariadb-mgmt service is ready")

				svc, err := k.CoreV1().Services(NamespaceResource).Get(context.Background(), backup.ProtonMariaDBServiceName, metav1.GetOptions{})
				if err != nil {
					return err
				} else if svc == nil {
					return fmt.Errorf("unable to get mariadb service:" + backup.ProtonMariaDBServiceName)
				}
				var ip = svc.Spec.ClusterIP
				if ip == "" {
					return errors.New("Unable to obtain the management terminal address of mariadb")
				}
				headers := map[string]string{
					"admin-key": base64.StdEncoding.EncodeToString([]byte(clusterConf.Proton_mariadb.Admin_user + ":" + clusterConf.Proton_mariadb.Admin_passwd)),
				}
				var httpclient = client.NewHttpClient(30)
				var mariadbRequest = MariaDBRecoverRequest{}
				var backupPath = filepath.Join(workDirectory, p.TargetRelPath)
				files, err := filepath.Glob(backupPath + "/*.gz")
				if err != nil {
					return err
				} else if len(files) == 0 {
					return fmt.Errorf("not found proton-mariadb backup file : %s", backupPath)
				}
				mariadbRequest.File = files[0]
				Recoverlog.Info(mariadbRequest.File)
				var url = fmt.Sprintf("http://%s:8888/api/proton-rds-mgmt/v2/restorations", ip)
				Recoverlog.Info(url)
				if s, b, err := httpclient.Post(url, headers, mariadbRequest); err != nil {
					return fmt.Errorf("unable to create mariadb recover, error: %w", err)
				} else if s != http.StatusAccepted {
					return fmt.Errorf("unable to create mariadb recover, http status code: %d, response body: %v", s, b)
				} else {
					time.Sleep(time.Second * 10)
					Recoverlog.Printf("create mariadb recover, http status code: %d, response body: %v \n", s, b)
					responseByte, err := jsoniter.Marshal(b)
					if err != nil {
						Recoverlog.Info("unable to create mariadb recover,The returned data cannot be converted to byte")
						return err
					}
					info := MariaDBRecoverResponse{}
					err = jsoniter.Unmarshal(responseByte, &info)
					if err != nil {
						Recoverlog.Info("unable to create mariadb recover,The returned data cannot be converted to MariaDBRecoverResponse")
						return err
					}
					for i := 1; i <= RecoverCheckCycleCount; i++ {
						if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:8888/api/proton-rds-mgmt/v2/restorations", ip), headers); err != nil {
							return fmt.Errorf("unable to get mariadb recover message, error: %w", err)
						} else if s != http.StatusOK {
							return fmt.Errorf("unable to get mariadb recover message, http status code: %d, response body: %v", s, b)
						} else {
							Recoverlog.Printf("get mariadb recover message, http status code: %d, response body: %v \n", s, b)
							responseByte, err := jsoniter.Marshal(b)
							if err != nil {
								Recoverlog.Info("unable to get mariadb recover message,The returned data cannot be converted to byte")
								return err
							}
							list := MariaDBRecoverStatusResponse{}
							err = jsoniter.Unmarshal(responseByte, &list)
							if err != nil {
								Recoverlog.Info("unable to get mariadb recover message,The returned data cannot be converted to MariaDBRecoverStatusResponse")
								return err
							}
							if list.Status == "success" {
								goto MariadbSuccess
							}
							if list.Status == "failed" {
								return fmt.Errorf("backup mariadb failed, http status code: %d, response body: %v", s, b)
							}
						}
						if i == RecoverCheckCycleCount {
							return fmt.Errorf("unable to get mariadb recover message：checked %d count", i)
						}
						time.Sleep(time.Second * 30)
					}
				}
			MariadbSuccess:
				Recoverlog.Info("mariadb recover success")
			}
		} else if info.Name == "proton-mongodb" {

			for _, p := range info.PathList {
				if clusterConf == nil {
					return fmt.Errorf("Unable to get cluster configuration file")
				} else if clusterConf.Proton_mongodb == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-mongodb configuration")
					}
					Recoverlog.Info("Unable to get Proton-mongodb configuration,skip recover")
					skipRecover = true
					continue
				} else if clusterConf.Proton_mongodb.Hosts == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-mongodb Hosts configuration")
					}
					Recoverlog.Info("Unable to get Proton-mongodb Hosts configuration,skip recover")
					skipRecover = true
					continue
				} else if len(clusterConf.Proton_mongodb.Hosts) == 0 {
					if p.Require {
						return fmt.Errorf("Proton-mongodb the number of copies is %d, and mongodb data cannot be backed up", len(clusterConf.Proton_mongodb.Hosts))
					}
					Recoverlog.Infof("Proton-mongodb the number of copies is %d, and mongodb data cannot be backed up,skip recover", len(clusterConf.Proton_mongodb.Hosts))
					skipRecover = true
					continue
				} else if !slices.Contains(clusterConf.Proton_mongodb.Hosts, conf.HostName) {
					if p.Require {
						return fmt.Errorf("The current node not have proton-mongodb")
					}
					Recoverlog.Infoln("The current node not have proton-mongodb,skip recover")
					skipRecover = true
					continue
				} else if clusterConf.Proton_mongodb.Admin_user == "" || clusterConf.Proton_mongodb.Admin_passwd == "" {
					return fmt.Errorf("Proton-mongodb username or password cannot be empty")
				}

				// 需要恢复的数据组件-已正常启动和访问
				for count := 1; count <= RecoverCycleCount; count++ {
					if count == RecoverCycleCount {
						return fmt.Errorf("mongodb-mongodb-mgmt service is not ready,checked %d count,timeout", count)
					}
					time.Sleep(time.Second * 10)
					// 检测mongodb-mongodb-mgmt服务是否正常
					mgmt, err := k.AppsV1().StatefulSets(NamespaceResource).Get(context.Background(), "mongodb-mongodb-mgmt", metav1.GetOptions{})
					if err != nil {
						Recoverlog.Infof("can not get mongodb-mongodb-mgmt StatefulSet")
						continue
					}
					if mgmt == nil || *mgmt.Spec.Replicas != mgmt.Status.ReadyReplicas {
						Recoverlog.Infof("mongodb-mongodb-mgmt pod is not ready")
						continue
					}
					svc, err := k.CoreV1().Services(NamespaceResource).Get(context.Background(), backup.ProtonMongoDBServiceName, metav1.GetOptions{})
					if err != nil {
						Recoverlog.Infof("can not get mongodb-mgmt-cluster Services")
						continue
					} else if svc == nil {
						Recoverlog.Infof("unable to get mongodb service:" + backup.ProtonMongoDBServiceName)
						continue
					}
					var ip = svc.Spec.ClusterIP
					if ip == "" {
						Recoverlog.Infof("Unable to obtain the management terminal address of mongodb")
						continue
					}
					headers := map[string]string{
						"admin-key": base64.StdEncoding.EncodeToString([]byte(clusterConf.Proton_mongodb.Admin_user + ":" + clusterConf.Proton_mongodb.Admin_passwd)),
					}
					var httpclient = client.NewHttpClient(30)
					if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:30281/api/proton-mongodb-mgmt/v2/backups", ip), headers); err != nil {
						Recoverlog.Infof("unable to get mongodb backup message, error: %s", err.Error())
					} else if s != http.StatusOK {
						Recoverlog.Infof("unable to get mongodb backup message, http status code: %d, response body: %v", s, b)
					} else {
						goto PROTON_MONGODB_READY
					}
				}

			PROTON_MONGODB_READY:
				Recoverlog.Info("mongodb-mgmt service is ready")

				svc, err := k.CoreV1().Services(NamespaceResource).Get(context.Background(), backup.ProtonMongoDBServiceName, metav1.GetOptions{})
				if err != nil {
					return err
				} else if svc == nil {
					return fmt.Errorf("unable to get service:" + backup.ProtonMongoDBServiceName)
				}
				var ip = svc.Spec.ClusterIP
				if ip == "" {
					return errors.New("Unable to obtain the management terminal address of mongodb")
				}
				headers := map[string]string{
					"admin-key": base64.StdEncoding.EncodeToString([]byte(clusterConf.Proton_mongodb.Admin_user + ":" + clusterConf.Proton_mongodb.Admin_passwd)),
				}
				var httpclient = client.NewHttpClient(30)
				var mongodbRequest = MongoDBRecoverRequest{}

				var backupPath = filepath.Join(workDirectory, p.TargetRelPath, "mongodb/rs0")
				files, err := filepath.Glob(backupPath + "/*.tar")
				if err != nil {
					return err
				} else if len(files) == 0 {
					return fmt.Errorf("not found proton-mongodb backup file : %s", backupPath)
				}
				mongodbRequest.Backup_dir = files[0]
				Recoverlog.Info(mongodbRequest.Backup_dir)
				var url = fmt.Sprintf("http://%s:30281/api/proton-mongodb-mgmt/v2/db_restore", ip)
				Recoverlog.Info(url)
				if s, b, err := httpclient.Post(url, headers, mongodbRequest); err != nil {
					return fmt.Errorf("unable to create mongodb recover, error: %w", err)
				} else if s != http.StatusOK {
					return fmt.Errorf("unable to create mongodb recover, http status code: %d, response body: %v", s, b)
				} else {
					time.Sleep(time.Second * 10)
					Recoverlog.Printf("create mongodb recover, http status code: %d, response body: %v \n", s, b)
					responseByte, err := jsoniter.Marshal(b)
					if err != nil {
						Recoverlog.Info("unable to create mongodb recover,The returned data cannot be converted to byte")
						return err
					}
					info := MongoDBRecoverResponse{}
					err = jsoniter.Unmarshal(responseByte, &info)
					if err != nil {
						Recoverlog.Info("unable to create mongodb recover,The returned data cannot be converted to MongoDBRecoverResponse")
						return err
					}
					for i := 1; i <= RecoverCheckCycleCount; i++ {
						if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:30281/api/proton-mongodb-mgmt/v2/restore_status", ip), headers); err != nil {
							return fmt.Errorf("unable to get mongodb recover message, error: %w", err)
						} else if s != http.StatusOK {
							return fmt.Errorf("unable to get mongodb recover message, http status code: %d, response body: %v", s, b)
						} else {
							Recoverlog.Printf("get mongodb recover message, http status code: %d, response body: %v \n", s, b)
							responseByte, err := jsoniter.Marshal(b)
							if err != nil {
								Recoverlog.Info("unable to get mongodb recover message,The returned data cannot be converted to byte")
								return err
							}
							list := MongoDBRecoverStatusResponse{}
							err = jsoniter.Unmarshal(responseByte, &list)
							if err != nil {
								Recoverlog.Info("unable to get mongodb recover message,The returned data cannot be converted to MongoDBRecoverStatusResponse")
								return err
							}
							if list.Status == "success" {
								goto MongodbSuccess
							}
							if list.Status == "failed" {
								return fmt.Errorf("recover mongodb failed, http status code: %d, response body: %v", s, b)
							}
						}
						if i == RecoverCheckCycleCount {
							return fmt.Errorf("unable to get mongodb recover message：checked %d count", i)
						}
						time.Sleep(time.Second * 30)
					}
				}
			MongodbSuccess:
				Recoverlog.Info("mongodb recover success")
			}
		} else if info.Name == "proton-etcd" {

			for _, p := range info.PathList {
				if clusterConf == nil {
					return fmt.Errorf("Unable to get cluster configuration file")
				} else if clusterConf.Proton_etcd == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-etcd configuration")
					}
					Recoverlog.Infof("Unable to get Proton-etcd configuration,skip recover proton-etcd")
					skipRecover = true
					continue
				} else if len(clusterConf.Proton_etcd.Hosts) == 0 {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-etcd Hosts configuration")
					}
					Recoverlog.Infof("Unable to get Proton-etcd Hosts configuration,skip recover proton-etcd")
					skipRecover = true
					continue
				} else if len(clusterConf.Proton_etcd.Hosts) == 0 || len(clusterConf.Proton_etcd.Hosts) > 1 {
					if p.Require {
						return fmt.Errorf("Proton-etcd the number of copies is %d, and Proton etcd data cannot be recover", len(clusterConf.Proton_etcd.Hosts))
					}
					Recoverlog.Infof("Proton-etcd the number of copies is %d, and Proton etcd data cannot be recover,skip recover proton-etcd", len(clusterConf.Proton_etcd.Hosts))
					skipRecover = true
					continue
				}
				// 需要恢复的数据组件-已正常启动和访问
				for count := 1; count <= RecoverCycleCount; count++ {
					// 检测proton-etcd服务是否正常
					sts, err := k.AppsV1().StatefulSets(NamespaceResource).Get(context.Background(), "proton-etcd", metav1.GetOptions{})
					if err != nil {
						Recoverlog.Infoln(err.Error())
					}
					if sts == nil || *sts.Spec.Replicas != sts.Status.ReadyReplicas {
						Recoverlog.Infoln("proton-etcd pod is not ready")
					} else {
						goto PROTON_ETCD_READY
					}
					if count == RecoverCycleCount {
						return fmt.Errorf("proton-etcd data service is not ready,checked %d count,timeout", count)
					}
					time.Sleep(time.Second * 10)
				}
			PROTON_ETCD_READY:
				Recoverlog.Info("proton-etcd data service is ready")

				var backupPath = filepath.Join(workDirectory, p.TargetRelPath)
				files, err := filepath.Glob(backupPath + "/*.db")
				if err != nil {
					return err
				} else if len(files) == 0 {
					return fmt.Errorf("not found proton-etcd backup file : %s", backupPath)
				}
				var etcdSnapshotPath = files[0]
				var destpath = filepath.Join(clusterConf.Proton_etcd.Data_path, "data")
				if exits, err := file.PathExists(destpath); err != nil {
					return err
				} else if exits {
					var destBackupPath = clusterConf.Proton_etcd.Data_path + time.Now().Format("20060102150405") + strconv.Itoa(time.Now().Nanosecond())
					err = file.Copy(destpath, destBackupPath, Recoverlog, nil)
					if err != nil {
						return err
					}
					err = os.RemoveAll(destpath)
					if err != nil {
						return err
					}
				} else {
					err := os.MkdirAll(destpath, os.ModePerm)
					if err != nil {
						Recoverlog.Errorln("创建proton etcd还原的目录异常：" + err.Error())
						return err
					}
				}
				proton_etcd_pod_files, err := filepath.Glob(backupPath + "/*.yaml")
				if err != nil {
					return err
				} else if len(proton_etcd_pod_files) == 0 {
					return fmt.Errorf("not found proton-etcd pod yaml file : %s", backupPath)
				}
				for _, yamlPath := range proton_etcd_pod_files {
					etcd_cfg := new(core_v1.Pod)
					bytes, err := os.ReadFile(yamlPath)
					if err != nil {
						return err
					}
					if err := yaml.Unmarshal(bytes, etcd_cfg); err != nil {
						return err
					}
					if etcd_cfg == nil || etcd_cfg.Spec.Containers == nil {
						return fmt.Errorf("Unable to create proton etcd pod struct: %s", yamlPath)
					}
					var etcd_initial_cluster = arrayutil.Filter(etcd_cfg.Spec.Containers[0].Env, func(s1 core_v1.EnvVar) bool {
						return !strings.EqualFold(s1.Name, "ETCD_ALL_ENDPOINTS")

					})
					var etcd_initial_advertise_peer_urls = arrayutil.Filter(etcd_cfg.Spec.Containers[0].Env, func(s1 core_v1.EnvVar) bool {
						return !strings.EqualFold(s1.Name, "ETCD_INITIAL_ADVERTISE_PEER_URLS")
					})

					if len(etcd_initial_cluster) == 0 || len(etcd_initial_cluster) > 1 {
						return fmt.Errorf("get proton etcd ETCD_ALL_ENDPOINTS env message error: %v", etcd_initial_cluster)
					}
					if len(etcd_initial_advertise_peer_urls) == 0 || len(etcd_initial_advertise_peer_urls) > 1 {
						return fmt.Errorf("get proton etcd ETCD_INITIAL_ADVERTISE_PEER_URLS env message error: %v", etcd_initial_advertise_peer_urls)
					}
					var name = etcd_cfg.Spec.Hostname
					var initial_cluster = etcd_initial_cluster[0].Value
					var listen_peer_urls = strings.Replace(etcd_initial_advertise_peer_urls[0].Value, "$(MY_POD_NAME)", etcd_cfg.Spec.Hostname, -1)

					Recoverlog.Info("proton etcd name:" + name)
					Recoverlog.Info("proton etcd initial_cluster:" + initial_cluster)
					Recoverlog.Info("proton etcd initial-advertise-peer-urls:" + listen_peer_urls)

					etcd.SnapshotRestoreCommandFunc(filepath.Join(RecoverLogDir, opt.Id+".log"), initial_cluster, "", destpath, "", listen_peer_urls, name, false, []string{etcdSnapshotPath})
					_, err = shellcommand.RunCommand("chown", "-R", "1001", destpath)
					if err != nil {
						return err
					}
					Recoverlog.Info("recover proton-etcd success")
				}
				// etcd.SnapshotRestoreCommandFunc("proton-etcd-0=https://proton-etcd-0.proton-etcd-headless.resource.svc.cluster.local:2380", "etcd-cluster-k8s", destpath, "", "https://proton-etcd-0.proton-etcd-headless.resource.svc.cluster.local:2380", "proton-etcd-0", false, []string{etcdSnapshotPath})
				// 查询pod列表
				pods, err := k.CoreV1().Pods(NamespaceResource).List(context.TODO(), metav1.ListOptions{
					LabelSelector: "app.kubernetes.io/instance=proton-etcd",
				})
				if err != nil {
					Recoverlog.Info("get proton-etcd pod error:", err)
					return err
				}
				for _, pod := range pods.Items {
					err = k.CoreV1().Pods(NamespaceResource).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
					if err != nil {
						Recoverlog.Info("delete proton-etcd pod error:", err)
						return err
					}
				}
				Recoverlog.Info("restart proton-etcd success")
			}
		} else if info.Name == "kubernetes-etcd" {
			etcd_cfg := new(core_v1.Pod)
			for _, p := range info.PathList {
				var backupPath = filepath.Join(workDirectory, p.TargetRelPath)
				// 判断是否存在etcd的k8s配置文件
				back_etcd_yaml_path := filepath.Join(backupPath, "etcd.yaml")
				exits, err := file.PathExists(back_etcd_yaml_path)
				if err != nil {
					return err
				} else {
					bytes, err := os.ReadFile(back_etcd_yaml_path)
					if err != nil {
						return err
					}
					if err := yaml.Unmarshal(bytes, etcd_cfg); err != nil {
						return err
					}
					if etcd_cfg == nil || etcd_cfg.Spec.Containers == nil {
						return fmt.Errorf("Unable to create kubernetes etcd pod struct: %s", back_etcd_yaml_path)
					}
				}
				if !exits {
					return fmt.Errorf("Unable to get kubernetes etcd yaml config: %s", back_etcd_yaml_path)
				}
				if clusterConf == nil {
					return fmt.Errorf("Unable to get cluster configuration file")
				} else if clusterConf.Cs == nil {
					if p.Require {
						return fmt.Errorf("Unable to get cs master configuration")
					}
					Recoverlog.Infof("Unable to get cs master configuration,skip recover kubernetes etcd")
					skipRecover = true
					continue
				} else if len(clusterConf.Cs.Master) == 0 || len(clusterConf.Cs.Master) > 1 {
					if p.Require {
						return fmt.Errorf("kubernetes the number of copies is %d, and kubernetes etcd data cannot be recover", len(clusterConf.Cs.Master))
					}
					Recoverlog.Infof("kubernetes the number of copies is %d, and Proton kubernetes data cannot be recover,skip recover kubernetes etcd", len(clusterConf.Cs.Master))
					skipRecover = true
					continue
				} else if !slices.Contains(clusterConf.Cs.Master, conf.HostName) {
					if p.Require {
						return fmt.Errorf("The current node is not the master")
					}
					Recoverlog.Infoln("The current node is not the master,skip recover kubernetes etcd")
					skipRecover = true
					continue
				}
				files, err := filepath.Glob(backupPath + "/*.db")
				if err != nil {
					return err
				} else if len(files) == 0 {
					return fmt.Errorf("not found kubernetes-etcd backup file : %s", backupPath)
				}
				var etcdSnapshotPath = files[0]
				var destpath = clusterConf.Cs.Etcd_data_dir
				if exits, err := file.PathExists(destpath); err != nil {
					return err
				} else if exits {
					var destBackupPath = clusterConf.Cs.Etcd_data_dir + time.Now().Format("20060102150405") + strconv.Itoa(time.Now().Nanosecond())
					err = file.Copy(destpath, destBackupPath, Recoverlog, nil)
					if err != nil {
						return err
					}
					err = os.RemoveAll(destpath)
					if err != nil {
						return err
					}
				} else {
					err := os.MkdirAll(destpath, os.ModePerm)
					if err != nil {
						Recoverlog.Errorln("创建etcd还原的目录异常：" + err.Error())
						return err
					}
				}

				var etcd_name = arrayutil.Filter(etcd_cfg.Spec.Containers[0].Command, func(s1 string) bool { return !strings.Contains(s1, "--name=") })
				var etcd_initial_cluster = arrayutil.Filter(etcd_cfg.Spec.Containers[0].Command, func(s1 string) bool { return !strings.Contains(s1, "--initial-cluster=") })
				var etcd_initial_advertise_peer_urls = arrayutil.Filter(etcd_cfg.Spec.Containers[0].Command, func(s1 string) bool { return !strings.Contains(s1, "--initial-advertise-peer-urls=") })
				if len(etcd_name) == 0 || len(etcd_name) > 1 {
					return fmt.Errorf("get etcd name command message error: %v", etcd_name)
				}
				if len(etcd_initial_cluster) == 0 || len(etcd_initial_cluster) > 1 {
					return fmt.Errorf("get etcd initial-cluster command message error: %v", etcd_initial_cluster)
				}
				if len(etcd_initial_advertise_peer_urls) == 0 || len(etcd_initial_advertise_peer_urls) > 1 {
					return fmt.Errorf("get etcd initial-advertise-peer-urls command message error: %v", etcd_initial_advertise_peer_urls)
				}
				var name = strings.Replace(etcd_name[0], "--name=", "", -1)
				var initial_cluster = strings.Replace(etcd_initial_cluster[0], "--initial-cluster=", "", -1)
				var initial_advertise_peer_urls = strings.Replace(etcd_initial_advertise_peer_urls[0], "--listen-peer-urls=", "", -1)
				Recoverlog.Info("etcd etcdSnapshotPath:" + etcdSnapshotPath)
				Recoverlog.Info("etcd name:" + name)
				Recoverlog.Info("etcd initial_cluster:" + initial_cluster)
				Recoverlog.Info("etcd initial-advertise-peer-urls:" + initial_advertise_peer_urls)
				etcd.SnapshotRestoreCommandFunc(filepath.Join(RecoverLogDir, opt.Id+".log"), initial_cluster, "", destpath, "", initial_advertise_peer_urls, name, false, []string{etcdSnapshotPath})
				Recoverlog.Info("recover kubernetes-etcd success")
				time.Sleep(10 * time.Second)
				// 所有集群中的节点重启kubelet
				err = restartNodesService(clusterConf.Nodes)
				if err != nil {
					return err
				}
				time.Sleep(time.Second * 120)
				//1 检测k8s恢复正常
				//2 后续需要恢复的数据组件-已正常启动和访问
				for count := 1; count <= RecoverCycleCount; count++ {
					_, k = client.NewK8sClient()
					if k == nil {
						return client.ErrKubernetesClientSetNil
					}
					clusterConf, err = configuration.LoadFromKubernetes(context.Background(), k)
					if err != nil {
						Recoverlog.Info(err.Error())
					}
					if clusterConf == nil {
						Recoverlog.Infof("Unable to get cluster configuration file")
					} else {
						goto KubernetesServiceReady
					}
					if count == RecoverCycleCount {
						return fmt.Errorf("kubernetes service is not ready,checked %d count,timeout", count)
					}
					time.Sleep(time.Second * 10)
				}
			KubernetesServiceReady:
				Recoverlog.Info("The Kubernetes is ready")
			}
		} else {
			for _, p := range info.PathList {
				if !p.Require {
					if exist, err := file.PathExists(filepath.Join(workDirectory, p.TargetRelPath)); err != nil {
						return err
					} else if !exist {
						continue
					}
				}
				err := file.Copy(filepath.Join(workDirectory, p.TargetRelPath), p.OsPath, Recoverlog, p.ExcludeathList)
				if err != nil {
					return err
				}
			}
			var errList []error
			switch info.Name {
			case "proton-cr":
				//还原配置之后，重启proton-cr服务
				if _, err := shellcommand.RunCommand("systemctl", "restart", "proton-cr", "proton-cr-registry", "proton-cr-chartmuseum"); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
				if _, err := shellcommand.RunCommand("systemctl", "enable", "proton-cr", "proton-cr-registry", "proton-cr-chartmuseum"); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
				if _, err := shellcommand.RunCommand("sleep", "5s"); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
				//proton-cr启动之后，上传镜像
				cr := &cr.Cr{
					Logger:        Recoverlog,
					ClusterConf:   clusterConf,
					PrePullImages: false,
				}
				ociPkgPath := "service-package/images"
				if _, err := os.Stat(ociPkgPath); err != nil {
					Recoverlog.Info("The oci images package directory does not exist in the current directory:" + ociPkgPath)
				} else {
					Recoverlog.Printf("oci package: %v", ociPkgPath)
					if err := cr.PushImages(ociPkgPath); err != nil {
						errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
					}
					Recoverlog.Printf("\033[1;37;42m%s\033[0m\n", "Push images success")
				}
			case "proton-slb":
				//还原配置之后，重启proton-slb服务
				if _, err := shellcommand.RunCommand("systemctl", "restart", "haproxy", "keepalived", "proton_slb_manager"); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
				if _, err := shellcommand.RunCommand("systemctl", "enable", "haproxy", "keepalived", "proton_slb_manager", "slb-nginx"); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
			case "firewalld":
				//还原配置之后，配置firewalld服务
				if _, err := shellcommand.RunCommand("systemctl", "restart", "firewalld"); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
				if _, err := shellcommand.RunCommand("systemctl", "enable", "firewalld"); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
			case "network":
				//还原配置之后，配置network服务
				if _, err := shellcommand.RunCommand("/bin/bash", "-c", `if [[ $(systemctl list-units --type service  | grep network.service | wc -l)  -gt 0   ]] && [[ $(systemctl is-active network.service) = active ]]; then systemctl restart network; fi`); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
				if _, err := shellcommand.RunCommand("/bin/bash", "-c", ` if [[ $(systemctl list-units --type service  | grep  NetworkManager.service | wc -l)  -gt 0   ]] && [[ $(systemctl is-active  NetworkManager.service) = active ]]; then systemctl restart  NetworkManager.service; fi`); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
			case "sysctl":
				//还原配置之后，配置sysctl服务
				if _, err := shellcommand.RunCommand("sysctl", "-p"); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
			case "eceph":
				//还原配置之后，配置eceph服务
				if exist, err := file.PathExists(filepath.Join(workDirectory, "eceph/ceph")); err != nil {
					return err
				} else if exist {
					if exist, err := file.PathExists("/etc/ceph/ceph.conf"); err != nil {
						return err
					} else if exist {
						content, err := shellcommand.RunCommand("/bin/bash", "-c", `cat /etc/ceph/ceph.conf 2>/dev/null| grep "host.m" | wc -l`)
						if err != nil {
							return err
						}
						content = strings.Replace(content, "\n", "", -1)
						count, err := strconv.Atoi(content)
						if err != nil {
							return err
						}
						if count > 0 {
							for i := 1; i <= count; i++ {
								content, err := shellcommand.RunCommand("/bin/bash", "-c", `cat /etc/ceph/ceph.conf | grep "host.m" | grep $(hostname) | tr '=' ' ' | awk '{print $1}' | sed "s/host.//g"`)
								if err != nil {
									return err
								}
								content = strings.Replace(content, "\n", "", -1)
								if content != "" {
									var servicename = "ceph-mon@" + content
									if _, err := shellcommand.RunCommand("systemctl", "start", servicename); err != nil {
										errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
									}
									if _, err := shellcommand.RunCommand("systemctl", "enable", servicename); err != nil {
										errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
									}
									if len(errList) > 0 {
										return utilerrors.NewAggregate(errList)
									}
								}
							}
						}
						files, err := filepath.Glob("/etc/ceph/keyring.osd.*")
						if err != nil {
							return err
						}
						if len(files) > 0 {
							for _, filepath := range files {
								var osd_number = strings.Replace(filepath, "/etc/ceph/keyring.osd.", "", -1)
								err := os.MkdirAll("/var/lib/ceph/osd/osd"+osd_number, os.ModePerm)
								if err != nil {
									Recoverlog.Errorln("创建还原配置解压目录异常：", conf.RecoverDirectory, err)
									return err
								}
								if _, err := shellcommand.RunCommand("systemctl", "start", "ceph-osd@"+osd_number); err != nil {
									errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
								}
								if _, err := shellcommand.RunCommand("systemctl", "enable", "ceph-osd@"+osd_number); err != nil {
									errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
								}
								if len(errList) > 0 {
									return utilerrors.NewAggregate(errList)
								}
							}
						}
						if _, err := shellcommand.RunCommand("systemctl", "enable", "ceph-mon.target"); err != nil {
							errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
						}
						if _, err := shellcommand.RunCommand("systemctl", "enable", "ceph-osd.target"); err != nil {
							errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
						}
						if _, err := shellcommand.RunCommand("systemctl", "enable", "ceph-radosgw.target"); err != nil {
							errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
						}
						if _, err := shellcommand.RunCommand("systemctl", "enable", "ceph-mgr.target"); err != nil {
							errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
						}
						if _, err := shellcommand.RunCommand("systemctl", "enable", "ceph.target"); err != nil {
							errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
						}
						if _, err := shellcommand.RunCommand("systemctl", "start", "ceph-radosgw@radosgw.$(hostname)"); err != nil {
							errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
						}
						if _, err := shellcommand.RunCommand("systemctl", "enable", "ceph-radosgw@radosgw.$(hostname)"); err != nil {
							errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
						}
						if _, err := shellcommand.RunCommand("systemctl", "start", "ceph-mgr@$(hostname)"); err != nil {
							errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
						}
						if _, err := shellcommand.RunCommand("systemctl", "enable", "ceph-mgr@$(hostname)"); err != nil {
							errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
						}

					}
				}
			}
			if len(errList) > 0 {
				return utilerrors.NewAggregate(errList)
			}
		}

		if !skipRecover {
			fo.Resource = append(fo.Resource, info.Name)
		}

	}
	return nil
}

// 重启集群中每个节点的kubelet服务
func restartNodesService(hosts []configuration.Node) error {
	var wg sync.WaitGroup
	var errList []error
	for i := range hosts {
		var host string
		if hosts[i].IP4 != "" {
			host = hosts[i].IP4
		} else {
			host = hosts[i].IP6
		}
		ecms := ecms.NewForHost(host)
		executor := exec.NewECMSExecutorForHost(ecms.Exec())
		sshConf := client.RemoteClientConf{
			Host:     host,
			HostName: hosts[i].Name,
		}
		wg.Add(1)
		go func(conf client.RemoteClientConf) {
			defer wg.Done()
			Recoverlog.Info(fmt.Sprintf("restart kubelet on node:%s", conf.Host))
			if err := executor.Command("systemctl", "restart", "kubelet").Run(); err != nil {
				errList = append(errList, fmt.Errorf("restart kubelet on node  %s fail: %w", sshConf.HostName, err))
				return
			}
			if err := executor.Command("systemctl", "enable", "kubelet").Run(); err != nil {
				errList = append(errList, fmt.Errorf("enable kubelet on node  %s fail: %w", sshConf.HostName, err))
				return
			}
			Recoverlog.Info(fmt.Sprintf("restart kubelet on node:%s,success", conf.Host))
		}(sshConf)
	}
	wg.Wait()
	if len(errList) > 0 {
		return utilerrors.NewAggregate(errList)
	}
	return nil
}
