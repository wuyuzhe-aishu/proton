package global

import (
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// 支持的migrate模式
const (
	MigrateECephAndAnyShare = "eceph-and-anyshare"
	MigrateUpdateECephCert  = "update-eceph-certificate"
)

var SupportedECephMigrationMode = []string{
	MigrateECephAndAnyShare,
	MigrateUpdateECephCert,
}

var (
	// reset 时清理服务数据目录
	ClearData                    = false
	ComponentManageDirectConnect = false
	LoggerLevel                  = "info"
	HostsPath                    = "/etc/hosts"
	RegistryDomain               = "registry.aishu.cn"
	ChartmuseumDomain            = "chartmuseum.aishu.cn"
	RpmDomain                    = "rpm.aishu.cn"
	ProtonCsDomain               = "proton-cs.lb.aishu.cn"
	KubeConfigPath               = "/%s/.kube/config"
	ProtonCSName                 = "proton-containerservice"
	ProtonCSVersion              = "1.2.3"

	// TODO: 使用其他方法替换这种使用 module 变量的行为，比如函数中的变量
	NodeAuthSecret []*v1.Secret
	ChartInfoList  []configuration.ChartInfo

	// Proton Package 提供的 service-package 所在的路径
	ServicePackage      = "service-package"
	ServicePackageECeph = "service-package-eceph"

	Helm3DefaultReleaseNamespace = configuration.GetProtonResourceNSFromFile()
)

const NetworkCfgPath = "/etc/sysconfig/network-scripts"
const NetworkSUSECfgPath = "/etc/sysconfig/network"
const CrConfPath = "/etc/proton-cr/proton-cr.yaml"
const ChronyConfPath = "/etc/chrony.conf"
const ClusterDataPath = "/sysvol"
const HelmRepo = "helm_repos"
const K8SAdminConfPath = "/etc/kubernetes/admin.conf"

const KubeletMaxPods = 256
const KubeletEvictNodefsAva = "5%"
const RootHomeDir = "/root"
const NodeStatusUpdateFrequency = 3
const NodeMonitorPeriod = "--node-monitor-period=2s"
const NodeMonitorGracePeriod = "--node-monitor-grace-period=12s"

// MariaDB 的监听的端口号
const MariaDBListenPort = 3330

// CR 相关的默认值
const (
	// RPM 仓库的默认端口号
	DefaultRPMPort = 5003

	// RPM 仓库的高可用默认端口号
	DefaultHighAvailabilityRPMPort = 15003

	// CR Manager 的默认端口号
	DefaultCRManagerPort = 5002

	// CR Manager 的高可用默认端口号
	DefaultHighAvailabilityCRManagerPort = 15002
)

// CS 相关定义
const (
	KubernetesControlPlaneEndpoint = "127.0.0.1:8443"
	KubernetesAPIServerBindPort    = 6443

	// LabelNodeRoleControlPlane specifies that a node hosts control-plane components
	LabelNodeRoleControlPlane = "node-role.kubernetes.io/control-plane"
)

// 是否支持双栈 由cs的配置控制
var EnableDualStack = false

// policy_engine 依赖etcd的节点数 由etcd的配置控制
var EtcdHosts = []string{}

func InitClusterConfigTmpPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "cluster.yaml")
}

// ImageRepository 返回访问指定 CR 的容器镜像仓库所用的地址、用户名、密码。如果用户名非
// 空则代表需要验证。如果CR 无可用的地址则返回空字符串,
func ImageRepository(cr *configuration.Cr) (host, username, password string) {
	switch {
	case cr == nil:
		return
	case cr.Local != nil:
		if cr.Local.Ha_ports.Registry != 0 {
			host = net.JoinHostPort(RegistryDomain, strconv.Itoa(cr.Local.Ha_ports.Registry))
		} else {
			host = RegistryDomain
		}
		return
	case cr.External != nil:
		switch cr.External.ImageRepo {
		case configuration.RepoRegistry:
			return cr.External.Registry.Host, strings.TrimSpace(cr.External.Registry.Username), strings.TrimSpace(cr.External.Registry.Password)
		case configuration.RepoOCI:
			return cr.External.OCI.Registry, strings.TrimSpace(cr.External.OCI.Username), strings.TrimSpace(cr.External.OCI.Password)
		case configuration.RepoDefault:
			return cr.External.Registry.Host, strings.TrimSpace(cr.External.Registry.Username), strings.TrimSpace(cr.External.Registry.Password)
		default:
			return "", "", ""
		}
	default:
		return
	}
}

// ImageRepositoryMulti 返回访问指定 CR 的容器镜像仓库所用的多个节点地址、用户名、密码。
// 当 CR 为 Local 类型时，返回所有节点地址；当为 External 类型时，返回单地址。
func ImageRepositoryMulti(cr *configuration.Cr) (hosts []string, username, password string) {
	switch {
	case cr == nil:
		return
	case cr.Local != nil:
		port := cr.Local.Ports.Registry
		for _, h := range cr.Local.Hosts {
			host := net.JoinHostPort(h, strconv.Itoa(port))
			hosts = append(hosts, host)
		}
		return
	case cr.External != nil:
		switch cr.External.ImageRepo {
		case configuration.RepoRegistry:
			return []string{cr.External.Registry.Host}, strings.TrimSpace(cr.External.Registry.Username), strings.TrimSpace(cr.External.Registry.Password)
		case configuration.RepoOCI:
			return []string{cr.External.OCI.Registry}, strings.TrimSpace(cr.External.OCI.Username), strings.TrimSpace(cr.External.OCI.Password)
		case configuration.RepoDefault:
			return []string{cr.External.Registry.Host}, strings.TrimSpace(cr.External.Registry.Username), strings.TrimSpace(cr.External.Registry.Password)
		default:
			return nil, "", ""
		}
	default:
		return
	}
}

// Chartmuseum 返回访问指定 CR 的 chartmusem 所用的地址、用户名、密码。如果用户
// 名非空则代表需要验证。如果CR 无可用的地址则返回空字符串
func Chartmuseum(cr *configuration.Cr) (host, username, password string) {
	switch {
	case cr.Local != nil:
		p := cr.Local.Ports.Chartmuseum
		if cr.Local.Ha_ports.Chartmuseum != 0 {
			p = cr.Local.Ha_ports.Chartmuseum
		}
		h := ChartmuseumDomain
		if p != 0 {
			h = net.JoinHostPort(h, strconv.Itoa(p))
		}
		hostURL := url.URL{
			Scheme: "http",
			Host:   h,
		}
		host = hostURL.String()
		return
	case cr.External != nil:
		switch cr.External.ChartRepo {
		case configuration.RepoChartmuseum:
			return cr.External.Chartmuseum.Host, cr.External.Chartmuseum.Username, cr.External.Chartmuseum.Password
		case configuration.RepoOCI:
			return "", "", ""
		case configuration.RepoDefault:
			return cr.External.Chartmuseum.Host, cr.External.Chartmuseum.Username, cr.External.Chartmuseum.Password
		default:
			return "", "", ""
		}

	default:
	}
	return
}

// ChartmuseumMulti 返回访问指定 CR 的 chartmuseum 所用的多个节点地址、用户名、密码。
// 当 CR 为 Local 类型时，返回所有节点地址；当为 External 类型时，返回单地址。
func ChartmuseumMulti(cr *configuration.Cr) (hosts []string, username, password string) {
	switch {
	case cr == nil:
		return
	case cr.Local != nil:
		port := cr.Local.Ports.Chartmuseum
		for _, h := range cr.Local.Hosts {
			host := net.JoinHostPort(h, strconv.Itoa(port))
			hostURL := url.URL{
				Scheme: "http",
				Host:   host,
			}
			hosts = append(hosts, hostURL.String())
		}
		return
	case cr.External != nil:
		switch cr.External.ChartRepo {
		case configuration.RepoChartmuseum:
			return []string{cr.External.Chartmuseum.Host}, cr.External.Chartmuseum.Username, cr.External.Chartmuseum.Password
		case configuration.RepoOCI:
			return nil, "", ""
		case configuration.RepoDefault:
			return []string{cr.External.Chartmuseum.Host}, cr.External.Chartmuseum.Username, cr.External.Chartmuseum.Password
		default:
			return nil, "", ""
		}

	default:
	}
	return
}
