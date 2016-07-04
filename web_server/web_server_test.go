package webServer_test

import (
	"crypto/tls"
	"fmt"
	"github.com/FidelityInternational/etcd-leader-monitor/bosh"
	webs "github.com/FidelityInternational/etcd-leader-monitor/web_server"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func Router(controller *webs.Controller) *mux.Router {
	server := &webs.Server{Controller: controller}
	r := server.Start()
	return r
}

func init() {
	var controller *webs.Controller
	http.Handle("/", Router(controller))
}

var _ = Describe("Server", func() {
	Describe("#CreateServer", func() {
		It("returns a server object", func() {
			Ω(webs.CreateServer(&bosh.Client{}, &http.Client{})).Should(BeAssignableToTypeOf(&webs.Server{}))
		})
	})
})

var _ = Describe("Contoller", func() {
	Describe("#CreateController", func() {
		It("returns a controller object", func() {
			controller := webs.CreateController(&bosh.Client{}, &http.Client{})
			Ω(controller).Should(BeAssignableToTypeOf(&webs.Controller{}))
		})
	})

	Describe("#CheckLeaders", func() {
		var (
			controller   *webs.Controller
			req          *http.Request
			mockRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			req, _ = http.NewRequest("GET", "http://example.com/", nil)
			Router(controller).ServeHTTP(mockRecorder, req)
		})

		Context("when a bosh deployment cannot be found", func() {
			BeforeEach(func() {
				boshConfig := &bosh.Config{
					Username:          "example_user",
					Password:          "example_password",
					BoshURI:           "bosh_uri.example.com",
					Port:              "25555",
					SkipSSLValidation: true,
				}
				boshClient := bosh.NewClient(boshConfig)
				controller = webs.CreateController(boshClient, &http.Client{})
				mockRecorder = httptest.NewRecorder()
			})

			It("returns an error 500", func() {
				Ω(mockRecorder.Code).Should(Equal(500))
			})
		})

		Context("when bosh.GetEtcdVMs returns an error", func() {
			BeforeEach(func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.String() == "http://bosh_uri.example.com:25555/deployments" {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `[
   {
      "name":"cf-12345",
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
					} else if r.URL.String() == "http://bosh_uri.example.com:25555/deployments/cf-12345/vms?format=full" {
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
						fmt.Fprintln(w, `":["30.30.30_id""11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
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
				httpClient := &http.Client{Transport: transport}

				boshConfig := &bosh.Config{
					Username:          "example_user",
					Password:          "example_password",
					BoshURI:           "bosh_uri.example.com",
					Port:              "25555",
					HTTPClient:        httpClient,
					InsecureTransport: true,
				}
				boshClient := bosh.NewClient(boshConfig)
				controller = webs.CreateController(boshClient, &http.Client{})
				mockRecorder = httptest.NewRecorder()
			})

			It("returns an error 500", func() {
				Ω(mockRecorder.Code).Should(Equal(500))
			})
		})

		Context("when fetching leader stats returns an error", func() {
			BeforeEach(func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.String() == "http://bosh_uri.example.com:25555/deployments" {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `[
		   {
		      "name":"cf-12345",
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
					} else if r.URL.String() == "http://bosh_uri.example.com:25555/deployments/cf-12345/vms?format=full" {
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
						fmt.Fprintln(w, `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
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
				httpClient := &http.Client{Transport: transport}

				boshConfig := &bosh.Config{
					Username:          "example_user",
					Password:          "example_password",
					BoshURI:           "bosh_uri.example.com",
					Port:              "25555",
					HTTPClient:        httpClient,
					InsecureTransport: true,
				}
				boshClient := bosh.NewClient(boshConfig)

				etcdServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintln(w, `lowers"{"a0294459200078aa":{"latency":{"current":0.001199,"average":0.0023682517720168754,"standardDeviation":0.4302199179552562,"minimum":0.000654,"maximum":1996.564157},"counts":{"fail":16,"success":21538911}},"b5c352b4495e4195":{"latency":{"current":0.001609,"average":0.002361467019756358,"standardDeviation":0.00506414137059054,"minimum":0.00088,"maximum":5.153269},"counts":{"fail":7,"success":1617908}}}}`)
				}))

				etcdTransport := &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(etcdServer.URL)
					},
					TLSClientConfig: &tls.Config{},
				}
				etcdHttpClient := &http.Client{Transport: etcdTransport}

				controller = webs.CreateController(boshClient, etcdHttpClient)
				mockRecorder = httptest.NewRecorder()
			})

			It("returns an error 500", func() {
				Ω(mockRecorder.Code).Should(Equal(500))
			})
		})

		Context("when the number of followers is incorrect", func() {
			BeforeEach(func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.String() == "http://bosh_uri.example.com:25555/deployments" {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `[
		   {
		      "name":"cf-12345",
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
					} else if r.URL.String() == "http://bosh_uri.example.com:25555/deployments/cf-12345/vms?format=full" {
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
						fmt.Fprintln(w, `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
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
				httpClient := &http.Client{Transport: transport}

				boshConfig := &bosh.Config{
					Username:          "example_user",
					Password:          "example_password",
					BoshURI:           "bosh_uri.example.com",
					Port:              "25555",
					HTTPClient:        httpClient,
					InsecureTransport: true,
				}
				boshClient := bosh.NewClient(boshConfig)

				etcdServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.String() == "http://30.30.30.30:4001/v2/stats/leader" {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `{"leader":"6a0b69a54415a491","followers":{"b5c352b4495e4195":{"latency":{"current":0.001609,"average":0.002361467019756358,"standardDeviation":0.00506414137059054,"minimum":0.00088,"maximum":5.153269},"counts":{"fail":7,"success":1617908}}}}`)
					} else {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `{"message":"not current leader"}`)
					}
				}))

				etcdTransport := &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(etcdServer.URL)
					},
					TLSClientConfig: &tls.Config{},
				}
				etcdHttpClient := &http.Client{Transport: etcdTransport}

				controller = webs.CreateController(boshClient, etcdHttpClient)
				mockRecorder = httptest.NewRecorder()
			})

			It("returns a suitable json response", func() {
				Ω(mockRecorder.Code).Should(Equal(200))
				Expect(mockRecorder.Body.String()).Should(Equal(`{"healthy": false, "message": "Incorrect number of followers"}`))
			})
		})

		Context("when more than one etcd thinks it is the leader", func() {
			BeforeEach(func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.String() == "http://bosh_uri.example.com:25555/deployments" {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `[
		   {
		      "name":"cf-12345",
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
					} else if r.URL.String() == "http://bosh_uri.example.com:25555/deployments/cf-12345/vms?format=full" {
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
						fmt.Fprintln(w, `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
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
				httpClient := &http.Client{Transport: transport}

				boshConfig := &bosh.Config{
					Username:          "example_user",
					Password:          "example_password",
					BoshURI:           "bosh_uri.example.com",
					Port:              "25555",
					HTTPClient:        httpClient,
					InsecureTransport: true,
				}
				boshClient := bosh.NewClient(boshConfig)

				etcdServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintln(w, `{"leader":"6a0b69a54415a491","followers":{"a0294459200078aa":{"latency":{"current":0.001199,"average":0.0023682517720168754,"standardDeviation":0.4302199179552562,"minimum":0.000654,"maximum":1996.564157},"counts":{"fail":16,"success":21538911}},"b5c352b4495e4195":{"latency":{"current":0.001609,"average":0.002361467019756358,"standardDeviation":0.00506414137059054,"minimum":0.00088,"maximum":5.153269},"counts":{"fail":7,"success":1617908}}}}`)
				}))

				etcdTransport := &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(etcdServer.URL)
					},
					TLSClientConfig: &tls.Config{},
				}
				etcdHttpClient := &http.Client{Transport: etcdTransport}

				controller = webs.CreateController(boshClient, etcdHttpClient)
				mockRecorder = httptest.NewRecorder()
			})

			It("returns a suitable json response", func() {
				Ω(mockRecorder.Code).Should(Equal(200))
				Expect(mockRecorder.Body.String()).Should(Equal(`{"healthy": false, "message": "Too many leaders"}`))
			})
		})

		Context("Not enough etcds are leaders", func() {
			BeforeEach(func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.String() == "http://bosh_uri.example.com:25555/deployments" {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `[
		   {
		      "name":"cf-12345",
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
					} else if r.URL.String() == "http://bosh_uri.example.com:25555/deployments/cf-12345/vms?format=full" {
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
						fmt.Fprintln(w, `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
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
				httpClient := &http.Client{Transport: transport}

				boshConfig := &bosh.Config{
					Username:          "example_user",
					Password:          "example_password",
					BoshURI:           "bosh_uri.example.com",
					Port:              "25555",
					HTTPClient:        httpClient,
					InsecureTransport: true,
				}
				boshClient := bosh.NewClient(boshConfig)

				etcdServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintln(w, `{"message":"not current leader"}`)
				}))

				etcdTransport := &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(etcdServer.URL)
					},
					TLSClientConfig: &tls.Config{},
				}
				etcdHttpClient := &http.Client{Transport: etcdTransport}

				controller = webs.CreateController(boshClient, etcdHttpClient)
				mockRecorder = httptest.NewRecorder()
			})

			It("returns a suitable json response", func() {
				Ω(mockRecorder.Code).Should(Equal(200))
				Expect(mockRecorder.Body.String()).Should(Equal(`{"healthy": false, "message": "Not enough leaders"}`))
			})
		})

		Context("When etcds are healthy and clustered correctly", func() {
			BeforeEach(func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.String() == "http://bosh_uri.example.com:25555/deployments" {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `[
		   {
		      "name":"cf-12345",
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
					} else if r.URL.String() == "http://bosh_uri.example.com:25555/deployments/cf-12345/vms?format=full" {
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
						fmt.Fprintln(w, `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
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
				httpClient := &http.Client{Transport: transport}

				boshConfig := &bosh.Config{
					Username:          "example_user",
					Password:          "example_password",
					BoshURI:           "bosh_uri.example.com",
					Port:              "25555",
					HTTPClient:        httpClient,
					InsecureTransport: true,
				}
				boshClient := bosh.NewClient(boshConfig)

				etcdServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.String() == "http://30.30.30.30:4001/v2/stats/leader" {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `{"leader":"6a0b69a54415a491","followers":{"a0294459200078aa":{"latency":{"current":0.001199,"average":0.0023682517720168754,"standardDeviation":0.4302199179552562,"minimum":0.000654,"maximum":1996.564157},"counts":{"fail":16,"success":21538911}},"b5c352b4495e4195":{"latency":{"current":0.001609,"average":0.002361467019756358,"standardDeviation":0.00506414137059054,"minimum":0.00088,"maximum":5.153269},"counts":{"fail":7,"success":1617908}}}}`)
					} else {
						w.WriteHeader(200)
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprintln(w, `{"message":"not current leader"}`)
					}
				}))

				etcdTransport := &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(etcdServer.URL)
					},
					TLSClientConfig: &tls.Config{},
				}
				etcdHttpClient := &http.Client{Transport: etcdTransport}

				controller = webs.CreateController(boshClient, etcdHttpClient)
				mockRecorder = httptest.NewRecorder()
			})

			It("returns a suitable json response", func() {
				Ω(mockRecorder.Code).Should(Equal(200))
				Expect(mockRecorder.Body.String()).Should(Equal(`{"healthy": true, "message": "Everything is healthy"}`))
			})
		})
	})
})
