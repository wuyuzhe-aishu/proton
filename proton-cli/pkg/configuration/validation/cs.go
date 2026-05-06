package validation

import (
	"github.com/go-test/deep"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

var SupportedKubernetesProvisioner = sets.NewString(
	string(configuration.KubernetesProvisionerLocal),
	string(configuration.KubernetesProvisionerExternal),
)

func ValidateCS(c *configuration.Cs, nodes []configuration.Node, fldPath *field.Path) (allErrs field.ErrorList) {
	var nodeNameSet = make(sets.Set[string])
	for _, n := range nodes {
		nodeNameSet.Insert(n.Name)
	}
	allErrs = append(allErrs, ValidateHosts(c.Master, nodeNameSet, fldPath.Child("hosts"))...)
	if !SupportedKubernetesProvisioner.Has(string(c.Provisioner)) {
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("provisioner"), c.Provisioner, SupportedKubernetesProvisioner.List()))
	}

	if c.Provisioner == configuration.KubernetesProvisionerLocal {
		allErrs = append(allErrs, ValidateCS_IPFamilies(c.IPFamilies, fldPath.Child("ipFamilies"))...)
	} else if c.IPFamilies != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("ipFamilies"), "provisioner isn't local"))
	}
	// TODO: Validate container runtime, if provisioner is local

	// 验证 Proton CS 所使用的节点是否存在指定 ip family 的地址
	for _, ipFamily := range c.IPFamilies {
		for i, node := range nodes {
			allErrs = append(allErrs, ValidateCS_NodeIPFamily(&node, ipFamily, field.NewPath("nodes").Index(i))...)
		}
	}
	allErrs = append(allErrs, ValidateCSAddons(c.Addons, fldPath.Child("addons"))...)
	allErrs = append(allErrs, ValidateCSAddonsConfig(c.AddonsConfig, fldPath.Child("addonsConfig"))...)
	allErrs = append(allErrs, ValidateCS_DualStack(c.IPFamilies, c.EnableDualStack, fldPath.Child("ipFamilies"))...)

	return
}

func ValidateCSAddonsConfig(config *configuration.CSAddonsConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	if config == nil || config.IngressNginx == nil {
		return nil
	}

	allErrs = append(allErrs, validateTCPPort(config.IngressNginx.HTTPPort, fldPath.Child("ingress-nginx").Child("httpPort"))...)
	allErrs = append(allErrs, validateTCPPort(config.IngressNginx.HTTPSPort, fldPath.Child("ingress-nginx").Child("httpsPort"))...)
	return
}

func validateTCPPort(port int, fldPath *field.Path) (allErrs field.ErrorList) {
	if port == 0 {
		return nil
	}
	if port < 1 || port > 65535 {
		allErrs = append(allErrs, field.Invalid(fldPath, port, "must be between 1 and 65535"))
	}
	return
}

func ValidateCSAddons(addons []configuration.CSAddonName, fldPath *field.Path) (allErrs field.ErrorList) {
	supportedAddons := sets.NewString(
		string(configuration.CSAddonNameNodeExporter),
		string(configuration.CSAddonNameStateMetrics),
		string(configuration.CSAddonNameIngressNginx),
	)
	allAddons := sets.NewString()
	for i, a := range addons {
		if !supportedAddons.Has(string(a)) {
			allErrs = append(allErrs, field.NotSupported(fldPath.Index(i), a, supportedAddons.List()))
			continue
		}
		if allAddons.Has(string(a)) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Index(i), a))
			continue
		}
		allAddons.Insert(string(a))
	}
	return
}

// ValidateCSUpdate 验证 Proton CS 配置更新是否合法
//
//   - ipFamilies 不支持修改
func ValidateCSUpdate(o, n *configuration.Cs, fldPath *field.Path) (allErrs field.ErrorList) {
	for _, a := range o.Addons {
		if !slices.Contains(n.Addons, a) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("addons"), a, "proton cs addons do not support uninstall"))
		}
	}
	if o.IPFamilies != nil {
		if deep.Equal(o.IPFamilies, n.IPFamilies) != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("ipFamilies"), "should not be changed"))
		}
	}
	if o.EnableDualStack != n.EnableDualStack {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("enableDualStack"), "should not be changed"))
	}
	return
}

const DetailProtonCSRequires = "proton-cs requires"

var supportedServiceIPFamily = sets.NewString(string(v1.IPv4Protocol), string(v1.IPv6Protocol))

// ValidateCS_IPFamilies 校验 ip families 的值是否合法
//
//   - 仅支持 IPv4 或 IPv6
//   - 不允许重复
//   - 至少设置一种
func ValidateCS_IPFamilies(ipFamilies []v1.IPFamily, fldPath *field.Path) (allErrs field.ErrorList) {
	if len(ipFamilies) == 0 {
		allErrs = append(allErrs, field.Required(fldPath, ""))
	}
	// ipFamilies stand alone validation must be either IPv4 or IPv6
	seen := make(sets.Set[string])
	for i, ipFamily := range ipFamilies {
		if !supportedServiceIPFamily.Has(string(ipFamily)) {
			allErrs = append(allErrs, field.NotSupported(fldPath.Index(i), ipFamily, supportedServiceIPFamily.List()))
			continue
		}
		// no duplicate check also ensures that ipFamilies is dual-stacked, in any order
		if seen.Has(string(ipFamily)) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Index(i), ipFamily))
			continue
		}
		seen.Insert(string(ipFamily))
	}
	return
}

// ValidateCS_NodeIPFamily 验证 Proton CS 所使用的节点是否存在指定 ip family 的
// 地址
func ValidateCS_NodeIPFamily(node *configuration.Node, ipFamily v1.IPFamily, fldPath *field.Path) (allErrs field.ErrorList) {
	switch ipFamily {
	case v1.IPv4Protocol:
		if node.IP4 == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("ip4"), DetailProtonCSRequires))
		}
	case v1.IPv6Protocol:
		if node.IP6 == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("ip6"), DetailProtonCSRequires))
		}
	}
	return
}

func ValidateCS_DualStack(ipFamilies []v1.IPFamily, enableDualStack bool, fldPath *field.Path) (allErrs field.ErrorList) {
	if enableDualStack && len(ipFamilies) != 2 {
		allErrs = append(allErrs, field.Invalid(fldPath, ipFamilies, "enableDualStack requires exactly two ip families"))
	}
	return
}
