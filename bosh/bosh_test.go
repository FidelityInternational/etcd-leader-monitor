package bosh_test

import (
	"crypto/tls"
	"fmt"
	"github.com/FidelityInternational/etcd-leader-monitor/bosh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"net/url"
)

var _ = Describe("#NewClient", func() {
	Context("When SkipSSLValidation is false", func() {
		It("returns a Bosh Client", func() {
			config := &bosh.Config{
				Username:          "example_user",
				Password:          "example_password",
				BoshURI:           "bosh_uri.example.com",
				Port:              "25555",
				SkipSSLValidation: false,
			}
			Expect(bosh.NewClient(config).Config).To(Equal(config))
		})
	})

	Context("When SkipSSLValidation is false", func() {
		It("returns a Bosh Client", func() {
			config := &bosh.Config{
				Username:          "example_user",
				Password:          "example_password",
				BoshURI:           "bosh_uri.example.com",
				Port:              "25555",
				SkipSSLValidation: true,
			}
			Expect(bosh.NewClient(config).Config).To(Equal(config))
		})
	})
})

var _ = Describe("#SearchDeployment", func() {
	var (
		client *bosh.Client
	)

	BeforeEach(func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `[
   {
      "name":"test-deployment-12345",
      "releases":[
         {
            "name":"example_release",
            "version":"2"
         }
      ],
      "stemcells":[
         {
            "name":"example_stemcell",
            "version":"1"
         }
      ]
   }
]`)
		}))

		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
			TLSClientConfig: &tls.Config{},
		}
		httpClient := &http.Client{Transport: transport}

		config := &bosh.Config{
			Username:          "example_user",
			Password:          "example_password",
			BoshURI:           "bosh_uri.example.com",
			Port:              "25555",
			HTTPClient:        httpClient,
			InsecureTransport: true,
		}
		client = bosh.NewClient(config)
	})

	Context("When the deployment can be found", func() {
		It("returns the deployment name as a string", func() {
			result, err := client.SearchDeployment("test-deployment.+")
			Expect(err).To(BeNil())
			Expect(result).To(Equal("test-deployment-12345"))
		})
	})

	Context("When the deployment cannot be found", func() {
		It("returns an error", func() {
			result, err := client.SearchDeployment("not_exist-.+")
			Expect(err).To(MatchError("No deployment was found"))
			Expect(result).To(Equal(""))
		})
	})
})

var _ = Describe("#GetEtcdVMs", func() {
	var (
		client     *bosh.Client
		httpClient *http.Client
	)

	JustBeforeEach(func() {
		config := &bosh.Config{
			Username:          "example_user",
			Password:          "example_password",
			BoshURI:           "bosh_uri.example.com",
			Port:              "25555",
			HTTPClient:        httpClient,
			InsecureTransport: true,
		}
		client = bosh.NewClient(config)
	})

	Context("when the VMs can be found", func() {
		BeforeEach(func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() == "http://bosh_uri.example.com:25555/deployments/test-deployment-12345/vms?format=full" {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Location", "http://bosh_uri.example.com:25555/tasks/1")
					w.WriteHeader(302)
					fmt.Fprintln(w, `{"id":1,"state":"queued","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`)
				} else if r.URL.String() == "http://bosh_uri.example.com:25555/tasks/1" {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Location", "http://bosh_uri.example.com:25555/tasks/1/output?type=result")
					w.WriteHeader(200)
					// state must be "done" to prevent infinite loop
					fmt.Fprintln(w, `{"id":1,"state":"done","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(200)
					fmt.Fprintln(w, `{"vm_cid":"1","ips":["1.1.1.1"],"agent_id":"1","job_name":"not_dea-partition-d284104a9345228c01e2","index":0}
          {"vm_cid":"2","ips":["2.2.2.2"],"agent_id":"2","job_name":"dea-partition-d284104a9345228c01e2","index":1}
          {"vm_cid":"3","ips":["3.3.3.3"],"agent_id":"3","job_name":"not_diego_cell-partition-d284104a9345228c01e2","index":0}
          {"vm_cid":"4","ips":["4.4.4.4"],"agent_id":"4","job_name":"diego_cell-partition-d284104a9345228c01e2","index":0}
          {"vm_cid":"5","ips":["5.5.5.5"],"agent_id":"5","job_name":"diego_cell-partition-d284104a9345228c01e2","index":1}
          {"vm_cid":"6","ips":["6.6.6.6"],"agent_id":"6","job_name":"dea-partition-d284104a9345228c01e2","index":2}
          {"vm_cid":"7","ips":["7.7.7.7"],"agent_id":"7","job_name":"dea-partition-d284104a9345228c01e2","index":3}
          {"vm_cid":"8","ips":["8.8.8.8"],"agent_id":"8","job_name":"dea-partition-d284104a9345228c01e2","index":4}
          {"vm_cid":"9","ips":["9.9.9.9"],"agent_id":"9","job_name":"dea-partition-d284104a9345228c01e2","index":5}
          {"vm_cid":"10","ips":["10.10.10.10"],"agent_id":"10","job_name":"dea-partition-d284104a9345228c01e2","index":6}
          {"vm_cid":"11","ips":["11.11.11.11"],"agent_id":"11","job_name":"dea-partition-d284104a9345228c01e2","index":0}
          {"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
          {"vm_cid":"2","ips":["31.31.31.31"],"agent_id":"2","job_name":"etcd_server-d284104a9345228c01e2","index":1}
          {"vm_cid":"6","ips":["32.32.32.32"],"agent_id":"6","job_name":"etcd_server-d284104a9345228c01e2","index":2}`)
				}
			}))

			transport := &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(server.URL)
				},
				TLSClientConfig: &tls.Config{},
			}
			httpClient = &http.Client{Transport: transport}
		})

		It("returns the Etcd VMs", func() {
			vms, err := client.GetEtcdVMs("test-deployment-12345")
			Expect(err).To(BeNil())
			Expect(vms).To(HaveLen(3))
			Expect(vms).To(ContainElement(&bosh.DeploymentVM{
				JobName: "etcd_server-d284104a9345228c01e2",
				Index:   0,
				VMCid:   "11",
				AgentID: "11",
				IPs:     []string{"30.30.30.30"},
			}))
			Expect(vms).To(ContainElement(&bosh.DeploymentVM{
				JobName: "etcd_server-d284104a9345228c01e2",
				Index:   1,
				VMCid:   "2",
				AgentID: "2",
				IPs:     []string{"31.31.31.31"},
			}))
			Expect(vms).To(ContainElement(&bosh.DeploymentVM{
				JobName: "etcd_server-d284104a9345228c01e2",
				Index:   2,
				VMCid:   "6",
				AgentID: "6",
				IPs:     []string{"32.32.32.32"},
			}))
		})
	})

	Context("when the VMs cannot be found", func() {
		BeforeEach(func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() == "http://bosh_uri.example.com:25555/deployments/test-deployment-12345/vms?format=full" {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintln(w, `{"id":1,"state":"queued","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`)
				} else if r.URL.String() == "http://bosh_uri.example.com:25555/tasks/1" {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					// state must be "done" to prevent infinite loop
					fmt.Fprintln(w, `{"id":1,"state":"done","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`)
				} else {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintln(w, `{"vm_cid":"1","ips":["1.1.1.1"],"agent_id":"1","job_name":"not_dea-partition-d284104a9345228c01e2","index":0}
          {"vm_cid":"3","ips":["3.3.3.3"],"agent_id":"3","job_name":"not_diego_cell-partition-d284104a9345228c01e2","index":0}`)
				}
			}))

			transport := &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(server.URL)
				},
				TLSClientConfig: &tls.Config{},
			}
			httpClient = &http.Client{Transport: transport}
		})
		It("returns an empty list", func() {
			vms, err := client.GetEtcdVMs("test-deployment-12345")
			Expect(err).To(BeNil())
			Expect(vms).To(HaveLen(0))
		})
	})
})

var _ = Describe("#GetAllIPs", func() {
	It("return IPs for the provided VMs", func() {
		var deploymentVMs = bosh.DeploymentVMs{
			{
				JobName: "dea-partition-d284104a9345228c01e2",
				Index:   0,
				VMCid:   "11",
				AgentID: "11",
				IPs:     []string{"11.11.11.11"},
			},
			{
				JobName: "dea-partition-d284104a9345228c01e2",
				Index:   1,
				VMCid:   "2",
				AgentID: "2",
				IPs:     []string{"2.2.2.2"},
			},
			{
				JobName: "dea-partition-d284104a9345228c01e2",
				Index:   2,
				VMCid:   "6",
				AgentID: "6",
				IPs:     []string{"6.6.6.6"},
			},
			{
				JobName: "dea-partition-d284104a9345228c01e2",
				Index:   3,
				VMCid:   "7",
				AgentID: "7",
				IPs:     []string{"7.7.7.7"},
			},
			{
				JobName: "dea-partition-d284104a9345228c01e2",
				Index:   4,
				VMCid:   "8",
				AgentID: "8",
				IPs:     []string{"8.8.8.8"},
			},
			{
				JobName: "dea-partition-d284104a9345228c01e2",
				Index:   5,
				VMCid:   "9",
				AgentID: "9",
				IPs:     []string{"9.9.9.9"},
			},
			{
				JobName: "dea-partition-d284104a9345228c01e2",
				Index:   6,
				VMCid:   "10",
				AgentID: "10",
				IPs:     []string{"10.10.10.10"},
			},
			{
				JobName: "diego_cell-partition-d284104a9345228c01e2",
				Index:   0,
				VMCid:   "4",
				AgentID: "4",
				IPs:     []string{"4.4.4.4"},
			},
			{
				JobName: "diego_cell-partition-d284104a9345228c01e2",
				Index:   1,
				VMCid:   "5",
				AgentID: "5",
				IPs:     []string{"5.5.5.5"},
			},
		}
		vmIPs := deploymentVMs.GetAllIPs()
		Expect(vmIPs).To(HaveLen(9))
		Expect(vmIPs).To(ContainElement("11.11.11.11"))
		Expect(vmIPs).To(ContainElement("2.2.2.2"))
		Expect(vmIPs).To(ContainElement("4.4.4.4"))
		Expect(vmIPs).To(ContainElement("5.5.5.5"))
		Expect(vmIPs).To(ContainElement("6.6.6.6"))
		Expect(vmIPs).To(ContainElement("7.7.7.7"))
		Expect(vmIPs).To(ContainElement("8.8.8.8"))
		Expect(vmIPs).To(ContainElement("9.9.9.9"))
		Expect(vmIPs).To(ContainElement("10.10.10.10"))
	})
})
