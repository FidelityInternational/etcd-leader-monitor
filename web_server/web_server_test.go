package webServer_test

import (
	"crypto/tls"
	"fmt"
	webs "github.com/FidelityInternational/etcd-leader-monitor/web_server"
	"github.com/cloudfoundry-community/gogobosh"
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
			Ω(webs.CreateServer(&gogobosh.Client{}, &http.Client{})).Should(BeAssignableToTypeOf(&webs.Server{}))
		})
	})
})

var _ = Describe("Contoller", func() {
	Describe("#CreateController", func() {
		It("returns a controller object", func() {
			controller := webs.CreateController(&gogobosh.Client{}, &http.Client{})
			Ω(controller).Should(BeAssignableToTypeOf(&webs.Controller{}))
		})
	})

	Describe("#LoadCerts", func() {
		var (
			c              *webs.Controller
			deployConfig   webs.Config
			deploymentName = "deployment-test"
		)

		AfterEach(func() {
			teardown()
		})

		Context("when the bosh deployment cannot be downloaded", func() {
			BeforeEach(func() {
				setup(MockRoute{"GET", "/deployments/deployment-test", `t": "---
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
        -----END RSA PRIVATE KEY-----"
}`, ""}, "basic")
				boshConfig := &gogobosh.Config{
					Username:    "example_user",
					Password:    "example_password",
					BOSHAddress: fakeServer.URL,
				}
				boshClient, _ := gogobosh.NewClient(boshConfig)
				c = webs.CreateController(boshClient, &http.Client{})
			})

			It("returns an error", func() {
				Ω(c.LoadCerts(deployConfig, deploymentName)).Should(MatchError("invalid character '\"' in literal true (expecting 'r')"))
			})
		})

		Context("when the bosh deployment can be downloaded", func() {
			var (
				clientCert, clientKey, caCert string
			)

			JustBeforeEach(func() {
				setup(MockRoute{"GET", "/deployments/deployment-test", fmt.Sprintf(`{
  "manifest": "---\njobs:\n- name: test-job1\n  properties:\n    etcd:\n      ca_cert : %s\n      client_cert: %s\n      client_key: %s"
}`, caCert, clientCert, clientKey), ""}, "basic")
				boshConfig := &gogobosh.Config{
					Username:    "example_user",
					Password:    "example_password",
					BOSHAddress: fakeServer.URL,
				}
				boshClient, _ := gogobosh.NewClient(boshConfig)
				c = webs.CreateController(boshClient, &http.Client{})
			})

			Context("and the client key is blank", func() {
				BeforeEach(func() {
					clientKey = ""
				})

				It("returns an error", func() {
					Ω(c.LoadCerts(deployConfig, deploymentName)).Should(MatchError("Etcd Client Key was blank"))
				})
			})

			Context("and the client key is not blank", func() {
				BeforeEach(func() {
					clientKey = `|\n        -----BEGIN RSA PRIVATE KEY-----\n        MIIEpQIBAAKCAQEA7OsveeyiRFC4UOj/Ufvh8fOBSAga3kIiN/YoQxYSDu2Eth3f\n        zMsJtYCR7MbK4OWxHLMtoaJPsuTqTgqRHeVnp3GETF1bU7ROpUYdNmTCD3wsb7zF\n        cXIpyN9DlLJs3sXqwCekHcI2BDOaoKTS0FjQ0IYB56S8bNq3/UUxqFL6rYcFmos3\n        VbvYHI1MZYgTU1GgNMp1etRnoJYAMQpSbvnXCPiMil4p/l8FJMeu5esbxUMJ2K0E\n        QoQbTmzMSUNnOH3LdL3zN6YyIikRh0dy1qfIqyxFVldr08DfxnaXtKDfZZAcXJ00\n        W/H9TjUToqPoE8Dk9PJF+Px1uGfjCdSw493WOwIDAQABAoIBAQCEF+z0kdc3N7pM\n        glw4nbOwbxAZ1UsFuOHTSc/Q97FB2XQxBh8N0/ap1/GEjqH3GpnRYqWChTjiiXeJ\n        46JdhNZyKXyWB0cwOEvRInKPLUZ108oC0sFTK0/Yy4KmmYTSAI+Oa4u5e9saJJuG\n        Dd/sgliKquzM9rCIELqc38N8EifqL1dFBpDHr0+hurZU4CTW/cRMyQBD7OmjFkqM\n        BIBobjA2hrJMpc1YCmQhXyfiW6tJgvFL4Qi9Ot82+4euy6dZhgivh2D71hGh1qCw\n        qTZspTCf+wNCdpa9A/7A4DWTY2YMsK22Wou6CrZ+BD94M1zDVf0/BLBGizyh2uak\n        a+9x0u0BAoGBAP+Fk/8K0Nu9dEGm/AAjOcdByB//JiHhOKEaevywxZRS3D0ZUDg9\n        m1LBKFDeJz9AbP62ai+qtBK1jYABg6bYqUH0pAjfy6BNX+jbWmszzAfuojxswz1f\n        IFuEh8+sCl+UJgLu1n5zxzsdArUzHevxFbsI9r7wdihYC1V4GyWsZPebAoGBAO1c\n        scqljBKzCqMdOSPBk3BKGU5qISHZQm3I2xS/pKkiZz2tz1pFGJShFZeUyQ/2fHV3\n        KF8Seg112de2OYBGZxZsLPfbNZTi76VwlejL8cVaLV1PhHrElq6KL6PUScJf5dY7\n        3Q8o1FtIWn0sxzfdJkwpAWOzR12txG9O11zEuZXhAoGBAP24crsVz0vSdETYfXPZ\n        hn6vd/sljISpsWRu+d493QKpwFy+7OPbcIacm96oqInq/A9zrD7GnuXQ9r87QbGD\n        g5WlSNgy+GulSO3cGY1HMnpR3zBmwvsGoQeesohoiShc392bsMqBRjwRU2X/at+k\n        VPKSNQhllr36ps5oY3RmGR+vAoGAFd1lD0rCpXJSt4XYnp+VSlG5FQ0XsjuGMADB\n        lZ61t1LQ+dCJ+kHFKuPPzl/JSawl+NgaIu/byGOjxoglsdhKZLlgRxCtVeK1uqKt\n        XH107v4IkcDibkCvtLJMAyZqCPq2fE6VZXEYZrQ6ia9XRqEbhwZ790grecipAKvd\n        kNEaW2ECgYEA3QqP9JNAjSNPUsRcX/lCzIx39ZFcOej+7ecKiUl07uTB12BNZXF0\n        DzfsaCjL3IiAOLqtOxNw197BxDjGQ8gPE0bLm19MrULv75S0/xhsVO/SJLbNtx+z\n        mdg3V3MFR3k0U8OYeHCBas73BikOoSBro84kKj4vQupscMNE2UDERA0=\n        -----END RSA PRIVATE KEY-----`
					clientCert = ""
				})

				Context("and the client cert is blank", func() {
					It("returns an error", func() {
						Ω(c.LoadCerts(deployConfig, deploymentName)).Should(MatchError("Etcd Client Cert was blank"))
					})
				})

				Context("and the client cert is not blank", func() {
					BeforeEach(func() {
						clientCert = `|\n        -----BEGIN CERTIFICATE-----\n        MIIEMzCCAxugAwIBAgIJANDdvvIwR2SfMA0GCSqGSIb3DQEBBQUAMG4xCzAJBgNV\n        BAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5ldCBX\n        aWRnaXRzIFB0eSBMdGQxJzAlBgNVBAMTHmV0Y2QtbGVhZGVyLW1vbml0b3IudGVz\n        dC5sb2NhbDAeFw0xNjEwMTkxNDI5MzJaFw0xNzEwMTkxNDI5MzJaMG4xCzAJBgNV\n        BAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5ldCBX\n        aWRnaXRzIFB0eSBMdGQxJzAlBgNVBAMTHmV0Y2QtbGVhZGVyLW1vbml0b3IudGVz\n        dC5sb2NhbDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAOzrL3nsokRQ\n        uFDo/1H74fHzgUgIGt5CIjf2KEMWEg7thLYd38zLCbWAkezGyuDlsRyzLaGiT7Lk\n        6k4KkR3lZ6dxhExdW1O0TqVGHTZkwg98LG+8xXFyKcjfQ5SybN7F6sAnpB3CNgQz\n        mqCk0tBY0NCGAeekvGzat/1FMahS+q2HBZqLN1W72ByNTGWIE1NRoDTKdXrUZ6CW\n        ADEKUm751wj4jIpeKf5fBSTHruXrG8VDCditBEKEG05szElDZzh9y3S98zemMiIp\n        EYdHctanyKssRVZXa9PA38Z2l7Sg32WQHFydNFvx/U41E6Kj6BPA5PTyRfj8dbhn\n        4wnUsOPd1jsCAwEAAaOB0zCB0DAdBgNVHQ4EFgQUcsFsvvn6oTvpVH79uLTbj9xd\n        +hcwgaAGA1UdIwSBmDCBlYAUcsFsvvn6oTvpVH79uLTbj9xd+hehcqRwMG4xCzAJ\n        BgNVBAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5l\n        dCBXaWRnaXRzIFB0eSBMdGQxJzAlBgNVBAMTHmV0Y2QtbGVhZGVyLW1vbml0b3Iu\n        dGVzdC5sb2NhbIIJANDdvvIwR2SfMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEF\n        BQADggEBAMdREkD0s0iuBlU9wkc8kI76xxbzGzTGcZiSBYBbzR1e2Fs06t9rH5OD\n        NXg0ACoCn2JJA1hLerhdTANmz9xSrODI4g3SjEq6bvXVc3r3JNqI1+jyThD+YMHU\n        eSJY9OMg0fF1Dk0f82M+F7mHm/jS0QFBtz0gRVLCLP+FedpyxnyzAO/8yz6+YzaK\n        2zjJlk3rVz31vJu2tPuqMHaxGmbYi83+SJSXLrv9QNo2Olccaj3wxEuyK7tTxsWn\n        gD8usU9Dn2Qi3WHK9ELxP4mang1IGGeJT2+KyNG3gWlJrEIM/wdVrOKof+tqh12p\n        DctbN5FDAvf/FlJoDmjO19JimsAKaZU=\n        -----END CERTIFICATE-----`
					})

					Context("and skip ssl verification is true", func() {
						BeforeEach(func() {
							deployConfig.SkipSSLVerification = true
						})

						Context("and the client cert is invalid", func() {
							BeforeEach(func() {
								clientCert = "not a valid cert"
							})

							It("returns an error", func() {
								Ω(c.LoadCerts(deployConfig, deploymentName)).Should(MatchError("tls: failed to find any PEM data in certificate input"))
							})
						})

						Context("and the client key is invalid", func() {
							BeforeEach(func() {
								clientKey = "not a valid key"
							})

							It("returns an error", func() {
								preRunHTTPClient := *c.EtcdHTTPClient
								Ω(c.LoadCerts(deployConfig, deploymentName)).Should(MatchError("tls: failed to find any PEM data in key input"))
								Ω(*c.EtcdHTTPClient).Should(Equal(preRunHTTPClient))
							})
						})

						Context("and the client cert and key are both valid", func() {
							It("returns nil and has c.EtcdHttpClient configured as expected", func() {
								preRunHTTPClient := *c.EtcdHTTPClient
								Ω(c.LoadCerts(deployConfig, deploymentName)).Should(BeNil())
								Ω(*c.EtcdHTTPClient).ShouldNot(Equal(preRunHTTPClient))
							})
						})
					})

					Context("and skip ssl verification is false", func() {
						BeforeEach(func() {
							deployConfig.SkipSSLVerification = false
						})

						Context("and the ca cert is blank", func() {
							BeforeEach(func() {
								caCert = ""
							})

							It("returns an error", func() {
								Ω(c.LoadCerts(deployConfig, deploymentName)).Should(MatchError("Etcd CA Cert was blank"))
							})
						})

						Context("and the ca cert is not blank", func() {
							Context("and the ca cert is invalid", func() {
								BeforeEach(func() {
									caCert = "Not a valid cert"
								})

								It("returns an error", func() {
									Ω(c.LoadCerts(deployConfig, deploymentName)).Should(MatchError("Could not add CA Cert, CA Cert was likely invalid"))
								})
							})

							Context("and the ca cert is valid", func() {
								BeforeEach(func() {
									caCert = `|\n        -----BEGIN CERTIFICATE-----\n        MIIDLjCCApegAwIBAgIJAJv02yBOOO/UMA0GCSqGSIb3DQEBBQUAMG4xCzAJBgNV\n        BAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5ldCBX\n        aWRnaXRzIFB0eSBMdGQxJzAlBgNVBAMTHmV0Y2QtbGVhZGVyLW1vbml0b3IudGVz\n        dC5sb2NhbDAeFw0xNjEwMTkxNTA1NDVaFw0yNjEwMTcxNTA1NDVaMG4xCzAJBgNV\n        BAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5ldCBX\n        aWRnaXRzIFB0eSBMdGQxJzAlBgNVBAMTHmV0Y2QtbGVhZGVyLW1vbml0b3IudGVz\n        dC5sb2NhbDCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEA1ifQagZzxyndVJ3n\n        WfdknpN6TmrBe7yyaMAo52dyYbK8iSDhdevY4CvRZAMMeUG7cK/c0ednzgY+wBm1\n        jRgUr+c/DoljhouyuLAP4+5wkg/1CsDGavPesbKMaNanoif2lkbZPy0NjJBXlZ3w\n        g+QWgUUuDEgSuyRDOR7AuJ/I4RECAwEAAaOB0zCB0DAdBgNVHQ4EFgQUhw9GeDin\n        OJECyGufyH8hrLFwvvUwgaAGA1UdIwSBmDCBlYAUhw9GeDinOJECyGufyH8hrLFw\n        vvWhcqRwMG4xCzAJBgNVBAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYD\n        VQQKExhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxJzAlBgNVBAMTHmV0Y2QtbGVh\n        ZGVyLW1vbml0b3IudGVzdC5sb2NhbIIJAJv02yBOOO/UMAwGA1UdEwQFMAMBAf8w\n        DQYJKoZIhvcNAQEFBQADgYEAHSHnZbgM+lcMLF5rNpcVOc68nm2zjwAdxZNxcrHq\n        dszJDV//pxGohgFr8qyASZZ8jWusoJZgeHwU9pon0/5xZcikk0LuoYC8j1Wc3yBL\n        13YyQI1ynTSta18KYsTw1pd88AeOhHO0HTKSXOqu4l46rmtX1kmv4CEB72XBJrY3\n        lxw=\n        -----END CERTIFICATE-----`
								})

								Context("and the client cert is invalid", func() {
									BeforeEach(func() {
										clientCert = "not a valid cert"
									})

									It("returns an error", func() {
										Ω(c.LoadCerts(deployConfig, deploymentName)).Should(MatchError("tls: failed to find any PEM data in certificate input"))
									})
								})

								Context("and the client key is invalid", func() {
									BeforeEach(func() {
										clientKey = "not a valid key"
									})

									It("returns an error", func() {
										preRunHTTPClient := *c.EtcdHTTPClient
										Ω(c.LoadCerts(deployConfig, deploymentName)).Should(MatchError("tls: failed to find any PEM data in key input"))
										Ω(*c.EtcdHTTPClient).Should(Equal(preRunHTTPClient))
									})
								})

								Context("and the client cert and key are both valid", func() {
									It("returns nil and has c.EtcdHttpClient configured as expected", func() {
										preRunHTTPClient := *c.EtcdHTTPClient
										Ω(c.LoadCerts(deployConfig, deploymentName)).Should(BeNil())
										Ω(*c.EtcdHTTPClient).ShouldNot(Equal(preRunHTTPClient))
									})
								})
							})
						})
					})
				})
			})
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
				setup(MockRoute{"GET", "/stemcells", `{}`, ""}, "basic")
				boshConfig := &gogobosh.Config{
					Username:    "example_user",
					Password:    "example_password",
					BOSHAddress: fakeServer.URL,
				}
				boshClient, _ := gogobosh.NewClient(boshConfig)
				controller = webs.CreateController(boshClient, &http.Client{})
				mockRecorder = httptest.NewRecorder()
			})

			AfterEach(func() {
				teardown()
			})

			It("returns an error 500", func() {
				Ω(mockRecorder.Code).Should(Equal(500))
			})
		})

		Context("when getting bosh vms returns an error", func() {
			BeforeEach(func() {
				setupMultiple([]MockRoute{
					{"GET", "/deployments", `[
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
]`, ""},
					{"GET", "/deployments/cf-12345/vms", `{"id":1,"state":"queued","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, fakeServer.URL + "/tasks/1"},
					{"GET", "/tasks/1", `{"id":1,"state":"done","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, ""},
					{"GET", "/tasks/1/output", `":["30.30.30_id""11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
{"vm_cid":"2","ips":["31.31.31.31"],"agent_id":"2","job_name":"etcd_server-d284104a9345228c01e2","index":1}
{"vm_cid":"6","ips":["32.32.32.32"],"agent_id":"6","job_name":"etcd_server-d284104a9345228c01e2","index":2}`, ""},
				}, "basic")

				boshConfig := &gogobosh.Config{
					Username:    "example_user",
					Password:    "example_password",
					BOSHAddress: fakeServer.URL,
				}
				boshClient, _ := gogobosh.NewClient(boshConfig)
				controller = webs.CreateController(boshClient, &http.Client{})
				mockRecorder = httptest.NewRecorder()
			})

			AfterEach(func() {
				teardown()
			})

			It("returns an error 500", func() {
				Ω(mockRecorder.Code).Should(Equal(500))
			})
		})

		Context("when fetching leader stats returns an error", func() {
			BeforeEach(func() {
				setupMultiple([]MockRoute{
					{"GET", "/deployments", `[
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
]`, ""},
					{"GET", "/deployments/cf-12345/vms", `{"id":1,"state":"queued","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, fakeServer.URL + "/tasks/1"},
					{"GET", "/tasks/1", `{"id":1,"state":"done","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, ""},
					{"GET", "/tasks/1/output", `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
{"vm_cid":"2","ips":["31.31.31.31"],"agent_id":"2","job_name":"etcd_server-d284104a9345228c01e2","index":1}
{"vm_cid":"6","ips":["32.32.32.32"],"agent_id":"6","job_name":"etcd_server-d284104a9345228c01e2","index":2}`, ""},
				}, "basic")

				boshConfig := &gogobosh.Config{
					Username:    "example_user",
					Password:    "example_password",
					BOSHAddress: fakeServer.URL,
				}

				boshClient, _ := gogobosh.NewClient(boshConfig)

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

			AfterEach(func() {
				teardown()
			})

			It("returns an error 500", func() {
				Ω(mockRecorder.Code).Should(Equal(500))
			})
		})

		Context("when the number of followers is incorrect", func() {
			BeforeEach(func() {
				setupMultiple([]MockRoute{
					{"GET", "/deployments", `[
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
]`, ""},
					{"GET", "/deployments/cf-12345/vms", `{"id":1,"state":"queued","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, fakeServer.URL + "/tasks/1"},
					{"GET", "/tasks/1", `{"id":1,"state":"done","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, ""},
					{"GET", "/tasks/1/output", `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
{"vm_cid":"2","ips":["31.31.31.31"],"agent_id":"2","job_name":"etcd_server-d284104a9345228c01e2","index":1}
{"vm_cid":"6","ips":["32.32.32.32"],"agent_id":"6","job_name":"etcd_server-d284104a9345228c01e2","index":2}`, ""},
				}, "basic")

				boshConfig := &gogobosh.Config{
					Username:    "example_user",
					Password:    "example_password",
					BOSHAddress: fakeServer.URL,
				}

				boshClient, _ := gogobosh.NewClient(boshConfig)

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

			AfterEach(func() {
				teardown()
			})

			It("returns a suitable json response", func() {
				Ω(mockRecorder.Code).Should(Equal(200))
				Expect(mockRecorder.Body.String()).Should(Equal(`{"healthy": false, "message": "Incorrect number of followers"}`))
			})
		})

		Context("when more than one etcd thinks it is the leader", func() {
			BeforeEach(func() {
				setupMultiple([]MockRoute{
					{"GET", "/deployments", `[
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
]`, ""},
					{"GET", "/deployments/cf-12345/vms", `{"id":1,"state":"queued","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, fakeServer.URL + "/tasks/1"},
					{"GET", "/tasks/1", `{"id":1,"state":"done","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, ""},
					{"GET", "/tasks/1/output", `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
{"vm_cid":"2","ips":["31.31.31.31"],"agent_id":"2","job_name":"etcd_server-d284104a9345228c01e2","index":1}
{"vm_cid":"6","ips":["32.32.32.32"],"agent_id":"6","job_name":"etcd_server-d284104a9345228c01e2","index":2}`, ""},
				}, "basic")

				boshConfig := &gogobosh.Config{
					Username:    "example_user",
					Password:    "example_password",
					BOSHAddress: fakeServer.URL,
				}

				boshClient, _ := gogobosh.NewClient(boshConfig)

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

			AfterEach(func() {
				teardown()
			})

			It("returns a suitable json response", func() {
				Ω(mockRecorder.Code).Should(Equal(200))
				Expect(mockRecorder.Body.String()).Should(Equal(`{"healthy": false, "message": "Too many leaders"}`))
			})
		})

		Context("Not enough etcds are leaders", func() {
			BeforeEach(func() {
				setupMultiple([]MockRoute{
					{"GET", "/deployments", `[
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
]`, ""},
					{"GET", "/deployments/cf-12345/vms", `{"id":1,"state":"queued","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, fakeServer.URL + "/tasks/1"},
					{"GET", "/tasks/1", `{"id":1,"state":"done","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, ""},
					{"GET", "/tasks/1/output", `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
{"vm_cid":"2","ips":["31.31.31.31"],"agent_id":"2","job_name":"etcd_server-d284104a9345228c01e2","index":1}
{"vm_cid":"6","ips":["32.32.32.32"],"agent_id":"6","job_name":"etcd_server-d284104a9345228c01e2","index":2}`, ""},
				}, "basic")

				boshConfig := &gogobosh.Config{
					Username:    "example_user",
					Password:    "example_password",
					BOSHAddress: fakeServer.URL,
				}

				boshClient, _ := gogobosh.NewClient(boshConfig)

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

			AfterEach(func() {
				teardown()
			})

			It("returns a suitable json response", func() {
				Ω(mockRecorder.Code).Should(Equal(200))
				Expect(mockRecorder.Body.String()).Should(Equal(`{"healthy": false, "message": "Not enough leaders"}`))
			})
		})

		Context("When etcds are healthy and clustered correctly", func() {
			BeforeEach(func() {
				setupMultiple([]MockRoute{
					{"GET", "/deployments", `[
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
]`, ""},
					{"GET", "/deployments/cf-12345/vms", `{"id":1,"state":"queued","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, fakeServer.URL + "/tasks/1"},
					{"GET", "/tasks/1", `{"id":1,"state":"done","description":"retrieve vm-stats","timestamp":1460639781,"result":"","user":"example_user"}`, ""},
					{"GET", "/tasks/1/output", `{"vm_cid":"11","ips":["30.30.30.30"],"agent_id":"11","job_name":"etcd_server-d284104a9345228c01e2","index":0}
{"vm_cid":"2","ips":["31.31.31.31"],"agent_id":"2","job_name":"etcd_server-d284104a9345228c01e2","index":1}
{"vm_cid":"6","ips":["32.32.32.32"],"agent_id":"6","job_name":"etcd_server-d284104a9345228c01e2","index":2}`, ""},
				}, "basic")

				boshConfig := &gogobosh.Config{
					Username:    "example_user",
					Password:    "example_password",
					BOSHAddress: fakeServer.URL,
				}

				boshClient, _ := gogobosh.NewClient(boshConfig)

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

			AfterEach(func() {
				teardown()
			})

			It("returns a suitable json response", func() {
				Ω(mockRecorder.Code).Should(Equal(200))
				Expect(mockRecorder.Body.String()).Should(Equal(`{"healthy": true, "message": "Everything is healthy"}`))
			})
		})
	})
})
