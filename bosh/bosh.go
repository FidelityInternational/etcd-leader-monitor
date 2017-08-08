package bosh

import (
	"fmt"
	"regexp"

	"github.com/cloudfoundry-community/gogobosh"
	"gopkg.in/yaml.v2"
)

type manifest struct {
	Jobs           []jobs           `yaml:"jobs"`
	InstanceGroups []instanceGroups `yaml:"instance_groups"`
}

type instanceGroups struct {
	Name       string                  `yaml:"name"`
	Jobs       []jobs                  `yaml:"jobs"`
	Properties diegoDatabaseProperties `yaml:"properties"`
}

type jobs struct {
	Name       string                  `yaml:"name"`
	Properties diegoDatabaseProperties `yaml:"properties"`
}

type diegoDatabaseProperties struct {
	Etcd EtcdCerts `yaml:"etcd"`
}

// EtcdCerts - A struct that defines the required certs for SSL secured etcd
type EtcdCerts struct {
	ClientKey  string `yaml:"client_key"`
	ClientCert string `yaml:"client_cert"`
	CaCert     string `yaml:"ca_cert"`
}

// GetEtcdCerts - Returns Client Key/Cert and CaCert that could be used for SSL secured etcd
func GetEtcdCerts(deploymentManifest string, configBlockRegex string) (EtcdCerts, error) {
	var deployManifest manifest
	if err := yaml.Unmarshal([]byte(deploymentManifest), &deployManifest); err != nil {
		return EtcdCerts{}, err
	}

	instanceGroups := deployManifest.InstanceGroups

	for _, instanceGroup := range instanceGroups {
		matched, _ := regexp.MatchString(configBlockRegex, instanceGroup.Name)
		if matched {
			for _, job := range instanceGroup.Jobs {
				fmt.Printf("matched. Using manifest: %v", deploymentManifest)
				if job.Properties.Etcd != (EtcdCerts{}) {
					fmt.Println("Using job properties")
					return job.Properties.Etcd, nil
				}
			}
			if instanceGroup.Properties.Etcd != (EtcdCerts{}) {
				fmt.Println("Using instanceGroup Job properties")
				return instanceGroup.Properties.Etcd, nil
			}
		}
	}

	for _, job := range deployManifest.Jobs {
		matched, _ := regexp.MatchString(configBlockRegex, job.Name)
		if matched {
			return job.Properties.Etcd, nil
		}
	}
	return EtcdCerts{}, nil
}

// FindDeployment - takes deployments and a regex to return the first matching deployment name
func FindDeployment(deployments []gogobosh.Deployment, regex string) string {
	for _, deployment := range deployments {
		matched, _ := regexp.MatchString(regex, deployment.Name)
		if matched {
			return deployment.Name
		}
	}
	return ""
}

// FindVMs - takes an array of VMs and a regex to filter on, returning a new array of all matching vms
func FindVMs(deploymentVMs []gogobosh.VM, regex string) []gogobosh.VM {
	var matchedVMs []gogobosh.VM
	for _, deploymentVM := range deploymentVMs {
		matched, _ := regexp.MatchString(regex, deploymentVM.JobName)
		if matched {
			matchedVMs = append(matchedVMs, deploymentVM)
		}
	}
	return matchedVMs
}
