package prometheus

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"path"
	"time"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"

	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	kubernetes_core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
)

// Manager 负责安装、更新服务 Prometheus，清理 Prometheus 的数据目录
type Manager struct {
	// contaienr image registry's address, host or host:port
	registry string
	// prometheus deployment configuraton
	spec *configuration.Prometheus
	// 远程操访问节点的客户端，这里是要部署Prometheus的节点
	nodes []v1alpha1.Interface
	// 所有节点，用于找到主节点
	allNodes []v1alpha1.Interface
	// 调用 helm 的接口
	helm helm3.Client
	// secret 是 Kubernetes 官方客户端接口，用于获取和写入secret中的证书信息
	secret kubernetes_core_v1.SecretInterface
	// csProvisioner记录当前集群为本地部署还是托管K8S部署，作为是否获取K8S内置ETCD证书的依据
	csProvisioner configuration.KubernetesProvisioner
	// isExistProtonETCD记录当前集群是否安装了ProtonETCD，如果未安装的话则不在Proton Prometheus中监控ProtonETCD
	isExistProtonETCD bool
	// masterNodeName 用于传递主节点名称，以选择适当节点读取主机目录中的K8S ETCD CA证书
	masterNodeName string

	servicePackage *servicepackage.ServicePackage
	logger         logrus.FieldLogger
	namespace      string
}

// New returns a Manager
func New(spec *configuration.Prometheus, registry string, nodes []v1alpha1.Interface, helm helm3.Client, logger logrus.FieldLogger, csp configuration.KubernetesProvisioner, mn *configuration.Node, ku kubernetes.Interface, existsProtonETCD bool, alln []v1alpha1.Interface, pkg *servicepackage.ServicePackage, namespace string) *Manager {
	// there are no existing kube client interface during reset so a nil is passed into here during reset
	var mnn string
	if csp == configuration.KubernetesProvisionerLocal {
		mnn = mn.Name
	} else {
		mnn = ""
	}
	s := ku.CoreV1().Secrets(namespace)
	return &Manager{registry: registry, spec: spec, nodes: nodes, helm: helm, logger: logger, secret: s, csProvisioner: csp, masterNodeName: mnn, isExistProtonETCD: existsProtonETCD, allNodes: alln, servicePackage: pkg, namespace: namespace}
}

func (m *Manager) Apply() error {
	m.logger.Info("check environment")
	if err := checkEnvironment(m.spec, m.nodes, m.logger); err != nil {
		return err
	}

	if p := m.spec.DataPath; p != "" {
		m.logger.Info("reconcile data directory")
		for _, n := range m.nodes {
			if err := universal.ReconcileDataDirectory(n, p, m.logger.WithField("node", n.Name())); err != nil {
				return fmt.Errorf("%v: %w", n.Name(), err)
			}
		}
	}

	if _, err := m.secret.Get(context.TODO(), K8SETCDResultSecretName, metav1.GetOptions{}); apierrors.IsNotFound(err) {
		if m.csProvisioner == configuration.KubernetesProvisionerLocal {
			if err := m.GeneratePrometheusK8SETCDCert(); err != nil {
				return err
			}
		}
	}

	// precheck that etcdssl-secret and etcdssl-secret-key exists
	_, errMissingETCDSSL4Prometheus := m.secret.Get(context.TODO(), ProtonETCDResultSecretName, metav1.GetOptions{})
	_, errMissingPE := m.secret.Get(context.TODO(), ProtonETCDCACertSecret, metav1.GetOptions{})
	_, errMissingPEKey := m.secret.Get(context.TODO(), ProtonETCDCACertKey, metav1.GetOptions{})
	missingProtonETCDSecrets := false
	if (apierrors.IsNotFound(errMissingPE) || apierrors.IsNotFound(errMissingPEKey)) && m.isExistProtonETCD {
		m.logger.Warning("One of the K8S secret etcdssl-secret or etcdssl-secret-key is missing in this cluster.")
		m.logger.Warning("Proton Prometheus will not monitor Proton-ETCD unless both the K8S secrets are present.")
		missingProtonETCDSecrets = true
	}
	if m.isExistProtonETCD && apierrors.IsNotFound(errMissingETCDSSL4Prometheus) && !missingProtonETCDSecrets {
		if err := m.GeneratePrometheusProtonETCDCert(); err != nil {
			return err
		}
	}

	m.logger.Info("reconcile helm release prometheus")
	cht := m.servicePackage.Charts().Get(ChartName, "")

	if err := m.helm.Upgrade(
		HelmReleaseName,
		&helm3.ChartRef{File: path.Join(m.servicePackage.BaseDir(), cht.Path)},
		helm3.WithUpgradeValues(m.values().ToMap()),
		helm3.WithUpgradeInstall(true),
	); err != nil {
		return fmt.Errorf("reconfile helm release prometheus fail: %w", err)
	}

	return nil
}

func (m *Manager) Reset() error {
	// The helm client is nil when kubernetes client config has been removed.
	if m.helm == nil {
		m.logger.Debug("kubernetes client config has been removed")
	} else if err := m.helm.Uninstall(HelmReleaseName, helm3.WithUninstallIgnoreNotFound(true)); err != nil {
		// m.logger.WithField("release", HelmReleaseName).Warn("tolerate failure to remove the helm release, because the kubernetes may has already been destroyed")
		return err
	}

	for _, name := range []string{
		ProtonETCDResultSecretName,
		K8SETCDResultSecretName,
	} {
		err := m.deleteSingleSecretTolerateNotFound(context.Background(), name)
		if err != nil {
			return err
		}
	}

	if m.spec.DataPath != "" {
		for _, n := range m.nodes {
			if err := universal.ClearDataDirViaNodeV1Alpha1(n, m.spec.DataPath, m.logger.WithField("node", n.Name())); err != nil {
				return fmt.Errorf("%s: %w", n.Name(), err)
			}
		}
	}

	return nil
}

// KubernetesService returns a kubernetes service object that the prometheus is
// exposed via.
//
// Currently it is generated from the prometheus configuration, not helm chart.
func (m *Manager) KubernetesService() *corev1.Service {
	var service = corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HelmReleaseName,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":                          HelmReleaseName,
				"app.kubernetes.io/managed-by": "helm",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "prometheus",
					Protocol:   corev1.ProtocolTCP,
					Port:       9090,
					TargetPort: intstr.FromString("prometheus"),
				},
			},
			Selector: map[string]string{
				"app": HelmReleaseName,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	if global.EnableDualStack {
		service.Spec.IPFamilies = []corev1.IPFamily{
			corev1.IPv6Protocol,
			corev1.IPv4Protocol,
		}
		ipFamilyPolicy := corev1.IPFamilyPolicyPreferDualStack
		service.Spec.IPFamilyPolicy = &ipFamilyPolicy
	}
	return &service
}

// GeneratePrometheusProtonETCDCert generates mTLS cert for prometheus to access Proton ETCD metrics
func (m *Manager) GeneratePrometheusProtonETCDCert() error {
	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("unable to generate proton-etcd for prometheus private key error: %w", err)
	}
	serverTemplate := newCertTemplate()

	serverTemplate.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	serverTemplate.IsCA = false
	serverTemplate.Subject.CommonName = PrometheusETCDCommonName
	serverTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	// read ProtonETCD CA cert then sign the new cert
	protonETCDCASecret, err := m.secret.Get(context.TODO(), ProtonETCDCACertSecret, meta_v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to read proton-etcd CA cert from secret error: %w", err)
	}
	protonETCDCACertPEM, _ := pem.Decode(protonETCDCASecret.Data[ProtonETCDCACertName])
	protonETCDCACert, err := x509.ParseCertificate(protonETCDCACertPEM.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse proton-etcd CA cert error: %w", err)
	}
	protonETCDCAKeySecret, err := m.secret.Get(context.TODO(), ProtonETCDCACertKey, meta_v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to read proton-etcd CA key from secret error: %w", err)
	}
	protonETCDCAKeyPEM, _ := pem.Decode(protonETCDCAKeySecret.Data[ProtonETCDCAKeyName])
	protonETCDCAPrivateKey, err := x509.ParsePKCS8PrivateKey(protonETCDCAKeyPEM.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse proton-etcd CA key error: %w", err)
	}
	serverCertDer, err := x509.CreateCertificate(rand.Reader, serverTemplate, protonETCDCACert, &serverPrivateKey.PublicKey, protonETCDCAPrivateKey) //DER 格式
	if err != nil {
		return fmt.Errorf("unable to create new proton-etcd for prometheus cert error: %w", err)
	}

	serverPrivateBytes, err := x509.MarshalPKCS8PrivateKey(serverPrivateKey)
	if err != nil {
		return fmt.Errorf("unable to convert proton-etcd for prometheus cert to PKCS #8, ASN.1 DER format error: %w", err)
	}

	// convert the privatekey to PEM format
	serverKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: serverPrivateBytes})

	// convert the cert to PEM format
	serverCrt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDer})

	// create the secret for prometheus and set it into the prometheus chart values (release) later
	expectData := map[string][]byte{
		ProtonETCDResultCAName:   protonETCDCASecret.Data[ProtonETCDCACertName],
		ProtonETCDResultCertName: serverCrt,
		ProtonETCDResultKeyName:  serverKey,
	}
	expectSecret := &corev1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: ProtonETCDResultSecretName,
		},
		Data: expectData,
	}
	_, err = m.secret.Create(context.TODO(), expectSecret, metav1.CreateOptions{})
	return err
}

// GeneratePrometheusK8SETCDCert generates mTLS cert for prometheus to access K8S built-in ETCD metrics
func (m *Manager) GeneratePrometheusK8SETCDCert() error {
	var ctx = context.TODO()
	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("unable to generate k8s-etcd for prometheus key error: %w", err)
	}
	serverTemplate := newCertTemplate()

	serverTemplate.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	serverTemplate.IsCA = false
	serverTemplate.Subject.CommonName = PrometheusETCDCommonName
	serverTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	// read K8S ETCD CA cert then sign the new cert
	var mn v1alpha1.Interface
	for _, n := range m.allNodes {
		if n.Name() == m.masterNodeName {
			mn = n
		}
	}
	if mn == nil {
		return fmt.Errorf("unable to find a master node in current cluster")
	}
	k8sETCDRawText, err := mn.ECMS().Files().ReadFile(ctx, K8SETCDCACertPath)
	if err != nil {
		return fmt.Errorf("unable to read k8s-etcd CA cert error: %w", err)
	}
	k8sETCDCACertPEM, _ := pem.Decode(k8sETCDRawText)
	k8sETCDCACert, err := x509.ParseCertificate(k8sETCDCACertPEM.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse k8s-etcd CA cert error: %w", err)
	}
	k8sETCDCAKeyRawText, err := mn.ECMS().Files().ReadFile(ctx, K8SETCDCACertKey)
	if err != nil {
		return fmt.Errorf("unable to read k8s-etcd CA key from filesystem error: %w", err)
	}
	k8sETCDCAKeyPEM, _ := pem.Decode(k8sETCDCAKeyRawText)
	k8sETCDCAPrivateKey, err := x509.ParsePKCS1PrivateKey(k8sETCDCAKeyPEM.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse k8s-etcd CA key error: %w", err)
	}
	serverCertDer, err := x509.CreateCertificate(rand.Reader, serverTemplate, k8sETCDCACert, &serverPrivateKey.PublicKey, k8sETCDCAPrivateKey) //DER 格式
	if err != nil {
		return fmt.Errorf("unable to create new k8s-etcd for prometheus cert error: %w", err)
	}

	serverPrivateBytes := x509.MarshalPKCS1PrivateKey(serverPrivateKey)

	// convert the privatekey to PEM format
	serverKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: serverPrivateBytes})

	// convert the cert to PEM format
	serverCrt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDer})

	// create the secret for prometheus and set it into the prometheus chart values (release) later
	expectData := map[string][]byte{
		K8SETCDResultCAName:   k8sETCDRawText,
		K8SETCDResultCertName: serverCrt,
		K8SETCDResultKeyName:  serverKey,
	}
	expectSecret := &corev1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: K8SETCDResultSecretName,
		},
		Data: expectData,
	}
	_, err = m.secret.Create(context.TODO(), expectSecret, metav1.CreateOptions{})
	return err
}

func newCertTemplate() *x509.Certificate {
	max := new(big.Int).Lsh(big.NewInt(1), 128)   //把 1 左移 128 位，返回给 big.Int
	serialNumber, _ := rand.Int(rand.Reader, max) //返回在 [0, max) 区间均匀随机分布的一个随机值

	template := &x509.Certificate{
		SerialNumber:          serialNumber, // SerialNumber 是 CA 颁布的唯一序列号，在此使用一个大随机数来代表它
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  false, // 指示证书是不是ca证书
		BasicConstraintsValid: false, // 指示证书是不是ca证书
	}
	return template
}

// deleteSecretsTolerateNotFound 删除单个 Secret，忽略错误 Not Found
func (m *Manager) deleteSingleSecretTolerateNotFound(ctx context.Context, name string) error {
	m.logger.Infof("delete single secret/%v", name)
	if err := m.secret.Delete(ctx, name, metav1.DeleteOptions{}); apierrors.IsNotFound(err) {
		m.logger.Debugf("secret/%v not found", name)
	} else if err != nil {
		m.logger.Warningf("delete secret/%v fail: %v", name, err)
		return err
	}
	return nil
}
