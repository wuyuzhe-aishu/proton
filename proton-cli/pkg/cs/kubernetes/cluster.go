package kubernetes

import (
	"fmt"
	"net"
	"regexp"
	"sync"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/constants"
)

type ChartmuseumInfo struct {
	Address  string
	Username string
	Password string
}

type KubernetesCluster struct {
	Logger             *logrus.Logger
	ImageRepository    string
	kubernetesVersion  string
	Masters            []Node
	Workers            []Node
	ChartRepo          *ChartmuseumInfo
	BIP                string
	Identify           string
	InsecureRegistries []string
	ETCDDataDir        string
	LoadBalancer       string
	IPv4PodCIDR        string
	IPv6PodCIDR        string
	IPv4ServiceCIDR    string
	IPv6ServiceCIDR    string
	SSHPort            int
	SSHUser            string
	SSHPasswd          string
	IPv6Interface      string
	IPv4Interface      string
	CoreDNS            *coreDNS
	Calico             *Calico
	Tiller             *Tiller
	errors             []error
	ContainerRuntime   *configuration.ContainerRuntimeSource
}

var (
	mu sync.Mutex
	wg sync.WaitGroup
)

func (kc *KubernetesCluster) getKubeadminInitYaml() ([]byte, error) {
	criSocket, err := getCRISocketFromContainerRuntimeSource(kc.ContainerRuntime)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`:\d+$`)
	loadBalancerIP := re.ReplaceAllString(kc.LoadBalancer, "")
	kubeConfig := &kubeadmInitConfig{
		Version:         kc.kubernetesVersion,
		ImageRepository: kc.ImageRepository,
		ETCDDataDir:     kc.ETCDDataDir,
		LoadBalancer:    kc.LoadBalancer,
		LoadBalancerIP:  loadBalancerIP,
		IPv4PodCIDR:     kc.IPv4PodCIDR,
		IPv6PodCIDR:     kc.IPv6PodCIDR,
		IPv4ServiceCIDR: kc.IPv4ServiceCIDR,
		IPv6ServiceCIDR: kc.IPv6ServiceCIDR,
		NodeIP:          kc.Masters[0].Ipaddresses,
		CRISocket:       criSocket,
	}
	if err := kubeConfig.renderTemplate(kubeadmInitYamlTemplate); err != nil {
		return nil, fmt.Errorf("render template failed: %v", err)
	}

	return kubeConfig.TemplateYAML, nil
}

func (kc *KubernetesCluster) getKubeadminJoinYaml(token, certhash, certkey, nodeip string) ([]byte, error) {
	return GenerateKubeadmJoinYAML(kc.ContainerRuntime, kc.LoadBalancer, token, certhash, certkey, nodeip)
}

func GenerateKubeadmJoinYAML(r *configuration.ContainerRuntimeSource, loadBalancer, token, certHash, certKey, nodeIP string) ([]byte, error) {
	s, err := getCRISocketFromContainerRuntimeSource(r)
	if err != nil {
		return nil, err
	}

	c := &kubeadmJoinConfig{
		LoadBalancer: loadBalancer,
		Token:        token,
		CertHash:     certHash,
		CertKey:      certKey,
		NodeIP:       nodeIP,
		CRISocket:    s,
	}

	if err := c.renderTemplate(kubeadmJoinYamlTemplate); err != nil {
		return nil, fmt.Errorf("render template failed: %v", err)
	}

	return c.TemplateYAML, nil
}

func (kc *KubernetesCluster) getCoreDNSConfig() error {
	var ips, ipFamilies []string
	if kc.IPv4ServiceCIDR != "" {
		_, network, err := net.ParseCIDR(kc.IPv4ServiceCIDR)
		if err != nil {
			return fmt.Errorf("parse service cidr %s failed: %v", kc.IPv4ServiceCIDR, err)
		}
		firstIP := network.IP.To4()
		for i := 1; i <= 10; i++ {
			firstIP[3]++
		}
		ips = append(ips, firstIP.String())
		ipFamilies = append(ipFamilies, "IPv4")
	}
	if kc.IPv6ServiceCIDR != "" {
		_, network, err := net.ParseCIDR(kc.IPv6ServiceCIDR)
		if err != nil {
			return fmt.Errorf("parse service cidr %s failed: %v", kc.IPv6ServiceCIDR, err)
		}
		firstIP := network.IP.To16()
		for i := 1; i <= 10; i++ {
			firstIP[15]++
		}
		ips = append(ips, firstIP.String())
		ipFamilies = append(ipFamilies, "IPv6")
	}
	var ipFamilyp string
	if len(ipFamilies) > 1 {
		ipFamilyp = "PreferDualStack"
	} else {
		ipFamilyp = "SingleStack"
	}
	kc.CoreDNS = &coreDNS{
		ClusterIP:       ips[0],
		ClusterIPs:      ips,
		IPFamilies:      ipFamilies,
		IPFamilyPolicy:  ipFamilyp,
		ImageRepository: kc.ImageRepository,
	}
	return nil
}

func (kc *KubernetesCluster) getTillerConfig() error {
	var ipFamilies []string
	if kc.IPv4ServiceCIDR != "" {
		_, _, err := net.ParseCIDR(kc.IPv4ServiceCIDR)
		if err != nil {
			return fmt.Errorf("parse service cidr %s failed: %v", kc.IPv4ServiceCIDR, err)
		}
		ipFamilies = append(ipFamilies, "IPv4")
	}
	if kc.IPv6ServiceCIDR != "" {
		_, _, err := net.ParseCIDR(kc.IPv6ServiceCIDR)
		if err != nil {
			return fmt.Errorf("parse service cidr %s failed: %v", kc.IPv6ServiceCIDR, err)
		}
		ipFamilies = append(ipFamilies, "IPv6")
	}
	var ipFamilyp string
	if len(ipFamilies) > 1 {
		ipFamilyp = "PreferDualStack"
	} else {
		ipFamilyp = "SingleStack"
	}
	kc.Tiller = &Tiller{
		ImageRepository: kc.ImageRepository,
		IPFamilies:      ipFamilies,
		IPFamilyPolicy:  ipFamilyp,
	}
	return nil
}

func (kc *KubernetesCluster) InitCluster() error {
	kc.Logger.Info("init kubernetes cluster begin")
	if kc.kubernetesVersion == "" {
		kc.kubernetesVersion = "v1.34.6"
		kc.Logger.Infof("kubernetes version is empty, use default %s", kc.kubernetesVersion)
	}
	master := kc.Masters[0]
	var joinMasters, joinWorkers []Node
	masterHostNames := make(map[string]bool)
	for _, node := range kc.Masters {
		masterHostNames[node.HostName] = true
		if node.HostName != master.HostName {
			joinMasters = append(joinMasters, node)
		}
	}
	for _, node := range kc.Workers {
		if !masterHostNames[node.HostName] {
			joinWorkers = append(joinWorkers, node)
		}
	}

	kc.Logger.Info("init node envirionments")
	var workers []Node
	for _, node := range kc.Workers {
		kc.Logger.Printf("initial %s node info\n", node.Ipaddress)
		if err := node.InitOS(); err != nil {
			return fmt.Errorf("initial %s node envirionments failed: %v", node.Ipaddress, err)
		}
		if err := node.InitialNodeInfo(); err != nil {
			return fmt.Errorf("initial %s node envirionments failed: %v", node.Ipaddress, err)
		}
		if err := node.InitialContainerRuntime(kc.ContainerRuntime); err != nil {
			return fmt.Errorf("initial %s node container runtime failed: %v", node.Ipaddress, err)
		}
		workers = append(workers, node)
	}
	kc.Workers = workers

	var hostNameTmp = make(map[string]string)
	var productUUIDTmp = make(map[string]string)
	for _, node := range kc.Workers {
		//check hostname/product uuid is not same
		if hostNameTmp[node.HostName] != "" {
			return fmt.Errorf("duplicate hostname: %s with %s", node.HostName, node.Ipaddress)
		} else {
			hostNameTmp[node.HostName] = node.Ipaddress
		}

		if productUUIDTmp[node.ProductUUID] != "" {
			return fmt.Errorf("duplicate product UUID: %s with %s, exist: %s", node.ProductUUID, node.Ipaddress, productUUIDTmp[node.ProductUUID])
		} else {
			productUUIDTmp[node.ProductUUID] = node.Ipaddress
		}
	}

	kc.Logger.Info("Initializing operating system")
	kubeadmConfig, err := kc.getKubeadminInitYaml()
	if err != nil {
		return fmt.Errorf("get kubeadm configuration failed: %v", err)
	}
	if err := kc.getCoreDNSConfig(); err != nil {
		return fmt.Errorf("get CoreDNS config failed: %v", err)
	}
	if err := kc.getTillerConfig(); err != nil {
		return fmt.Errorf("get Tiller config failed: %v", err)
	}
	kc.Calico = &Calico{
		Version:          "v3.25.2",
		CurrentVersion:   "v3.25.2",
		ImageRepository:  kc.ImageRepository,
		IPv6Interface:    kc.IPv6Interface,
		PodNetworkCIDRv6: kc.IPv6PodCIDR,
		PodNetworkCIDRv4: kc.IPv4PodCIDR,
	}
	kc.Calico.CalicoVethMTU, err = master.GetInterfaceMTU()
	if err != nil {
		return fmt.Errorf("%s: get master interface mtu failed: %v", master.HostName, err)
	}
	kc.Calico.IptablesBackend, err = master.GetIptablesBackend()
	if err != nil {
		return fmt.Errorf("%s: get master iptables version failed: %v", master.HostName, err)
	}
	if err := master.InitKubeadm(kubeadmConfig); err != nil {
		return fmt.Errorf("%s: init kubernetes cluster failed: %v", master.HostName, err)
	}
	if err := master.InitCalico(kc.Calico); err != nil {
		return fmt.Errorf("%s: init kubernetes calico failed: %v", master.HostName, err)
	}
	if err := master.InitCoreDNS(kc.CoreDNS); err != nil {
		return fmt.Errorf("%s: init kubernetes coredns failed: %v", master.HostName, err)
	}
	if err := master.EnableKubeletService(); err != nil {
		return fmt.Errorf("%s: enable kubelet service failed: %v", master.HostName, err)
	}

	masterToken, masterCerthash, masterCertkey, err := master.GetJoinKeys(true)
	if err != nil {
		return fmt.Errorf("get kubeadm join cluster master role command failed: %v", err)
	}

	workerToken, workerCerthash, workerCertkey, err := master.GetJoinKeys(false)
	if err != nil {
		return fmt.Errorf("get kubeadm join cluster master role command failed: %v", err)
	}

	for _, node := range joinMasters {
		joinMasterConfig, err := kc.getKubeadminJoinYaml(masterToken, masterCerthash, masterCertkey, node.Ipaddress)
		if err != nil {
			return fmt.Errorf("get kubeadm join cluster master role command failed: %v", err)
		}
		if err := node.RewriteKubeletConfig(); err != nil {
			return fmt.Errorf("%s: rewrite kubelet config failed: %v", node.HostName, err)
		}
		if err := node.JoinKubernetesWithYaml(joinMasterConfig); err != nil {
			return fmt.Errorf("%s: join to cluster master role failed: %v", node.HostName, err)
		}
		if err := node.EnableKubeletService(); err != nil {
			return fmt.Errorf("%s: enable kubelet service failed: %v", node.HostName, err)
		}
	}

	for _, node := range joinWorkers {
		wg.Add(1)
		go func(node *Node) {
			defer wg.Done()
			joinWorkerConfig, err := kc.getKubeadminJoinYaml(workerToken, workerCerthash, workerCertkey, node.Ipaddress)
			if err != nil {
				kc.logErr(err, node.Ipaddress, "get join kubernetes kubeadm config")
			}
			if err := node.RewriteKubeletConfig(); err != nil {
				kc.logErr(err, node.Ipaddress, "rewrite kubelet config")
			}
			if err := node.JoinKubernetesWithYaml(joinWorkerConfig); err != nil {
				kc.logErr(err, node.Ipaddress, "join kubernetes cluster")
			}
			if err := node.EnableKubeletService(); err != nil {
				kc.logErr(err, node.Ipaddress, "enable kubelet service")
			}
		}(&node)
	}
	wg.Wait()
	if len(kc.errors) > 0 {
		return fmt.Errorf("%s", kc.errors)
	}
	master.RemoveTaint(kc.Workers)

	kc.Logger.Info("init kubernetes cluster end")
	return nil
}

func (kc *KubernetesCluster) logErr(err error, ip string, action string) {
	mu.Lock()
	defer mu.Unlock()
	kc.Logger.Infof("%s: failed %s: %v", ip, action, err)
	kc.errors = append(kc.errors, fmt.Errorf("%s: failed %s: %v", ip, action, err))
}

func getCRISocketFromContainerRuntimeSource(s *configuration.ContainerRuntimeSource) (socket string, err error) {
	switch {
	case s.Containerd != nil:
		socket = constants.CRISocketContainerd
	default:
		err = fmt.Errorf("unsupported container runtime source: %v", s)
	}
	return
}
