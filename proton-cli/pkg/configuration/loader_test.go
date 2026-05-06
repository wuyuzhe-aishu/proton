package configuration

import (
	"testing"

	"github.com/go-test/deep"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestLoadFromFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *ClusterConfig
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: "testdata/valid-cluster-config.yaml",
			},
			want: &ClusterConfig{
				ApiVersion: "v1",
				Nodes: []Node{
					{
						Name: "vm-14-71-centos",
						IP4:  "10.4.14.71",
					},
				},
				Cs: &Cs{
					Master: []string{
						"vm-14-71-centos",
					},
					Addons: []CSAddonName{},
					Host_network: &HostNetWork{
						Bip:              "172.33.0.1/16",
						Pod_network_cidr: "192.169.0.0/16",
						Service_cidr:     "10.96.0.0/12",
					},
					Ha_port:           8443,
					Etcd_data_dir:     "/sysvol/proton_data/cs_etcd_data",
					Cs_controller_dir: "./service-package",
				},
				Cr: &Cr{
					Local: &LocalCR{
						Hosts: []string{
							"vm-14-71-centos",
						},
						Ports: Ports{
							Chartmuseum: 5001,
							Registry:    5000,
							Rpm:         5003,
							Cr_manager:  5002,
						},
						Ha_ports: Ports{
							Chartmuseum: 15001,
							Registry:    15000,
							Rpm:         15003,
							Cr_manager:  15002,
						},
						Storage: "/sysvol/proton_data/cr_data",
					},
				},
				CMS: &CMS{},
				Kafka: &Kafka{
					Hosts: []string{
						"vm-14-71-centos",
					},
					Data_path: "/sysvol/kafka",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("10m"),
							v1.ResourceMemory: resource.MustParse("16Mi"),
						},
					},
				},
				ZooKeeper: &ZooKeeper{
					Hosts: []string{
						"vm-14-71-centos",
					},
					Data_path: "/sysvol/zookeeper",
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("10m"),
							v1.ResourceMemory: resource.MustParse("16Mi"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadFromFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("len(Cs.Addons) = %v, Cs.Addons is nil = %v", len(got.Cs.Addons), got.Cs.Addons == nil)
			for _, diff := range deep.Equal(got, tt.want) {
				t.Errorf("LoadFromFile() difference between got and want: %v", diff)
			}
		})
	}
}

func TestLoadFromBytesIgnoresLegacyDeployModeAndDevicespec(t *testing.T) {
	data := []byte(`
apiVersion: v1
deploy:
  mode: standard
  devicespec: AS10000
  namespace: resource
  serviceaccount: proton-sa
`)

	got, err := LoadFromBytes(data)
	if err != nil {
		t.Fatalf("LoadFromBytes() error = %v", err)
	}
	if got.Deploy == nil {
		t.Fatalf("expected deploy to be loaded")
	}
	if got.Deploy.Namespace != "resource" {
		t.Fatalf("expected namespace resource, got %q", got.Deploy.Namespace)
	}
	if got.Deploy.ServiceAccount != "proton-sa" {
		t.Fatalf("expected serviceaccount proton-sa, got %q", got.Deploy.ServiceAccount)
	}
}
