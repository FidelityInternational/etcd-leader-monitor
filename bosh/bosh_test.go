package bosh_test

import (
	"github.com/FidelityInternational/etcd-leader-monitor/bosh"
	. "github.com/cloudfoundry-community/gogobosh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#GetEtcdCerts", func() {
	var manifest string

	BeforeEach(func() {
		manifest = `---
jobs:
- name: test-job1
  properties:
    etcd:
      ca_cert : |
        -----BEGIN CERTIFICATE-----
        IAmAFakeCACert
        -----END CERTIFICATE-----
      client_cert: |
        -----BEGIN CERTIFICATE-----
        IAmAFakeClientCert
        -----END CERTIFICATE-----
      client_key: |
        -----BEGIN RSA PRIVATE KEY-----
        IAmAFakeClientKey
        -----END RSA PRIVATE KEY-----
`
	})

	Context("when the job does not match the regex", func() {
		It("returns an empty certs object", func() {
			certs := bosh.GetEtcdCerts(manifest, "not-matching")
			Ω(certs).Should(Equal(bosh.EtcdCerts{}))
		})

		Context("when the job does match the regex", func() {
			It("returns the certs object for the first regex matched job", func() {
				certs := bosh.GetEtcdCerts(manifest, "^test-job.*")
				Ω(certs.CaCert).Should(Equal(`-----BEGIN CERTIFICATE-----
IAmAFakeCACert
-----END CERTIFICATE-----
`))
				Ω(certs.ClientCert).Should(Equal(`-----BEGIN CERTIFICATE-----
IAmAFakeClientCert
-----END CERTIFICATE-----
`))
				Ω(certs.ClientKey).Should(Equal(`-----BEGIN RSA PRIVATE KEY-----
IAmAFakeClientKey
-----END RSA PRIVATE KEY-----
`))
			})
		})
	})
})

var _ = Describe("#FindDeployment", func() {
	var deployments []Deployment

	BeforeEach(func() {
		deployments = []Deployment{
			{
				Name:        "cf-warden-12345",
				CloudConfig: "none",
				Releases: []Resource{
					{
						Name:    "cf",
						Version: "223",
					},
				},
				Stemcells: []Resource{
					{
						Name:    "bosh-warden-boshlite-ubuntu-trusty-go_agent",
						Version: "3126",
					},
				},
			},
			{
				Name:        "cf-garden-12345",
				CloudConfig: "none",
				Releases: []Resource{
					{
						Name:    "cf",
						Version: "223",
					},
				},
				Stemcells: []Resource{
					{
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
