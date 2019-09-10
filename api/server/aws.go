package server

import (
	"fmt"
	"gopkg.in/yaml.v2"

	"github.com/kubernauts/tk8/api"
	"github.com/kubernauts/tk8/pkg/common"
	"github.com/sirupsen/logrus"
)

type AwsYaml struct {
	Aws *Aws `yaml:"aws"`
}
type Aws struct {
	Clustername                  string `yaml:"clustername" json:"clustername"`
	Os                           string `yaml:"os" json:"os"`
	AwsAccessKeyID               string `yaml:"aws_access_key_id" json:"aws_access_key_id"`
	AwsSecretAccessKey           string `yaml:"aws_secret_access_key" json:"aws_secret_access_key"`
	AwsSSHKeypair                string `yaml:"aws_ssh_keypair" json:"aws_ssh_keypair"`
	AwsDefaultRegion             string `yaml:"aws_default_region" json:"aws_default_region"`
	AwsVpcCidrBlock              string `yaml:"aws_vpc_cidr_block" json:"aws_vpc_cidr_block"`
	AwsCidrSubnetsPrivate        string `yaml:"aws_cidr_subnets_private" json:"aws_cidr_subnets_private"`
	AwsCidrSubnetsPublic         string `yaml:"aws_cidr_subnets_public" json:"aws_cidr_subnets_public"`
	AwsBastionSize               string `yaml:"aws_bastion_size" json:"aws_bastion_size"`
	AwsKubeMasterNum             int    `yaml:"aws_kube_master_num" json:"aws_kube_master_num"`
	AwsKubeMasterSize            string `yaml:"aws_kube_master_size" json:"aws_kube_master_size"`
	AwsEtcdNum                   int    `yaml:"aws_etcd_num" json:"aws_etcd_num"`
	AwsEtcdSize                  string `yaml:"aws_etcd_size"  json:"aws_etcd_size"`
	AwsKubeWorkerNum             int    `yaml:"aws_kube_worker_num"  json:"aws_kube_worker_num"`
	AwsKubeWorkerSize            string `yaml:"aws_kube_worker_size"  json:"aws_kube_worker_size"`
	AwsElbAPIPort                int    `yaml:"aws_elb_api_port"  json:"aws_elb_api_port"`
	K8SSecureAPIPort             int    `yaml:"k8s_secure_api_port"  json:"k8s_secure_api_port"`
	KubeInsecureApiserverAddress string `yaml:"kube_insecure_apiserver_address"  json:"kube_insecure_apiserver_address"`
	KubeadmEnabled               bool   `yaml:"kubeadm_enabled"  json:"kubeadm_enabled"`
	KubeNetworkPlugin            string `yaml:"kube_network_plugin"  json:"kube_network_plugin"`
}

// CreateCluster creates AWS cluster
func (a *Aws) CreateCluster() error {

	// create AWS cluster config file
	configFileName := "aws-" + a.Clustername + ".yaml"
	s := NewStore(common.REST_API_STORAGE, configFileName, common.REST_API_STORAGEPATH, common.REST_API_STORAGEREGION)

	provisioner := "aws"
	// validateJSON
	err := s.ValidateConfig()
	if err != nil {
		logrus.Errorf("Error validating config ::: %s", err)
		return err
	}

	err = getProvisioner(provisioner)
	if err != nil {
		logrus.Errorf("Error getting provisioner ::: %s", err)
		return err
	}
	err = s.CreateConfig(a)
	if err != nil {
		logrus.Errorf("Error creating config ::: %s", err)
		return err
	}
	go func() {
		Provisioners[provisioner].Init(nil)
		Provisioners[provisioner].Setup(nil)
	}()
	return nil
}

// DestroyCluster destroys AWS cluster
func (a *Aws) DestroyCluster() error {
	configFileName := "aws-" + a.Clustername + ".yaml"
	s := NewStore(common.REST_API_STORAGE, configFileName, common.REST_API_STORAGEPATH, common.REST_API_STORAGEREGION)

	exists, _ := s.CheckConfigExists()
	if !exists {
		logrus.Errorf("Error , no such cluster with name %s", a.Clustername)
		return fmt.Errorf("No such cluster exists with name - %s", a.Clustername)
	}

	go func() {
		Provisioners["aws"].Destroy(nil)
	}()

	// Delete AWS cluster config file
	err := s.DeleteConfig()
	if err != nil {
		logrus.Errorf("Error , deleting cluster named  %s , Error details are %s ", a.Clustername, err.Error())
		return fmt.Errorf("Error deleting cluster %s", a.Clustername)
	}

	return nil
}

// GetCluster gets the details of thge requested AWS cluster
func (a *Aws) GetCluster(name string) (api.Cluster, error) {

	configFileName := "aws-" + name + ".yaml"
	s := NewStore(common.REST_API_STORAGE, configFileName, common.REST_API_STORAGEPATH, common.REST_API_STORAGEREGION)
	exists, _ := s.CheckConfigExists()

	if !exists {
		return nil, fmt.Errorf("No cluster found with the provided name ::: %s", name)
	}

	awsConfig := &AwsYaml{}
	yamlFile, err := s.GetConfig()
	if err != nil {
		logrus.Errorf("Error getting details of cluster named %s , Error details are %s ", name, err.Error())
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, awsConfig)
	if err != nil {
		logrus.Errorf("unable to decode into rke config struct, %v", err)
		return nil, err
	}
	return awsConfig.Aws, nil

}