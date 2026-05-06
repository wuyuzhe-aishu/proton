package configuration

import (
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
)

func GetConf(valueTag bool) string {
	b, _ := os.ReadFile("/etc/proton-cli/conf/cluster.yaml")
	return string(b)
}

// AccessAddress 描述集群的对外访问地址
type AccessAddress struct {
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Scheme string `json:"scheme,omitempty"`
	Path   string `json:"path,omitempty"`
}

// StorageConfig 描述集群存储配置
type StorageConfig struct {
	StorageClassName string `json:"storageClassName,omitempty"`
}

// EnvConfig 描述运行环境配置
type EnvConfig struct {
	Language string `json:"language,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

// cluster conf
type ClusterConfig struct {
	ApiVersion           string          `json:"apiVersion"`
	Deploy               *Deploy         `json:"deploy,omitempty"`
	Nodes                []Node          `json:"nodes"`
	Chrony               *Chrony         `json:"chrony,omitempty"`
	Firewall             Firewall        `json:"firewall,omitzero"`
	Cs                   *Cs             `json:"cs,omitempty"`
	Cr                   *Cr             `json:"cr,omitempty"`
	Proton_monitor       *ProtonMonitor  `json:"proton_monitor,omitempty"`
	Proton_mariadb       *ProtonMariaDB  `json:"proton_mariadb,omitempty"`
	Proton_mongodb       *ProtonDB       `json:"proton_mongodb,omitempty"`
	Proton_redis         *ProtonDB       `json:"proton_redis,omitempty"`
	Proton_mq_nsq        *ProtonDataConf `json:"proton_mq_nsq,omitempty"`
	Proton_policy_engine *ProtonDataConf `json:"proton_policy_engine,omitempty"`
	Proton_etcd          *ProtonDataConf `json:"proton_etcd,omitempty"`
	OpenSearch           *OpenSearch     `json:"opensearch,omitempty"`

	// AnyShare 的 configuration management service
	CMS             *CMS                 `json:"cms,omitempty"`
	ComponentManage *ComponentManagement `json:"component_management,omitempty"`
	// nvidia-device-plugin for GPU computation on Kubernetes
	NvidiaDevicePlugin *NvidiaDevicePlugin `json:"nvidia_device_plugin,omitempty"`

	// AnyRobot 使用的 Kafka
	Kafka *Kafka `json:"kafka,omitempty"`
	// AnyRobot 使用的 ZooKeeper
	ZooKeeper *ZooKeeper `json:"zookeeper,omitempty"`

	// Proton 坯观测性朝务使用的 Prometheus
	Prometheus *Prometheus `json:"prometheus,omitempty"`
	// Proton 坯观测性朝务使用的 Grafana
	Grafana *Grafana `json:"grafana,omitempty"`

	// Nebula Graph 的部署酝置，如果为 nil 则丝会安装 Nebula Graph
	Nebula *Nebula `json:"nebula,omitempty"`

	// 集群对外访问地址（用于 helm values 生成）
	AccessAddress *AccessAddress `json:"access_address,omitempty"`
	// 存储配置（用于 helm values 生成）
	Storage *StorageConfig `json:"storage,omitempty"`
	// 运行环境配置（language/timezone）
	Env *EnvConfig `json:"env,omitempty"`

	//  基础组件连接信息都在此保存，替代cms的保存
	ResourceConnectInfo *ResourceConnectInfo `json:"resource_connect_info,omitempty" mapstructure:"resourceConnectInfo"`

	// Proton 包管睆朝务
	PackageStore *PackageStore `json:"package-store,omitempty"`
}

type Deploy struct {
	Namespace      string `json:"namespace,omitempty"`
	ServiceAccount string `json:"serviceaccount,omitempty"`
}

type Node struct {
	Name        string `json:"name"`
	IP4         string `json:"ip4,omitempty"`
	IP6         string `json:"ip6,omitempty"`
	Internal_ip string `json:"internal_ip,omitempty"`
}

// IP 返回节点的 IPv4 或 IPv6 地址，优先返回 IPv4 地址。
func (n *Node) IP() string {
	if n.IP4 != "" {
		return n.IP4
	}
	return n.IP6
}

// GetNodeNameByIP 返回指定节点列表包坫指定 IP 地址的节点坝称，如果节点丝存在则返回空字符串
func GetNodeNameByIP(ip string, nodes []Node) string {
	for _, node := range nodes {
		if node.IP4 == ip || node.IP6 == ip {
			return node.Name
		}
	}
	return ""
}

// GetIPByNodeName 返回指定节点列表中指定节点坝称的节点 IP 地址，如果节点丝存在则返回空字符串
func GetIPByNodeName(name string, nodes []Node) string {
	for _, node := range nodes {
		if node.Name == name {
			return node.IP()
		}
	}
	return ""
}

type ByGetName []Node

func (a ByGetName) Len() int           { return len(a) }
func (a ByGetName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByGetName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type Cs struct {
	// Kubernetes 的杝供者
	Provisioner KubernetesProvisioner `json:"provisioner,omitempty"`
	// Enable Kubernetes DualStack
	EnableDualStack bool `json:"enableDualStack"`
	// IPFamilies is a list of IP families (e.g. IPv4, IPv6) used by kubernetes.
	IPFamilies []v1.IPFamily `json:"ipFamilies,omitempty"`
	// 容器运行时，仅当 Provisioner 为 local 时需要配置
	ContainerRuntime ContainerRuntimeSource `json:"container_runtime,omitzero"`

	Master            []string     `json:"master"`
	Host_network      *HostNetWork `json:"host_network"`
	Ha_port           int          `json:"ha_port"`
	Etcd_data_dir     string       `json:"etcd_data_dir"`
	Cs_controller_dir string       `json:"cs_controller_dir"`
	// Proton CS 坯用的杒件列表, 因为需覝区分 `nil` 和 `[]` 所以丝能使用
	// omitempty
	Addons       []CSAddonName   `json:"addons"`
	AddonsConfig *CSAddonsConfig `json:"addonsConfig,omitempty"`
}

type CSAddonsConfig struct {
	IngressNginx *CSAddonIngressNginxConfig `json:"ingress-nginx,omitempty"`
}

type CSAddonIngressNginxConfig struct {
	HTTPPort  int `json:"httpPort,omitempty"`
	HTTPSPort int `json:"httpsPort,omitempty"`
}

// Kubernetes 的容器运行时，有且只有一个运行时
type ContainerRuntimeSource struct {
	Containerd *ContainerdContainerRuntimeSource `json:"containerd,omitzero"`
}

// 容器运行时 containerd
type ContainerdContainerRuntimeSource struct {
	// Root is the path to a directory where containerd will store persistent
	// data
	Root string `json:"root,omitzero"`
	// SandboxImage is the image used by sandbox container.
	SandboxImage string `json:"sandbox_image,omitzero"`
	// Registry-specific configuration
	Registries []RegistryHostConfig `json:"registries,omitzero"`
}

type RegistryHostConfig struct {
	// Server specifies the default server. When `host` is
	// also specified, those hosts are tried first.
	Server string `toml:"server" json:"server,omitzero"`

	// HostConfigs store the per-host configuration
	HostConfigs map[string]RegistryHostFileConfig `toml:"host" json:"host_configs,omitzero"`
}

// copy from github.com/containerd/containerd@v1.7.25/remotes/docker/config.hostFileConfig
type RegistryHostFileConfig struct {
	// Capabilities determine what operations a host is
	// capable of performing. Allowed values
	//  - pull
	//  - resolve
	//  - push
	Capabilities []string `toml:"capabilities" json:"capabilities,omitzero"`

	// CACert are the public key certificates for TLS
	// Accepted types
	// - string - Single file with certificate(s)
	// - []string - Multiple files with certificates
	CACert any `toml:"ca" json:"ca_cert,omitzero"`

	// Client keypair(s) for TLS with client authentication
	// Accepted types
	// - string - Single file with public and private keys
	// - []string - Multiple files with public and private keys
	// - [][2]string - Multiple keypairs with public and private keys in separate files
	Client any `toml:"client" json:"client,omitzero"`

	// SkipVerify skips verification of the server's certificate chain
	// and host name. This should only be used for testing or in
	// combination with other methods of verifying connections.
	SkipVerify *bool `toml:"skip_verify" json:"skip_verify,omitzero"`

	// Header are additional header files to send to the server
	Header map[string]any `toml:"header" json:"header,omitzero"`

	// OverridePath indicates the API root endpoint is defined in the URL
	// path rather than by the API specification.
	// This may be used with non-compliant OCI registries to override the
	// API root endpoint.
	OverridePath bool `toml:"override_path" json:"override_path,omitzero"`

	// TODO: Credentials: helper? name? username? alternate domain? token?
}

type Chrony struct {
	// 时间服务器相关的配置
	Mode   string   `json:"mode,omitempty"`
	Server []string `json:"server,omitempty"`
}

// FirewallMode 防火墙模式，代表 proton-cli 使用什么软件作为防火墙配置规则
type FirewallMode string

const (
	// 防火墙由用户自行管理，proton-cli 不修改防火墙配置
	FirewallUserManaged FirewallMode = "usermanaged"
	// 防火墙由 firewalld 实现 https://www.firewalld.org
	FirewallFirewalld FirewallMode = "firewalld"
)

// Firewall 代表集群的防火墙配置
type Firewall struct {
	Mode FirewallMode `json:"mode,omitzero"`
}

const (
	IPVersionIPV4 string = "ipv4"

	IPVersionIPV6 string = "ipv6"
)

const (
	// usermanaged 用户自行管理Chrony配置，ProtonCLI不会对Chrony配置进行任何变更，原有ProtonCLI配置文件脱离管控
	ChronyModeUserManaged string = "usermanaged"
	// localmaster 删除所有现有的时间服务器，然后随机选择一个master节点作为集群内唯一的时间服务器
	ChronyModeLocalMaster string = "localmaster"
	// externalntp 删除所有现有的时间服务器，然后将用户指定的时间服务器作为集群内唯一的时间服务器
	ChronyModeExternalNTP string = "externalntp"
)

type KubernetesProvisioner string

// Valid kubernetes provisioner
const (
	// Proton CLI 创建的 kubernetes
	KubernetesProvisionerLocal KubernetesProvisioner = "local"
	// 外部的 kubernetes，Proton CLI 仅作使用
	KubernetesProvisionerExternal KubernetesProvisioner = "external"
)

type HostNetWork struct {
	Bip              string `json:"bip"`
	Pod_network_cidr string `json:"pod_network_cidr"`
	Service_cidr     string `json:"service_cidr"`
	Ipv4_interface   string `json:"ipv4_interface,omitempty"`
	Ipv6_interface   string `json:"ipv6_interface,omitempty"`
}

// Cr contains elements describing Cr configuration.
type Cr struct {
	// Local provides configuration knobs for configuring the local cr instance
	// Local and External are mutually exclusive
	Local *LocalCR `json:"local,omitempty"`

	// External describes how to connect to an external cr cluster
	// Local and External are mutually exclusive
	External *ExternalCR `json:"external,omitempty"`
}

// LocalCR describes that proton-cli should run an cr cluster locally
type LocalCR struct {
	Hosts    []string `json:"hosts"`
	Ports    Ports    `json:"ports"`
	Ha_ports Ports    `json:"ha_ports"`
	Storage  string   `json:"storage"`
}

// ExternalCR describes an external cr cluster.
type ExternalCR struct {
	// Registry holds configuration for registry.
	Registry *Registry `json:"registry,omitempty"`
	// Chartmuseum holds configuration for chartmuseum.
	Chartmuseum *Chartmuseum `json:"chartmuseum,omitempty"`
	// OCI holds configuration for oci, it can be used for image/chart repository
	OCI *OCI `json:"oci,omitempty"`

	ChartRepo string `json:"chart_repository"`
	ImageRepo string `json:"image_repository"`
}

func (cr *ExternalCR) ImageRepository() string {
	switch cr.ImageRepo {
	case RepoDefault:
		return cr.Registry.Host
	case RepoRegistry:
		return cr.Registry.Host
	case RepoOCI:
		return cr.OCI.Registry
	}
	// Unreachable
	return ""
}

func (cr *ExternalCR) ValidateExternalCR() error {
	switch cr.ImageRepo {
	case RepoOCI:
		if cr.OCI == nil {
			return fmt.Errorf("image repository is using oci, please provide valid oci info")
		}
	case RepoRegistry, RepoDefault:
		if cr.Registry == nil {
			return fmt.Errorf("image repository is using registry, please provide valid registry info")
		}
	default:
		return fmt.Errorf("image repository is using an unsupported repository type")
	}
	switch cr.ChartRepo {
	case RepoOCI:
		if cr.OCI == nil {
			return fmt.Errorf("chart repository is using oci, please provide valid oci info")
		}
	case RepoChartmuseum, RepoDefault:
		if cr.Chartmuseum == nil {
			return fmt.Errorf("chart repository is using chartmuseum, please provide valid chartmuseum info")
		}
	default:
		return fmt.Errorf("chart repository is using an unsupported repository type")
	}
	return nil
}

func (cr *Cr) UseChartmuseum() bool {
	if cr.Local != nil {
		return true
	}
	if cr.External != nil {
		return cr.External.ChartRepo == RepoChartmuseum || cr.External.ChartRepo == RepoDefault
	}

	return false
}

const (
	RepoRegistry    = "registry"
	RepoChartmuseum = "chartmuseum"
	RepoOCI         = "oci"
	RepoDefault     = ""
)

// Registry describes an external container registry.
type Registry struct {
	Host     string `json:"host,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type OCI struct {
	Registry  string `json:"registry,omitempty"`
	PlainHTTP bool   `json:"plain_http,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
}

// Chartmuseum describes an external chartmuseum.
type Chartmuseum struct {
	Host     string `json:"host,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Ports struct {
	Chartmuseum int `json:"chartmuseum"`
	Registry    int `json:"registry"`
	Rpm         int `json:"rpm"`
	Cr_manager  int `json:"cr_manager"`
}

// HelmComponent 包坫 helm 组件的通用酝置
type HelmComponent struct {
	// ExtraValues 是传递给 Proton OpenSearch 的额外的 Helm Values
	ExtraValues map[string]any `json:"extraValues,omitempty"`
}

type ProtonMariaDB struct {
	ReplicaCount int                   `json:"replica_count,omitempty"`
	Hosts        []string              `json:"hosts"`
	Config       *ProtonMariaDBConfigs `json:"config"`
	Admin_user   string                `json:"admin_user"`
	Admin_passwd string                `json:"admin_passwd"`
	Data_path    string                `json:"data_path"`

	// 使用的 storage class 的坝称，空字符串代表丝使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`
}

const (
	// redis 的 Chart 坝称
	ChartNameRedis = "proton-redis"
	// redis 的 Helm release 坝称
	ReleaseNameRedis = "proton-redis"
)

type ProtonDB struct {
	ReplicaCount int      `json:"replica_count,omitempty"`
	Hosts        []string `json:"hosts"`
	Admin_user   string   `json:"admin_user"`
	Admin_passwd string   `json:"admin_passwd"`
	Data_path    string   `json:"data_path"`

	// 使用的 storage class 的坝称，空字符串代表丝使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`

	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

const (
	// mongodb operator 的 chart release 坝称
	ChartNameMongodbOperator   = "mongodb-operator"
	ReleaseNameMongodbOperator = "mongodb-operator"
)

const (
	// policyEngine 的 Chart 坝称
	ChartNamePolicyEngine = "proton-policy-engine"
	// policyEngine 的 Helm release 坝称
	ReleaseNamePolicyEngine = "proton-policy-engine"
)

type ProtonDataConf struct {
	ReplicaCount int      `json:"replica_count,omitempty"`
	Hosts        []string `json:"hosts"`
	Data_path    string   `json:"data_path"`

	// 使用的 storage class 的坝称，空字符串代表丝使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`

	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type ProtonMariaDBConfigs struct {
	LowerCaseTableNames *int `json:"lower_case_table_names,omitempty"`

	Thread_handling          string `json:"thread_handling,omitempty"`
	Innodb_buffer_pool_size  string `json:"innodb_buffer_pool_size"`
	Resource_requests_memory string `json:"resource_requests_memory"`
	Resource_limits_memory   string `json:"resource_limits_memory"`
}

const ChartNameEtcd = "proton-etcd"

const (
	// Opensearch 的 Chart 坝称
	ChartNameOpensearch = "proton-opensearch"
	// Opensearch 的 Helm release 坝称
	// release坝称由“opensearch”+mode拼接
)

type OpenSearch struct {
	HelmComponent `json:",inline"`

	ReplicaCount int               `json:"replica_count,omitempty"`
	Hosts        []string          `json:"hosts"`
	Data_path    string            `json:"data_path"`
	Mode         OpenSearchMode    `json:"mode"`
	Config       OpenSearchConfigs `json:"config"`

	// OpenSearch 的原生坂数
	// 	Detail: "https://opensearch.org/docs/latest/security/configuration/yaml/#opensearchyml"
	Settings map[string]any `json:"settings,omitempty"`
	// OpenSearch 需求的资溝
	Resources         *v1.ResourceRequirements `json:"resources,omitempty"`
	ExporterResources *v1.ResourceRequirements `json:"exporter_resources,omitempty"`

	// 使用的 storage class 的坝称，空字符串代表丝使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`
}

type OpenSearchMode string

type Version string

// Valid OpenSearch mode
const (
	OpenSearchModeHot    OpenSearchMode = "hot"
	OpenSearchModeMaster OpenSearchMode = "master"
	OpenSearchModeWarm   OpenSearchMode = "warm"

	// 支挝的opensearch版本
	Version564 Version = "5.6.4"
	Version710 Version = "7.10.0"
)

type OpenSearchConfigs struct {
	JvmOptions              string `json:"jvmOptions"`
	HanlpRemoteextDict      string `json:"hanlpRemoteextDict"`
	HanlpRemoteextStopwords string `json:"hanlpRemoteextStopwords"`
}

const (
	// orientdb 的 Chart 坝称
	ChartNameOrientDB = "kg-orientdb"

	// orientdb 的 Helm release 坝称，与 chart 坝称一致
	ReleaseNameOrientDB = "kg-orientdb"
)

const (
	// CMS 的 Chart 名称
	ChartNameCMS = "configuration-management-service"

	// CMS 的 Helm release 名称，与 chart 名称一致
	ReleaseNameCMS = "configuration-management-service"
)

// CMS 代表 AnyShare 的 configuration management service
type CMS struct {
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// InstallerService 代表 AnyShare 的 installer service
type InstallerService struct {
}

type ComponentManagement struct {
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

const (
	// nvidia-device-plugin 的 Chart 名称
	ChartNameNvDevPlugin = "nvidia-device-plugin"

	// nvidia-device-plugin 的 Helm release 名称，与 chart 名称一致
	ReleaseNameNvDevPlugin = "nvidia-device-plugin"
)

// NvidiaDevicePlugin 代表 nvidia-device-plugin，在K8S集群中使用英伟达GPU时需要使用
type NvidiaDevicePlugin struct {
}

// slb
type HaProxyConf struct {
	Conf Conf `json:"conf"`
}

type Conf struct {
	Global   `json:"global"`
	Defaults `json:"defaults"`
	Frontend `json:"frontend"`
	Backend  `json:"backend"`
}
type Global struct {
	Log     string `json:"log"`
	Maxconn string `json:"maxconn"`
}
type Defaults struct {
	Maxconn string `json:"maxconn"`
	Mode    string `json:"mode"`
	Timeout string `json:"timeout"`
	Option  string `json:"option"`
}

type Frontend struct {
	CsFrontend            `json:"cs"`
	CrChartmuseumFrontend `json:"cr-chartmuseum"`
	CrRpmFrontend         `json:"cr-rpm"`
	CrRegistryFrontend    `json:"cr-registry"`
}
type CsFrontend struct {
	Bind           string `json:"bind"`
	Mode           string `json:"mode"`
	DefaultBackend string `json:"default_backend"`
}
type CrChartmuseumFrontend struct {
	Bind           string `json:"bind"`
	Mode           string `json:"mode"`
	DefaultBackend string `json:"default_backend"`
}
type CrRpmFrontend struct {
	Bind           string `json:"bind"`
	Mode           string `json:"mode"`
	DefaultBackend string `json:"default_backend"`
}
type CrRegistryFrontend struct {
	Bind           string `json:"bind"`
	Mode           string `json:"mode"`
	DefaultBackend string `json:"default_backend"`
}

type Backend struct {
	CsBackend            `json:"cs"`
	CrRpmBackend         `json:"cr-rpm"`
	CrRegistryBackend    `json:"cr-registry"`
	CrChartmuseumBackend `json:"cr-chartmuseum"`
}

type CsBackend struct {
	Option        []string `json:"option"`
	DefaultServer string   `json:"default-server"`
	HTTPCheck     string   `json:"http-check"`
	Balance       string   `json:"balance"`
	Server        []string `json:"server"`
	Mode          string   `json:"mode"`
}
type CrRpmBackend struct {
	Option        []string `json:"option"`
	DefaultServer string   `json:"default-server"`
	HTTPCheck     string   `json:"http-check"`
	Balance       string   `json:"balance"`
	Server        []string `json:"server"`
	Mode          string   `json:"mode"`
}
type CrRegistryBackend struct {
	Option        []string `json:"option"`
	DefaultServer string   `json:"default-server"`
	HTTPCheck     string   `json:"http-check"`
	Balance       string   `json:"balance"`
	Server        []string `json:"server"`
	Mode          string   `json:"mode"`
}
type CrChartmuseumBackend struct {
	Option        []string `json:"option"`
	DefaultServer string   `json:"default-server"`
	HTTPCheck     string   `json:"http-check"`
	Balance       string   `json:"balance"`
	Server        []string `json:"server"`
	Mode          string   `json:"mode"`
}

const InternalIPCfg = `DEVICE=%s
PREFIX=%s
BOOTPROTO=static
IPADDR=%s
ONBOOT=yes
`

const InternalIPV6Cfg = `DEVICE=%s
PREFIX=%s
BOOTPROTO=static
IPV6ADDR=%s
ONBOOT=yes
`

const HaDefaultConf = `{
    "conf": {
        "global": {
            "log": "/dev/log local0",
            "maxconn": "50000"
        },
        "frontend": {
            "cs": {
                "bind": ":::8443 v4v6",
                "mode": "tcp",
                "default_backend": "cs"
            },
            "cr-chartmuseum": {
                "bind": ":::15001 v4v6",
                "mode": "tcp",
                "default_backend": "cr-chartmuseum"
            },
            "cr-registry": {
                "bind": ":::15000 v4v6",
                "mode": "tcp",
                "default_backend": "cr-registry"
            },
            "cr-rpm": {
                "bind": ":::15003 v4v6",
                "mode": "tcp",
                "default_backend": "cr-rpm"
            }
        },
        "defaults": {
            "maxconn": "5000",
            "mode": "   tcp",
            "timeout": "tunnel       86400s",
            "option": " dontlognull"
        },
        "backend": {
            "cs": {
                "option": [
                    " httpchk GET /readyz HTTP/1.0",
                    " log-health-checks"
                ],
                "default-server": "verify none check-ssl inter 3s downinter 5s rise 2 fall 2 slowstart 60s maxconn 5000 maxqueue 5000 weight 100",
                "http-check": "expect status 200",
                "balance": "first",
                "server": [
                ],
                "mode": "tcp"
            },
            "cr-chartmuseum": {
                "option": [
                    " httpchk GET /health HTTP/1.0",
                    " log-health-checks"
                ],
                "http-check": "expect status 200",
                "balance": "first",
                "server": [
                ],
                "mode": "tcp"
            },
            "cr-registry": {
                "option": [
                    " httpchk GET / HTTP/1.0",
                    " log-health-checks"
                ],
                "http-check": "expect status 200",
                "balance": "first",
                "server": [

                ],
                "mode": "tcp"
            },
            "cr-rpm": {
                "option": [
                    " httpchk GET / HTTP/1.0",
                    " log-health-checks"
                ],
                "http-check": "expect status 200",
                "balance": "first",
                "server": [
                ],
                "mode": "tcp"
            }
        }
    }
}`

//cr conf

type CrConf struct {
	ConfigFile ConfigFile `json:"configfile"`
	Port       Port       `json:"port"`
	Storage    string     `json:"storage,omitempty"`
}

type ConfigFile struct {
	Chartmuseum string `json:"chartmuseum,omitempty"`
	Registry    string `json:"registry,omitempty"`
	Rpm         string `json:"rpm,omitempty"`
}

type Port struct {
	Chartmuseum int `json:"chartmuseum,omitempty"`
	Crmanager   int `json:"crmanager,omitempty"`
	Registry    int `json:"registry,omitempty"`
	Rpm         int `json:"rpm,omitempty"`
}

// chrony

var ChronyDefaultConf = `
# These servers were defined in the installation:
# Use public servers from the pool.ntp.org project.
# Please consider joining the pool (http://www.pool.ntp.org/join.html).

# Ignore stratum in source selection.
stratumweight 0

# Record the rate at which the system clock gains/losses time.
driftfile /var/lib/chrony/drift

# Enable kernel RTC synchronization.
rtcsync

# In first three updates step the system clock instead of slew
# if the adjustment is larger than 10 seconds.
makestep 10 3

# Allow NTP client access from local network.
allow 0/0

# Listen for commands only on localhost.
bindcmdaddress 127.0.0.1
bindcmdaddress ::1

# Serve time even if not synchronized to any NTP server.
local stratum 10

keyfile /etc/chrony.keys

# Specify the key used as password for chronyc.
commandkey 1

# Generate command key if missing.
generatecommandkey

# Disable logging of client accesses.
noclientlog

# Send a message to syslog if a clock adjustment is larger than 0.5 seconds.
logchange 0.5

logdir /var/log/chrony
#log measurements statistics tracking
`

var ChronyServerConf = `
#master node
# These servers were defined in the installation:
# Use public servers from the pool.ntp.org project.
# Please consider joining the pool (http://www.pool.ntp.org/join.html).

# Ignore stratum in source selection.
stratumweight 0

# Record the rate at which the system clock gains/losses time.
driftfile /var/lib/chrony/drift

# Enable kernel RTC synchronization.
rtcsync

# In first three updates step the system clock instead of slew
# if the adjustment is larger than 10 seconds.
makestep 10 3

# Allow NTP client access from local network.
allow 0/0

# Listen for commands only on localhost.
bindcmdaddress 127.0.0.1
bindcmdaddress ::1

# Serve time even if not synchronized to any NTP server.
local stratum 10 orphan

keyfile /etc/chrony.keys

# Specify the key used as password for chronyc.
commandkey 1

# Generate command key if missing.
generatecommandkey

# Disable logging of client accesses.
noclientlog

# Send a message to syslog if a clock adjustment is larger than 0.5 seconds.
logchange 0.5

logdir /var/log/chrony
#log measurements statistics tracking
`

var ChronyClientConf = `#slave node
# These servers were defined in the installation:
# Use public servers from the pool.ntp.org project.
# Please consider joining the pool (http://www.pool.ntp.org/join.html).
%s

# Ignore stratum in source selection.
stratumweight 0

# Record the rate at which the system clock gains/losses time.
driftfile /var/lib/chrony/drift

# Enable kernel RTC synchronization.
rtcsync

# In first three updates step the system clock instead of slew
# if the adjustment is larger than 10 seconds.
makestep 10 3

# Allow NTP client access from local network.
allow

# Listen for commands only on localhost.
bindcmdaddress 127.0.0.1
bindcmdaddress ::1

# Serve time even if not synchronized to any NTP server.
local stratum 10

keyfile /etc/chrony.keys

# Specify the key used as password for chronyc.
commandkey 1

# Generate command key if missing.
generatecommandkey

# Disable logging of client accesses.
noclientlog

# Send a message to syslog if a clock adjustment is larger than 0.5 seconds.
logchange 0.5

logdir /var/log/chrony
#log measurements statistics tracking
`

type ChartInfo struct {
	ApiVersion  string `json:"apiVersion"`
	AppVersion  string `json:"appVersion"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Version     string `json:"version"`
}

// kubeconfig configmap struct

type ClusterConfiguration struct {
	//metav1.TypeMeta
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
	// Etcd holds configuration for etcd.
	Etcd Etcd `json:"etcd"`

	// Networking holds configuration for the networking topology of the cluster.
	Networking Networking `json:"networking"`
	// KubernetesVersion is the target version of the control plane.
	KubernetesVersion string `json:"kubernetesVersion"`

	// ControlPlaneEndpoint sets a stable IP address or DNS name for the control plane; it
	// can be a valid IP address or a RFC-1123 DNS subdomain, both with optional TCP port.
	// In case the ControlPlaneEndpoint is not specified, the AdvertiseAddress + BindPort
	// are used; in case the ControlPlaneEndpoint is specified but without a TCP port,
	// the BindPort is used.
	// Possible usages are:
	// e.g. In a cluster with more than one control plane instances, this field should be
	// assigned the address of the external load balancer in front of the
	// control plane instances.
	// e.g.  in environments with enforced node recycling, the ControlPlaneEndpoint
	// could be used for assigning a stable DNS to the control plane.
	ControlPlaneEndpoint string `json:"controlPlaneEndpoint"`

	// APIServer contains extra settings for the API server control plane component
	APIServer APIServer `json:"apiServer"`

	// ControllerManager contains extra settings for the controller manager control plane component
	ControllerManager ControlPlaneComponent `json:"controllerManager"`

	// Scheduler contains extra settings for the scheduler control plane component
	Scheduler ControlPlaneComponent `json:"scheduler"`

	// DNS defines the options for the DNS add-on installed in the cluster.
	DNS DNS `json:"dns"`

	// CertificatesDir specifies where to store or look for all required certificates.
	CertificatesDir string `json:"certificatesDir"`

	// ImageRepository sets the container registry to pull images from.
	// If empty, `k8s.gcr.io` will be used by default; in case of kubernetes version is a CI build (kubernetes version starts with `ci/`)
	// `gcr.io/k8s-staging-ci-images` will be used as a default for control plane components and for kube-proxy, while `k8s.gcr.io`
	// will be used for all the other images.
	ImageRepository string `json:"imageRepository"`

	// CIImageRepository is the container registry for core images generated by CI.
	// Useful for running kubeadm with images from CI builds.
	// +k8s:conversion-gen=false
	//CIImageRepository string

	// FeatureGates enabled by the user.
	//FeatureGates map[string]bool

	// The cluster name
	ClusterName string `json:"clusterName"`
}

type Etcd struct {

	// Local provides configuration knobs for configuring the local etcd instance
	// Local and External are mutually exclusive
	Local *LocalEtcd `json:"local"`

	// External describes how to connect to an external etcd cluster
	// Local and External are mutually exclusive
	//External *ExternalEtcd
}

type LocalEtcd struct {
	// ImageMeta allows to customize the container used for etcd
	//ImageMeta `json:",inline"`

	// DataDir is the directory etcd will place its data.
	// Defaults to "/var/lib/etcd".
	DataDir string `json:"dataDir"`

	// ExtraArgs are extra arguments provided to the etcd binary
	// when run inside a static pod.
	// A key in this map is the flag name as it appears on the
	// command line except without leading dash(es).
	ExtraArgs map[string]string `json:"extraArgs"`

	// ServerCertSANs sets extra Subject Alternative Names for the etcd server signing cert.
	//ServerCertSANs []string
	// PeerCertSANs sets extra Subject Alternative Names for the etcd peer signing cert.
	//PeerCertSANs []string
}

type ExternalEtcd struct {

	// Endpoints of etcd members. Useful for using external etcd.
	// If not provided, kubeadm will run etcd in a static pod.
	Endpoints []string
	// CAFile is an SSL Certificate Authority file used to secure etcd communication.
	CAFile string
	// CertFile is an SSL certification file used to secure etcd communication.
	CertFile string
	// KeyFile is an SSL key file used to secure etcd communication.
	KeyFile string
}

type Networking struct {
	// ServiceSubnet is the subnet used by k8s services. Defaults to "10.96.0.0/12".
	ServiceSubnet string `json:"serviceSubnet"`
	// PodSubnet is the subnet used by pods.
	PodSubnet string `json:"podSubnet"`
	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string `json:"dnsDomain"`
}

type APIServer struct {
	ControlPlaneComponentExtra `json:",inline"`

	// CertSANs sets extra Subject Alternative Names for the API Server signing cert.
	//CertSANs []string

	// TimeoutForControlPlane controls the timeout that we use for API server to appear
	TimeoutForControlPlane string `json:"timeoutForControlPlane"`
}

type ControlPlaneComponent struct {
	// ExtraArgs is an extra set of flags to pass to the control plane component.
	// A key in this map is the flag name as it appears on the
	// command line except without leading dash(es).
	// TODO: This is temporary and ideally we would like to switch all components to
	// use ComponentConfig + ConfigMaps.
	ExtraArgs map[string]string `json:"extraArgs,inline"`

	// ExtraVolumes is an extra set of host volumes, mounted to the control plane component.
	//ExtraVolumes []HostPathMount
}
type ControlPlaneComponentExtra struct {
	// ExtraArgs is an extra set of flags to pass to the control plane component.
	// A key in this map is the flag name as it appears on the
	// command line except without leading dash(es).
	// TODO: This is temporary and ideally we would like to switch all components to
	// use ComponentConfig + ConfigMaps.
	ExtraArgs map[string]string `json:"extraArgs"`

	// ExtraVolumes is an extra set of host volumes, mounted to the control plane component.
	//ExtraVolumes []HostPathMount
}

type HostPathMount struct {
	// Name of the volume inside the pod template.
	Name string
	// HostPath is the path in the host that will be mounted inside
	// the pod.
	HostPath string
	// MountPath is the path inside the pod where hostPath will be mounted.
	MountPath string
	// ReadOnly controls write access to the volume
	ReadOnly bool
	// PathType is the type of the HostPath.
	PathType v1.HostPathType
}

type DNS struct {
	// Type defines the DNS add-on to be used
	// TODO: Used only in validation over the internal type. Remove with v1beta2 https://github.com/kubernetes/kubeadm/issues/2459
	//Type DNSAddOnType

	// ImageMeta allows to customize the image used for the DNS component
	//ImageMeta `json:",inline"`
}
type DNSAddOnType string

type ImageMeta struct {
	// ImageRepository sets the container registry to pull images from.
	// if not set, the ImageRepository defined in ClusterConfiguration will be used instead.
	ImageRepository string

	// ImageTag allows to specify a tag for the image.
	// In case this value is set, kubeadm does not change automatically the version of the above components during upgrades.
	ImageTag string
}

const KubeadmJoinDefault = `
apiVersion: kubeadm.k8s.io/v1beta2
controlPlane:
  localAPIEndpoint:
    advertiseAddress: ''
  certificateKey: e6a2eb8581237ab72a4f494f30285ec12a9694d750b9785706a83bfcbbbd2204
discovery:
  bootstrapToken:
    apiServerEndpoint: proton-cs.lb.aishu.cn:9443
    token: 783bde.3f89s0fje9f38fhf
    unsafeSkipCAVerification: true
kind: JoinConfiguration
nodeRegistration:
  taints: []
`

const KubeadmInitDefault = `
apiVersion: kubeadm.k8s.io/v1beta2
kind: InitConfiguration
bootstrapTokens:
  - description: another bootstrap token
    token: 783bde.3f89s0fje9f38fhf
nodeRegistration:
  criSocket: /var/run/dockershim.sock
  taints: []
  kubeletExtraArgs:
    allowed-unsafe-sysctls: net.core.somaxconn
    node-ip: 10.4.71.158

localAPIEndpoint:
  advertiseAddress: ''
certificateKey: e6a2eb8581237ab72a4f494f30285ec12a9694d750b9785706a83bfcbbbd2204
`

const KubeadmKubeletDefault = `
apiVersion: kubelet.config.k8s.io/v1beta1
containerLogMaxSize: 10Mi
containerLogMaxFiles: 3
evictionHard:
  nodefs.available: 0%
  imagefs.available: 0%
imageGCHighThresholdPercent: 100
imageGCLowThresholdPercent: 99
kind: KubeletConfiguration
`

// JoinConfiguration
type KubeadmInitDefaultStruct struct {
	APIVersion       string               `json:"apiVersion"`
	Kind             string               `json:"kind"`
	BootstrapTokens  []BootstrapTokens    `json:"bootstrapTokens"`
	NodeRegistration NodeRegistrationInit `json:"nodeRegistration"`
	LocalAPIEndpoint LocalAPIEndpoint     `json:"localAPIEndpoint"`
	CertificateKey   string               `json:"certificateKey"`
}
type BootstrapTokens struct {
	Description string `json:"description"`
	Token       string `json:"token"`
}
type NodeRegistrationInit struct {
	CriSocket        string           `json:"criSocket"`
	Taints           []any            `json:"taints"`
	KubeletExtraArgs KubeletExtraArgs `json:"kubeletExtraArgs"`
}
type KubeletExtraArgs struct {
	AllowedUnsafeSysctls string `json:"allowed-unsafe-sysctls"`
	NodeIP               string `json:"node-ip"`
}
type LocalAPIEndpoint struct {
	AdvertiseAddress string `json:"advertiseAddress"`
}

// JoinConfiguration

type KubeadmJoinDefaultStruct struct {
	APIVersion       string               `json:"apiVersion"`
	ControlPlane     ControlPlane         `json:"controlPlane"`
	Discovery        Discovery            `json:"discovery"`
	Kind             string               `json:"kind"`
	NodeRegistration NodeRegistrationJoin `json:"nodeRegistration"`
}

type ControlPlane struct {
	LocalAPIEndpoint LocalAPIEndpoint `json:"localAPIEndpoint"`
	CertificateKey   string           `json:"certificateKey"`
}
type BootstrapToken struct {
	Token                    string `json:"token"`
	UnsafeSkipCAVerification bool   `json:"unsafeSkipCAVerification"`
	ApiServerEndpoint        string `json:"apiServerEndpoint"`
}
type Discovery struct {
	BootstrapToken BootstrapToken `json:"bootstrapToken"`
}

type NodeRegistrationJoin struct {
	Taints []any `json:"taints"`
}

// KubeletConfiguration

type KubeadmKubeletDefaultStruct struct {
	APIVersion                  string       `json:"apiVersion"`
	EvictionHard                EvictionHard `json:"evictionHard"`
	ImageGCHighThresholdPercent int          `json:"imageGCHighThresholdPercent"`
	ImageGCLowThresholdPercent  int          `json:"imageGCLowThresholdPercent"`
	Kind                        string       `json:"kind"`
}
type EvictionHard struct {
	NodefsAvailable  string `json:"nodefs.available"`
	ImagefsAvailable string `json:"imagefs.available"`
}

// kube config struct /etc/kubernetes/kubelet.conf
type KubeConfig struct {
	ApiVersion      string        `json:"apiVersion"`
	Clusters        []ClusterInfo `json:"clusters"`
	Contexts        []ContextInfo `json:"contexts"`
	Current_context string        `json:"current-context"`
	Kind            string        `json:"kind"`
	Preferences     any           `json:"preferences"`
	Users           []UserInfo    `json:"users"`
}

type ClusterInfo struct {
	Cluster Cluster `json:"cluster"`
	Name    string  `json:"name"`
}
type Cluster struct {
	Certificate_authority_data string `json:"certificate-authority-data"`
	Server                     string `json:"server"`
}

type ContextInfo struct {
	Context Context `json:"context"`
	Name    string  `json:"name"`
}
type Context struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

type UserInfo struct {
	Name string `json:"name"`
	User User   `json:"user"`
}
type User struct {
	Client_certificate      string `json:"client-certificate,omitempty"`
	Client_key              string `json:"client-key,omitempty"`
	Client_certificate_data string `json:"client-certificate-data,omitempty"`
	Client_key_data         string `json:"client-key-data,omitempty"`
}

// 基础组件连接信杯
type ResourceConnectInfo struct {
	Rds          *RdsInfo          `json:"rds,omitempty"`
	Mongodb      *MongodbInfo      `json:"mongodb,omitempty"`
	Redis        *RedisInfo        `json:"redis,omitempty"`
	Mq           *MqInfo           `json:"mq,omitempty"`
	OpenSearch   *OpensearchInfo   `json:"opensearch,omitempty"`
	PolicyEngine *PolicyEngineInfo `json:"policy_engine,omitempty"`
	Etcd         *EtcdInfo         `json:"etcd,omitempty"`
}

type SourceType string

const (
	Internal SourceType = "internal"
	External SourceType = "external"
)

// Valid RDS type
type RDSType string

const (
	// 支挝的关系型数杮库
	DM8      RDSType = "DM8"
	MySQL    RDSType = "MySQL"
	MariaDB  RDSType = "MariaDB"
	GoldenDB RDSType = "GoldenDB"
	TiDB     RDSType = "TiDB"
	TaurusDB RDSType = "TaurusDB"
	KDB9     RDSType = "KDB9"
)

// 当前支持的关系型数杮类型集
var RdsTypeList = []RDSType{DM8, MariaDB, MySQL, GoldenDB, TiDB, TaurusDB, KDB9}

type RdsInfo struct {
	SourceType SourceType `json:"source_type,omitempty"`

	RdsType RDSType `json:"rds_type,omitempty" mapstructure:"dbType" `

	Hosts    string `json:"hosts,omitempty" mapstructure:"host"`
	Port     int    `json:"port,omitempty" mapstructure:"port"`
	Username string `json:"username,omitempty" mapstructure:"user"`
	Password string `json:"password,omitempty" mapstructure:"password"`

	// 内置时扝有的字段
	HostsRead string `json:"hosts_read,omitempty" mapstructure:"hostRead"`
	PortRead  int    `json:"port_read,omitempty" mapstructure:"portRead"`

	// MgmtInfo
	MgmtHost string `json:"mgmt_host,omitempty" mapstructure:"mgmt_host"`
	MgmtPort int    `json:"mgmt_port,omitempty" mapstructure:"mgmt_port"`
	AdminKey string `json:"admin_key,omitempty" mapstructure:"admin_key"`
}

type MongodbInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`

	Hosts      string `json:"hosts,omitempty" mapstructure:"host"`
	Port       int    `json:"port,omitempty" mapstructure:"port"`
	ReplicaSet string `json:"replica_set,omitempty" mapstructure:"replicaSet"`
	Username   string `json:"username,omitempty" mapstructure:"user"`
	Password   string `json:"password,omitempty" mapstructure:"password"`
	SSL        bool   `json:"ssl,omitempty" mapstructure:"ssl"`
	AuthSource string `json:"auth_source,omitempty" mapstructure:"authSource"`
	Options    any    `json:"options,omitempty" mapstructure:"options"`

	// MgmtInfo
	MgmtHost string `json:"mgmt_host,omitempty" mapstructure:"mgmt_host"`
	MgmtPort int    `json:"mgmt_port,omitempty" mapstructure:"mgmt_port"`
	AdminKey string `json:"admin_key,omitempty" mapstructure:"admin_key"`
}

type ConnectType string

const (
	SentinelMode     ConnectType = "sentinel"
	MasterSlaverMode ConnectType = "master-slave"
	StandAlonelMode  ConnectType = "standalone"
	ClusterMode      ConnectType = "cluster"
)

type RedisInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"connectType"`

	ConnectType ConnectType `json:"connect_type,omitempty"`

	Username string `json:"username,omitempty" mapstructure:"username"`
	Password string `json:"password,omitempty" mapstructure:"password"`

	// masterSlave 字段
	MasterHosts string `json:"master_hosts,omitempty" mapstructure:"masterHost"`
	MasterPort  int    `json:"master_port,omitempty" mapstructure:"masterPort"`
	SlaveHosts  string `json:"slave_hosts,omitempty" mapstructure:"slaveHost"`
	SlavePort   int    `json:"slave_port,omitempty" mapstructure:"slavePort"`
	// sentinel字段
	SentinelHosts    string `json:"sentinel_hosts,omitempty" mapstructure:"sentinelHost"`
	SentinelPort     int    `json:"sentinel_port,omitempty" mapstructure:"sentinelPort"`
	SentinelUsername string `json:"sentinel_username,omitempty" mapstructure:"sentinelUsername"`
	SentinelPassword string `json:"sentinel_password,omitempty" mapstructure:"sentinelPassword"`
	MasterGroupName  string `json:"master_group_name,omitempty" mapstructure:"masterGroupName"`
	// standAlone字段
	Hosts string `json:"hosts,omitempty" mapstructure:"host"`
	Port  int    `json:"port,omitempty" mapstructure:"port"`
}

type MqType string

type MechanType string

const (
	// 支挝的mq类型
	Nsq       MqType = "nsq"
	Tonglink  MqType = "tonglink"
	Htp2      MqType = "htp20"
	Htp202    MqType = "htp202"
	KafkaType MqType = "kafka"
	Besmq     MqType = "bmq"

	//支挝的算法
	Sha512 MechanType = "SCRAM-SHA-512"
	Sha256 MechanType = "SCRAM-SHA-256"
	Plain  MechanType = "PLAIN"
)

// 当剝支挝的消杯中间件类型集
var MqTypeList = [6]MqType{Nsq, Tonglink, Htp2, Htp202, KafkaType, Besmq}

// 当剝支挝的tls坝议加密类型
var MechanTypeList = [3]MechanType{Plain, Sha256, Sha512}

type MqInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`

	MqType MqType `json:"mq_type,omitempty" mapstructure:"mqType"`

	MqHosts        string `json:"mq_hosts,omitempty" mapstructure:"mqHost"`
	MqPort         int    `json:"mq_port,omitempty" mapstructure:"mqPort"`
	MqLookupdHosts string `json:"mq_lookupd_hosts,omitempty" mapstructure:"mqLookupdHost"`
	// lookupd坪有nsq使用，其它中间件丝填
	MqLookupdPort int `json:"mq_lookupd_port,omitempty" mapstructure:"mqLookupdPort"`
	// 当剝坪有kafka支挝酝置auth，其它中间件丝填
	Auth *Auth `json:"auth,omitempty" mapstructure:"auth"`
}

type Auth struct {
	Username string `json:"username,omitempty" mapstructure:"username"`
	Password string `json:"password,omitempty" mapstructure:"password"`

	Mechanism MechanType `json:"mechanism,omitempty" mapstructure:"mechanism"`
}

type OpensearchInfo struct {
	SourceType   SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`
	Hosts        string     `json:"hosts,omitempty" mapstructure:"host"`
	Port         int        `json:"port,omitempty" mapstructure:"port"`
	Username     string     `json:"username,omitempty" mapstructure:"user"`
	Password     string     `json:"password,omitempty" mapstructure:"password"`
	Protocol     string     `json:"protocol,omitempty" mapstructure:"protocol"`
	Version      Version    `json:"version,omitempty" mapstructure:"version"`
	Distribution string     `json:"distribution,omitempty" mapstructure:"distribution"`
}

type PolicyEngineInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`
	Hosts      string     `json:"hosts,omitempty" mapstructure:"host"`
	Port       int        `json:"port,omitempty" mapstructure:"port"`
}

type EtcdInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`

	Hosts  string `json:"hosts,omitempty" mapstructure:"host"`
	Port   int    `json:"port,omitempty" mapstructure:"port"`
	Secret string `json:"secret,omitempty" mapstructure:"secret"`
}

type ProtonCliEnvConfig struct {
	ResourceNamespace        string `json:"resource_namespace,omitempty"`
	ProtonCliConfigNamespace string `json:"proton_cli_config_namespace,omitempty"`
}
