package bosh_test

import (
	"github.com/FidelityInternational/etcd-leader-monitor/bosh"
	. "github.com/cloudfoundry-community/gogobosh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#GetEtcdCerts", func() {
	var (
		manifest     string
		certs        bosh.EtcdCerts
		err          error
		jobNameRegex string
	)

	JustBeforeEach(func() {
		certs, err = bosh.GetEtcdCerts(manifest, jobNameRegex)
	})

	Context("when unmarshalling a bosh response from yaml returns an error", func() {
		BeforeEach(func() {
			manifest = `---
jobs:sa
\:''
name: test-job1
  properties:d:
      ca_c|
        -----BEGIN RSA PRIVATE KEY-----
        IAmAFakeClientKey
        -----END RSA PRIVATE KEY-----
`
		})
		It("returns the error", func() {
			Ω(certs).Should(Equal(bosh.EtcdCerts{}))
			Ω(err).Should(MatchError("yaml: line 3: mapping values are not allowed in this context"))
		})
	})

	Context("when the bosh manifest format uses 'jobs'", func() {
		BeforeEach(func() {
			manifest = `---
jobs:
- name: test-job1
  properties:
    etcd:
      ca_cert: |
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
			BeforeEach(func() {
				jobNameRegex = "not-matching"
			})

			It("returns an empty certs object", func() {
				Ω(certs).Should(Equal(bosh.EtcdCerts{}))
				Ω(err).Should(BeNil())
			})

			Context("when the job does match the regex", func() {
				BeforeEach(func() {
					jobNameRegex = "^test-job.*"
				})

				It("returns the certs object for the first regex matched job", func() {
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
					Ω(err).Should(BeNil())
				})
			})
		})
	})

	Context("when the bosh manifest format uses 'instance_groups'", func() {
		BeforeEach(func() {
			manifest = `---
instance_groups:
- name: test-job1
  properties:
    etcd:
      ca_cert: |
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
			BeforeEach(func() {
				jobNameRegex = "not-matching"
			})

			It("returns an empty certs object", func() {
				Ω(certs).Should(Equal(bosh.EtcdCerts{}))
				Ω(err).Should(BeNil())
			})

			Context("when the job does match the regex and does not have certs properties", func() {
				BeforeEach(func() {
					jobNameRegex = "^test-job1.*"
				})

				It("returns the certs object of the instance group", func() {
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
					Ω(err).Should(BeNil())
				})
			})

			Context("when the job does match the regex and does have certs properties", func() {
				BeforeEach(func() {
					jobNameRegex = "^test-job1.*"
					manifest = `---
    instance_groups:
    - name: test-job1
      jobs:
      - name: etcd
        properties:
          etcd:
            ca_cert: |
              -----BEGIN CERTIFICATE-----
              IAmAFakeCACertInAJob
              -----END CERTIFICATE-----
            client_cert: |
              -----BEGIN CERTIFICATE-----
              IAmAFakeClientCertInAJob
              -----END CERTIFICATE-----
            client_key: |
              -----BEGIN RSA PRIVATE KEY-----
              IAmAFakeClientKeyInAJob
              -----END RSA PRIVATE KEY-----
      properties:
        etcd:
          ca_cert: |
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
				It("returns the certs object of the job", func() {
					Ω(certs.CaCert).Should(Equal(`-----BEGIN CERTIFICATE-----
IAmAFakeCACertInAJob
-----END CERTIFICATE-----
`))
					Ω(certs.ClientCert).Should(Equal(`-----BEGIN CERTIFICATE-----
IAmAFakeClientCertInAJob
-----END CERTIFICATE-----
`))
					Ω(certs.ClientKey).Should(Equal(`-----BEGIN RSA PRIVATE KEY-----
IAmAFakeClientKeyInAJob
-----END RSA PRIVATE KEY-----
`))
					Ω(err).Should(BeNil())
				})
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
