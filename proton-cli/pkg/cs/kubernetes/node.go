package kubernetes

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	os_exec "os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

type Node struct {
	Logger      *logrus.Logger
	HostName    string
	Ipaddress   string
	Ipaddresses string
	ProductUUID string
	CPUNumber   int
	MemorySize  int
	ECMS        ecms.Interface
}

type Repo struct {
	Name string `yaml:"Name"`
	URL  string `yaml:"URL"`
}

var (
	CSSysctl = map[string]string{
		"net.bridge.bridge-nf-call-ip6tables": "1",
		"net.bridge.bridge-nf-call-iptables":  "1",
		"net.ipv4.ip_forward":                 "1",
		"net.ipv4.conf.all.rp_filter":         "0",
		"net.ipv6.conf.all.forwarding":        "1",
		"net.ipv4.conf.all.forwarding":        "1",
		"fs.inotify.max_user_instances":       "8192",
		"fs.inotify.max_user_watches":         "524288",
	}
	CSSysctlPath             = "/usr/lib/sysctl.d/proton-cs.conf"
	IPSetName                = "proton-cs-host"
	FirewallZoneName         = "proton-cs"
	KUBEADM_CONFIG_PATH      = "/tmp/kubeadm-config.yaml"
	KUBEADM_JOIN_CONFIG_PATH = "/tmp/kubeadm-join.yaml"
	COREDNS_CONFIG_PATH      = "/tmp/coredns.yaml"
	TILLER_CONFIG_PATH       = "/tmp/tiller.yaml"
	CALICO_CONFIG_PATH       = "/tmp/calico.yaml"
	KUBE_JOIN_CONFIG_PATH    = "/tmp/kubeadm-join.yaml"
	KUBELET_KUBEADM_CONTENT  = `KUBELET_KUBEADM_ARGS="--network-plugin=cni --pod-infra-container-image=registry.aishu.cn:15000/public/pause:3.6"`
	KUBELET_KUBEADM_PATH     = "/var/lib/kubelet/kubeadm-flags.env"
	KUBELET_KUBEADM_DIR      = "/var/lib/kubelet"
	PROTON_CS_ZONE           = `<?xml version="1.0" encoding="utf-8"?>
<zone target="ACCEPT">
  <interface name="cali+"/>
  <interface name="tunl0"/>
  <source ipset="proton-cs-host"/>
  <source ipset="proton-cs-host6"/>
</zone>
`
	PROTON_CS_HOST_IPSET = `<?xml version="1.0" encoding="utf-8"?>
<ipset type="hash:ip">
</ipset>
`
	PROTON_CS_HOST6_IPSET = `<?xml version="1.0" encoding="utf-8"?>
<ipset type="hash:ip">
  <option name="family" value="inet6">
  </option>
</ipset>
`
	PROTON_CS_ZONE_PATH        = "/etc/firewalld/zones/proton-cs.xml"
	PROTON_CS_HOST_IPSET_PATH  = "/etc/firewalld/ipsets/proton-cs-host.xml"
	PROTON_CS_HOST6_IPSET_PATH = "/etc/firewalld/ipsets/proton-cs-host6.xml"
)

func NewNode(logger *logrus.Logger, hostname, ipaddress, ipaddresses string) (*Node, error) {
	// Always use IPv4 address for SSH connection in dual-stack setup
	connectIP := ipaddress
	if strings.Contains(ipaddress, ",") {
		// If dual-stack, use the IPv4 address for connection
		ips := strings.SplitSeq(ipaddress, ",")
		for ip := range ips {
			if !strings.Contains(ip, ":") {
				connectIP = ip
				break
			}
		}
	}

	return &Node{
		Logger:      logger,
		HostName:    hostname,
		Ipaddress:   connectIP,   // Use IPv4 for direct communication
		Ipaddresses: ipaddresses, // Keep both addresses for Kubernetes config
		ECMS:        ecms.NewForHost(connectIP),
	}, nil
}

func (n *Node) InitialNodeInfo() error {
	var ctx = context.TODO()
	// executor
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())

	cpu_num, err := executor.Command("getconf", "_NPROCESSORS_CONF").Output()
	if err != nil {
		return fmt.Errorf("%s: failed to get cpu number: %v", n.Ipaddress, err)
	}
	n.CPUNumber, _ = strconv.Atoi(strings.TrimSpace(string(cpu_num)))

	hostName, err := executor.Command("hostname").Output()
	if err != nil {
		return fmt.Errorf("%s: failed to get hostname: %v", n.Ipaddress, err)
	}
	n.HostName = strings.TrimSpace(string(hostName))

	var memSize string
	{
		out, err := executor.Command("free", "-m").Output()
		if err != nil {
			return fmt.Errorf("%s: failed to get memory size: %v", n.Ipaddress, err)
		}
		s := bufio.NewScanner(bytes.NewReader(out))
		for s.Scan() {
			fields := strings.Fields(s.Text())
			if fields[0] != "Mem:" {
				continue
			}
			memSize = fields[1]
		}
	}
	n.MemorySize, _ = strconv.Atoi(strings.TrimSpace(string(memSize)))

	productUUID, err := n.ECMS.Files().ReadFile(ctx, "/sys/class/dmi/id/product_uuid")
	if err != nil {
		return fmt.Errorf("%s: failed to get product uuid: %v", n.Ipaddress, err)
	}
	n.ProductUUID = strings.TrimSpace(string(productUUID))

	return nil
}

func (n *Node) InitOS() error {
	var ctx = context.TODO()
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	n.Logger.Printf("%s: setting kernel modprobe", n.Ipaddress)
	if err := executor.Command("modprobe", "--quiet", "br_netfilter").Run(); err != nil {
		return fmt.Errorf("%s: failed to run modprobe: %v", n.Ipaddress, err)
	}

	n.Logger.Printf("%s: setting sysctl", n.Ipaddress)
	var sb strings.Builder
	for key, value := range CSSysctl {
		sb.WriteString(fmt.Sprintf("%s = %s\n", key, value))
	}
	sysctlContent := sb.String()
	if err := n.ECMS.Files().Create(ctx, CSSysctlPath, false, []byte(sysctlContent)); err != nil {
		return err
	}
	if err := executor.Command("sysctl", "--load", CSSysctlPath).Run(); err != nil {
		return fmt.Errorf("%s: failed to load sysctl config: %v", n.Ipaddress, err)
	}

	n.Logger.Printf("%s: setting selinux disabled", n.Ipaddress)
	output, err := executor.Command("getenforce").Output()
	if errors.Is(err, os_exec.ErrNotFound) {
		// 命令未找到
		n.Logger.Warnf("%s: getenforce command not found (exit code 127), assuming SELinux is not installed", n.Ipaddress)
	} else if err != nil {
		return fmt.Errorf("%s: failed to get selinux status: %v", n.Ipaddress, err)
	}
	if strings.Contains(string(output), "Enforcing") {
		if err := executor.Command("setenforce", "0").Run(); err != nil {
			return fmt.Errorf("%s: failed to disable selinux: %v", n.Ipaddress, err)
		}
		if err := executor.Command("sed", "-i", "s/^SELINUX=.*/SELINUX=disabled/", "/etc/selinux/config").Run(); err != nil {
			return fmt.Errorf("%s: failed to disable selinux: %v", n.Ipaddress, err)
		}
	}

	n.Logger.Printf("%s: setting swapoff", n.Ipaddress)
	if err := executor.Command("swapoff", "--all").Run(); err != nil {
		return fmt.Errorf("%s: failed to disable selinux: %v", n.Ipaddress, err)
	}
	if err := executor.Command("sed", "-i", "/swap/d", "/etc/fstab").Run(); err != nil {
		return fmt.Errorf("%s: failed to disable selinux: %v", n.Ipaddress, err)
	}

	n.Logger.Printf("%s: stop kubelet service", n.Ipaddress)
	if err := executor.Command("systemctl", "stop", "kubelet").Run(); err != nil {
		return fmt.Errorf("%s: failed to disable selinux: %v", n.Ipaddress, err)
	}
	return nil
}

func (n *Node) InitialContainerRuntime(s *configuration.ContainerRuntimeSource) (err error) {
	switch {
	case s.Containerd != nil:
		err = n.InitContainerd(s.Containerd)
	default:
		err = fmt.Errorf("unsupported container runtime source: %v", s)
	}
	return
}

func (n *Node) InitContainerd(s *configuration.ContainerdContainerRuntimeSource) error {
	var ctx = context.TODO()
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	// create server config
	n.Logger.WithField("source", s).Info("create containerd config file")
	if err := createContainerdConfig(n.ECMS.Files(), s); err != nil {
		return err
	}

	// create cri-base.json
	n.Logger.Info("create base runtime spec file")
	if err := createContainerdBaseRuntimeSpecFile(n.ECMS.Files()); err != nil {
		return err
	}

	// create registry host configs
	for _, r := range s.Registries {
		u, err := url.Parse(r.Server)
		if err != nil {
			return err
		}

		dir := filepath.Join("/etc/containerd/certs.d", u.Host)

		n.Logger.WithField("host", r).Debug("create registry host directory")
		if err := n.ECMS.Files().Create(ctx, dir, true, nil); err != nil {
			return err
		}

		n.Logger.WithField("host", r).Debug("create registry host config file")
		if err := createContainerdHostConfigFile(n.ECMS.Files(), dir, &r); err != nil {
			return err
		}

	}

	// start or restart containerd
	n.Logger.WithField("unit", "containerd.service").Info("start or restart containerd")
	if err := executor.Command("systemctl", "restart", "containerd.service").Run(); err != nil {
		return err
	}

	// enable containerd
	n.Logger.WithField("unit", "containerd.service").Info("enable containerd")
	if err := executor.Command("systemctl", "enable", "containerd.service").Run(); err != nil {
		return err
	}

	return nil
}

func (n *Node) RewriteKubeletConfig() error {
	var ctx = context.TODO()

	// change kubelet.conf ARGS, add node-ip and --allow kernel modules
	n.Logger.Infof("%s: enabling kubelet service", n.Ipaddress)
	var kubeletArgs string
	if kubeletConfigContent, err := n.ECMS.Files().ReadFile(ctx, KUBELET_KUBEADM_PATH); err == nil {
		kubeletArgs = strings.Split(strings.TrimRight(string(kubeletConfigContent), "\n"), "KUBELET_KUBEADM_ARGS=")[1]
	} else if errors.Is(err, fs.ErrNotExist) {
		kubeletArgs = strings.Split(strings.TrimRight(KUBELET_KUBEADM_CONTENT, "\n"), "KUBELET_KUBEADM_ARGS=")[1]
	} else {
		return fmt.Errorf("%s: failed to open %s: %v", n.Ipaddress, KUBELET_KUBEADM_PATH, err)
	}

	kubeletArgs = strings.TrimSpace(kubeletArgs)
	args := strings.Split(kubeletArgs, " ")
	if len(args) == 1 {
		args = regexp.MustCompile(`\s+`).Split(kubeletArgs, -1)
	}
	nodeIP := fmt.Sprintf("--node-ip=%s", n.Ipaddresses)
	var newKubeletArgs []string
	flag := false
	sysctlFlag := false
	for _, arg := range args {
		if strings.HasPrefix(strings.Trim(arg, "\""), "--node-ip=") {
			newKubeletArgs = append(newKubeletArgs, nodeIP)
			flag = true
		} else if strings.HasPrefix(strings.Trim(arg, "\""), "--allowed-unsafe-sysctls=") {
			sysctlFlag = true
		} else {
			newKubeletArgs = append(newKubeletArgs, strings.Trim(arg, "\""))
		}
	}
	if !flag {
		_ = os.Remove("/etc/sysconfig/kubelet")
		newKubeletArgs = append(newKubeletArgs, nodeIP)
	}
	if !sysctlFlag {
		newKubeletArgs = append(newKubeletArgs, "--allowed-unsafe-sysctls=net.core.somaxconn")
	}
	kubeletEnv := "KUBELET_KUBEADM_ARGS=\"" + strings.Join(newKubeletArgs, " ") + "\""
	n.Logger.Infof("%s: new kubelet args: %s", n.Ipaddress, kubeletEnv)
	if err := n.ECMS.Files().Create(ctx, KUBELET_KUBEADM_DIR, true, nil); err != nil {
		return fmt.Errorf("%s: failed to create %s: %v", n.Ipaddress, KUBELET_KUBEADM_DIR, err)
	}
	if err := n.ECMS.Files().Create(ctx, KUBELET_KUBEADM_PATH, false, []byte(kubeletEnv)); err != nil {
		return fmt.Errorf("%s: failed to write %s: %v", n.Ipaddress, KUBELET_KUBEADM_PATH, err)
	}
	return nil
}

func (n *Node) EnableKubeletService() error {
	n.Logger.Printf("%s: enabling kubelet service", n.Ipaddress)
	if err := exec.NewECMSExecutorForHost(n.ECMS.Exec()).Command("systemctl", "enable", "kubelet.service").Run(); err != nil {
		return fmt.Errorf("%s: failed to enable kubelet service: %v", n.Ipaddress, err)
	}
	return nil
}

func (n *Node) InitKubeadm(kubeadmConfig []byte) error {
	/*
		Set up the Kubernetes control plane

		1. Generate kubeadm configuration
		2. Write kubeadm configuration to /tmp/kubeadm-config.yaml
		3. Run command `kubeadm init --skip-phases=addon/coredns --config=/tmp/kubeadm-config.yaml
		4. Copy /etc/kubernetes/admin.conf to $HOME/.kube/config
	*/
	var ctx = context.TODO()
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	n.Logger.Printf("%s: initial kubernete with kubeadm begin", n.Ipaddress)
	if err := n.ECMS.Files().Create(ctx, KUBEADM_CONFIG_PATH, false, kubeadmConfig); err != nil {
		return fmt.Errorf("%s: failed to write kubeadm config: %v", n.Ipaddress, err)

	}
	if err := executor.Command("kubeadm", "init", "--skip-phases=addon/coredns", "--config", KUBEADM_CONFIG_PATH, "--upload-certs").Run(); err != nil {
		return fmt.Errorf("%s: failed to run kubeadm init: %v", n.Ipaddress, err)
	}

	// create root's kube config
	if err := n.ECMS.Files().Create(ctx, "/root/.kube", true, nil); err != nil {
		return err
	}
	if err := executor.Command("cp", "/etc/kubernetes/admin.conf", "/root/.kube/config").Run(); err != nil {
		return fmt.Errorf("%s: failed to copy kubeconfig: %v", n.Ipaddress, err)
	}

	if err := executor.Command("kubeadm", "init", "phase", "upload-config", "all", "--config", KUBEADM_CONFIG_PATH).Run(); err != nil {
		return fmt.Errorf("%s: failed to upload-config: %v", n.Ipaddress, err)
	}

	if err := executor.Command("kubeadm", "init", "phase", "bootstrap-token", "--config", KUBEADM_CONFIG_PATH).Run(); err != nil {
		return fmt.Errorf("%s: failed to bootstrap-token: %v", n.Ipaddress, err)
	}

	// create root's kube config
	if err := n.ECMS.Files().Create(ctx, "/root/.kube", true, nil); err != nil {
		return err
	}

	// 创建 root 用户的 kubeconfig 因为已有的很多逻辑都依赖使用 root 用户执行 kubectl
	if err := executor.Command("cp", "/etc/kubernetes/admin.conf", "/root/.kube/config").Run(); err != nil {
		return fmt.Errorf("%s: failed to copy kubeconfig: %v", n.Ipaddress, err)
	}

	// 获取当前用户，认为所有 kubernetes control plane 节点都有这个用户
	u, err := user.Current()
	if err != nil {
		return err
	}

	kubeDir := filepath.Join(u.HomeDir, ".kube")
	// TODO: 假定所有节点的用户 u 的 home 目录都相同。应该通过系统命令、syscall查询，或通过配置指定。
	if err := n.ECMS.Files().Create(ctx, kubeDir, true, nil); err != nil {
		return err
	}
	// 修改 ~/.kube 的所有者为“当前用户”
	if err := executor.Command("chown", u.Uid+":"+u.Gid, kubeDir).Run(); err != nil {
		return err
	}

	kubeconfig := filepath.Join(kubeDir, "config")
	if err := executor.Command("cp", "/etc/kubernetes/admin.conf", kubeconfig).Run(); err != nil {
		return fmt.Errorf("%s: failed to copy kubeconfig: %v", n.Ipaddress, err)
	}

	if err := executor.Command("chown", u.Uid+":"+u.Gid, kubeconfig).Run(); err != nil {
		return fmt.Errorf("%s: failed to change kubeconfig owner: %v", n.Ipaddress, err)
	}

	n.Logger.Printf("%s: initial kubernetes with kubeadm end", n.Ipaddress)

	return nil
}

func (n *Node) InitCoreDNS(corednsCfg *coreDNS) error {
	/*
			Deploy CoreDNS on Kubernetse

		        1. Generate manifest from template
		        2. Write manifest to /tmp/coredns.yaml
		        3. Run command `kubectl apply --file=/tmp/coredns.yaml
	*/
	var ctx = context.TODO()
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	n.Logger.Printf("%s: initial CoreDNS begin", n.Ipaddress)
	if err := corednsCfg.renderTemplate(coreDNSv186YamlTemplate); err != nil {
		return fmt.Errorf("%s: failed to render coredns template: %v", n.Ipaddress, err)
	}

	if err := n.ECMS.Files().Create(ctx, COREDNS_CONFIG_PATH, false, corednsCfg.TemplateYAML); err != nil {
		return fmt.Errorf("%s: failed to write coredns config: %v", n.Ipaddress, err)
	}
	if err := executor.Command("kubectl", "apply", "-f", COREDNS_CONFIG_PATH).Run(); err != nil {
		return fmt.Errorf("%s: failed to apply coredns: %v", n.Ipaddress, err)
	}
	n.Logger.Printf("%s: initial CoreDNS end", n.Ipaddress)

	return nil
}

func (n *Node) InitTiller(tillerCfg *Tiller) error {
	var ctx = context.TODO()
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	n.Logger.Printf("%s: initial Tiller begin", n.Ipaddress)
	if err := tillerCfg.renderTemplate(TillerYamlTemplate); err != nil {
		return fmt.Errorf("%s: failed to render tiller template: %v", n.Ipaddress, err)
	}

	if err := n.ECMS.Files().Create(ctx, TILLER_CONFIG_PATH, false, tillerCfg.TemplateYAML); err != nil {
		return fmt.Errorf("%s: failed to write tiller config: %v", n.Ipaddress, err)
	}
	if err := executor.Command("kubectl", "apply", "-f", TILLER_CONFIG_PATH).Run(); err != nil {
		return fmt.Errorf("%s: failed to apply tiller: %v", n.Ipaddress, err)
	}

	n.Logger.Printf("%s: initial Tiller end", n.Ipaddress)

	return nil
}

func (n *Node) WaitTillerReady() error {
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	n.Logger.Info("waiting for tiller pod ready with helm version in 300 seconds")
	for range 60 {
		if err := executor.Command("helm", "version").Run(); err != nil {
			n.Logger.Printf("%s: tiller not ready, retry in 5 seconds", n.Ipaddress)
			time.Sleep(5 * time.Second)
		} else {
			n.Logger.Printf("%s: tiller ready", n.Ipaddress)
			return nil
		}
	}

	return nil
}

func (n *Node) GetInterfaceMTU() (string, error) {
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	out, err := executor.Command("proton-cli", "alpha", "net-interface-mtu-by-address", n.Ipaddress).Output()
	if err != nil {
		return "", fmt.Errorf("%s: get network interface mtu by address fail: %v", n.Ipaddress, err)
	}
	out = bytes.TrimSpace(out)

	mtu, err := strconv.Atoi(string(out))
	if err != nil {
		return "", err
	}

	// mtu - 60 for tunnel
	return strconv.Itoa(mtu - 60), nil
}

func (n *Node) GetIptablesBackend() (string, error) {
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	output, err := executor.Command("iptables", "-V").Output()
	if err != nil {
		return "", fmt.Errorf("%s: failed to get iptables version: %v", n.Ipaddress, err)
	}
	if strings.Contains(string(output), "nf_tables") {
		return "NFT", nil
	}
	return "Legacy", nil
}

func (n *Node) InitCalico(calico *Calico) error {
	var ctx = context.TODO()
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	n.Logger.Printf("%s: initial Calico begin", n.Ipaddress)
	if err := calico.renderTemplate(calicoV3252YamlTemplate); err != nil {
		return fmt.Errorf("%s: failed to render calico template: %v", n.Ipaddress, err)
	}
	if err := n.ECMS.Files().Create(ctx, CALICO_CONFIG_PATH, false, calico.TemplateYAML); err != nil {
		return fmt.Errorf("%s: failed to write calico config: %v", n.Ipaddress, err)
	}
	if err := executor.Command("kubectl", "apply", "-f", CALICO_CONFIG_PATH).Run(); err != nil {
		return fmt.Errorf("%s: failed to apply calico: %v", n.Ipaddress, err)
	}
	n.Logger.Printf("%s: initial Calico end", n.Ipaddress)

	return nil
}

func (n *Node) GetJoinKeys(isMaster bool) (string, string, string, error) {
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	//return loadbalancer,token,certhash,certkey
	n.Logger.Infof("%s: get join keys kubeadm token create", n.Ipaddress)
	output, err := executor.Command("kubeadm", "token", "create", "--print-join-command").Output()
	if err != nil {
		return "", "", "", fmt.Errorf("%s: failed to get join cmd: %v", n.Ipaddress, err)
	}
	//output example: kubeadm join 10.2.184.19:6443 --token kcr6u0.20kq4vorg8ytvt5k --discovery-token-ca-cert-hash sha256:64348f3a7432d4bef6ab04e068660c82b84cdbd0509457bdb7d02233e540c308
	outputFields := strings.Split(strings.TrimSpace(string(output)), " ")

	if !isMaster {
		return strings.TrimSpace(outputFields[4]), strings.TrimSpace(outputFields[6]), "", nil
	}

	// output:
	// > [upload-certs] Storing the certificates in Secret "kubeadm-certs" in the "kube-system" Namespace
	// > [upload-certs] Using certificate key:
	// > e6a2eb8581237ab72a4f494f30285ec12a9694d750b9785706a83bfcbbbd2204
	if output, err = executor.Command("kubeadm", "init", "phase", "upload-certs", "--upload-certs", "--config=/tmp/kubeadm-config.yaml").Output(); err != nil {
		return "", "", "", fmt.Errorf("%s: failed to get certificate: %v", n.Ipaddress, err)
	}

	s := bufio.NewScanner(bytes.NewReader(output))

	for i := 0; s.Scan(); i++ {
		if i < 2 {
			continue
		}
		return strings.TrimSpace(outputFields[4]), strings.TrimSpace(outputFields[6]), s.Text(), nil
	}
	return "", "", "", fmt.Errorf("%s: failed to get certificate: %v", n.Ipaddress, err)
}

func (n *Node) JoinKubernetesWithYaml(kubeadmConfig []byte) error {
	var ctx = context.TODO()
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	n.Logger.Printf("%s: initial kubernete with kubeadm begin", n.Ipaddress)
	if err := n.ECMS.Files().Create(ctx, KUBEADM_JOIN_CONFIG_PATH, false, kubeadmConfig); err != nil {
		return fmt.Errorf("%s: failed to write kubeadm join config: %v", n.Ipaddress, err)

	}

	if err := executor.Command("kubeadm", "join", "--config", KUBEADM_JOIN_CONFIG_PATH).Run(); err != nil {
		return fmt.Errorf("%s: failed to run kubeadm join: %v", n.Ipaddress, err)
	}

	if strings.Contains(string(kubeadmConfig), "controlPlane") {
		if err := n.ECMS.Files().Create(ctx, "/root/.kube", true, nil); err != nil {
			return err
		}
		if err := executor.Command("cp", "/etc/kubernetes/admin.conf", "/root/.kube/config").Run(); err != nil {
			return fmt.Errorf("%s: failed to copy kubeconfig: %v", n.Ipaddress, err)
		}
	}

	return nil
}

func (n *Node) RemoveTaint(nodes []Node) {
	var executor = exec.NewECMSExecutorForHost(n.ECMS.Exec())
	for _, node := range nodes {
		_ = executor.Command("kubectl", "taint", "nodes", node.HostName, "node-role.kubernetes.io/control-plane:NoSchedule-").Run()
	}
}

// Execute command `rpm --query "${names[@]}"`
func (n *Node) Query(names ...string) ([]string, error) {
	var args []string
	args = append(args, "--query")
	args = append(args, names...)
	out, err := exec.NewECMSExecutorForHost(n.ECMS.Exec()).Command("rpm", args...).Output()
	return strings.Split(strings.TrimSpace(string(out)), "\n"), err
}
