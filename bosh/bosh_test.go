package bosh_test

import (
	"github.com/FidelityInternational/etcd-leader-monitor/bosh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/srbry/gogobosh"
)

var _ = Describe("#FindDeployment", func() {
	var deployments []Deployment

	BeforeEach(func() {
		deployments = []Deployment{
			{
				Name:        "cf-warden-12345",
				CloudConfig: "none",
				Releases: []Resource{
					Resource{
						Name:    "cf",
						Version: "223",
					},
				},
				Stemcells: []Resource{
					Resource{
						Name:    "bosh-warden-boshlite-ubuntu-trusty-go_agent",
						Version: "3126",
					},
				},
			},
			{
				Name:        "cf-garden-12345",
				CloudConfig: "none",
				Releases: []Resource{
					Resource{
						Name:    "cf",
						Version: "223",
					},
				},
				Stemcells: []Resource{
					Resource{
						Name:    "bosh-warden-boshlite-ubuntu-trusty-go_agent",
						Version: "3126",
					},
				},
			},
		}
	})

	Context("when a deployment can be found", func() {
		It("finds the first matching deployment name based on a regex", func() {
			Ω(bosh.FindDeployment(deployments, "cf-garden*")).Should(Equal("cf-garden-12345"))
		})
	})

	Context("when a deployment cannot be found", func() {
		It("returns an empty string", func() {
			Ω(bosh.FindDeployment(deployments, "bosh*")).Should(BeEmpty())
		})
	})
})

var _ = Describe("#FindVMs", func() {
	It("Returns an array of all VMs matching the given regex", func() {
		vms := []VM{
			{
				IPs:     []string{"1.1.1.1"},
				JobName: "etcd_server-12344",
			},
			{
				IPs:     []string{"4.4.4.4"},
				JobName: "consul_server-567887",
			},
			{
				IPs:     []string{"3.3.3.3"},
				JobName: "etcd_server-98764",
			},
			{
				IPs:     []string{"4.4.4.4"},
				JobName: "consul_server-12344",
			},
			{
				IPs:     []string{"5.5.5.5"},
				JobName: "etcd_server-567887",
			},
		}
		matchedVMs := bosh.FindVMs(vms, "^etcd_server.+$")
		Ω(matchedVMs).Should(HaveLen(3))
		Ω(matchedVMs).Should(ContainElement(VM{
			IPs:     []string{"1.1.1.1"},
			JobName: "etcd_server-12344",
		}))
		Ω(matchedVMs).Should(ContainElement(VM{
			IPs:     []string{"3.3.3.3"},
			JobName: "etcd_server-98764",
		}))
		Ω(matchedVMs).Should(ContainElement(VM{
			IPs:     []string{"5.5.5.5"},
			JobName: "etcd_server-567887",
		}))
	})
})
