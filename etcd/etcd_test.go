package etcd_test

import (
	"crypto/tls"
	"fmt"
	"github.com/FidelityInternational/etcd-leader-monitor/etcd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"net/url"
)

var _ = Describe("#NewClient", func() {
	It("creates an etcd client", func() {
		config := &etcd.Config{
			EtcdIP:     "1.1.1.1",
			HTTPClient: &http.Client{},
		}
		client := etcd.NewClient(config)
		Ω(client.Config.HTTPClient).Should(BeAssignableToTypeOf(&http.Client{}))
		Ω(client.Config.EtcdIP).Should(Equal("1.1.1.1"))
	})
})

var _ = Describe("#GetLeaderStats", func() {
	Context("when the http requests raises an error", func() {
		var client *etcd.Client

		BeforeEach(func() {
			config := &etcd.Config{
				EtcdIP:     "1.1.1.1:1",
				HTTPClient: &http.Client{},
			}
			client = etcd.NewClient(config)
		})

		It("returns the error", func() {
			_, _, err := client.GetLeaderStats()
			Ω(err).Should(MatchError(`Get http://1.1.1.1:1:4001/v2/stats/leader: dial tcp: too many colons in address 1.1.1.1:1:4001`))
		})
	})

	Context("when unmarshalling the json response raises an error", func() {
		var client *etcd.Client

		BeforeEach(func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `leader":"6a0b69a54415a491","followers":{"a0294459200078aa":{"latency":{"current":0.001199,"average":0.0023682517720168754,"standardDeviation":0.4302199179552562,"minimum":0.000654,"maximum":1996.564157},"counts":{"fail":16,"success":21538911}},"b5c352b4495e4195":{"latency":{"current":0.001609,"average":0.002361467019756358,"standardDeviation":0.00506414137059054,"minimum":0.00088,"maximum":5.153269},"counts":{"fail":7,"success":1617908}}}}`)
			}))

			transport := &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(server.URL)
				},
				TLSClientConfig: &tls.Config{},
			}
			httpClient := &http.Client{Transport: transport}

			config := &etcd.Config{
				EtcdIP:     "1.1.1.1",
				HTTPClient: httpClient,
			}
			client = etcd.NewClient(config)
		})

		It("returns the error", func() {
			_, _, err := client.GetLeaderStats()
			Ω(err).Should(MatchError("invalid character 'l' looking for beginning of value"))
		})
	})

	Context("when no errors are raised", func() {
		Context("and the etcd is a leader", func() {
			var client *etcd.Client

			BeforeEach(func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintln(w, `{"leader":"6a0b69a54415a491","followers":{"a0294459200078aa":{"latency":{"current":0.001199,"average":0.0023682517720168754,"standardDeviation":0.4302199179552562,"minimum":0.000654,"maximum":1996.564157},"counts":{"fail":16,"success":21538911}},"b5c352b4495e4195":{"latency":{"current":0.001609,"average":0.002361467019756358,"standardDeviation":0.00506414137059054,"minimum":0.00088,"maximum":5.153269},"counts":{"fail":7,"success":1617908}}}}`)
				}))

				transport := &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(server.URL)
					},
					TLSClientConfig: &tls.Config{},
				}
				httpClient := &http.Client{Transport: transport}

				config := &etcd.Config{
					EtcdIP:     "1.1.1.1",
					HTTPClient: httpClient,
				}
				client = etcd.NewClient(config)
			})

			It("returns a count of followers and leader true", func() {
				leader, followers, _ := client.GetLeaderStats()
				Ω(leader).Should(BeTrue())
				Ω(followers).Should(Equal(2))
			})
		})

		Context("and the etcd is not a leader", func() {
			var client *etcd.Client

			BeforeEach(func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintln(w, `{"message":"not current leader"}`)
				}))

				transport := &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(server.URL)
					},
					TLSClientConfig: &tls.Config{},
				}
				httpClient := &http.Client{Transport: transport}

				config := &etcd.Config{
					EtcdIP:     "1.1.1.1",
					HTTPClient: httpClient,
				}
				client = etcd.NewClient(config)
			})

			It("returns leader false", func() {
				leader, _, _ := client.GetLeaderStats()
				Ω(leader).Should(BeFalse())
			})
		})
	})
})
