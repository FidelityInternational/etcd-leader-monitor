package webServer

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/FidelityInternational/etcd-leader-monitor/bosh"
	"github.com/FidelityInternational/etcd-leader-monitor/etcd"
	"github.com/caarlos0/env"
	"github.com/cloudfoundry-community/gogobosh"
	"net/http"
)

// Controller struct
type Controller struct {
	BoshClient     *gogobosh.Client
	EtcdHTTPClient *http.Client
}

// Config struct
type Config struct {
	CfDeploymentName    string `env:"CF_DEPLOYMENT_NAME" envDefault:"cf-"`
	EtcdJobName         string `env:"ETCD_JOB_NAME" envDefault:"etcd_server"`
	SSLEnabled          bool   `env:"SSL_ENABLED" envDefault:false`
	SkipSSLVerification bool   `env:"SKIP_SSL_VERIFICATION" envDefault:false`
}

// CreateController - returns a populated controller object
func CreateController(boshClient *gogobosh.Client, etcdHTTPClient *http.Client) *Controller {
	return &Controller{
		BoshClient:     boshClient,
		EtcdHTTPClient: etcdHTTPClient,
	}
}

// CheckLeaders - checks if leaders are in a healthy state
func (c *Controller) CheckLeaders(w http.ResponseWriter, r *http.Request) {
	var (
		leaderInfo          map[bool]int
		leaderList          map[string]map[bool]int
		leaderCount         int
		httpResponseMessage string
		etcdProtocol        = `http`
	)

	fmt.Println("Checking Leaders...")
	fmt.Println("Fetching Bosh deployment...")
	deployments, err := c.BoshClient.GetDeployments()
	if err != nil {
		errorPrint(err, w)
		return
	}
	deployconfig := Config{}
	env.Parse(&deployconfig)
	deployment := bosh.FindDeployment(deployments, fmt.Sprintf("^%s*", deployconfig.CfDeploymentName))
	fmt.Println("Found deployment: ", deployment)
	if deployconfig.SSLEnabled {
		fmt.Println("Fetching Etcd Certs...")
		etcdProtocol = "https"
		boshDeployment, err := c.BoshClient.GetDeployment(deployment)
		if err != nil {
			errorPrint(err, w)
			return
		}
		etcdCerts := bosh.GetEtcdCerts(boshDeployment.Manifest, fmt.Sprintf("^%s*", deployconfig.EtcdJobName))
		caCert := x509.NewCertPool()
		caCert.AppendCertsFromPEM([]byte(etcdCerts.CaCert))
		clientCert, err := tls.X509KeyPair([]byte(etcdCerts.ClientCert), []byte(etcdCerts.ClientKey))
		if err != nil {
			errorPrint(err, w)
			return
		}
		tlsConfig := &tls.Config{
			RootCAs:            caCert,
			Certificates:       []tls.Certificate{clientCert},
			InsecureSkipVerify: deployconfig.SkipSSLVerification,
		}
		tr := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		c.EtcdHTTPClient.Transport = tr
	}
	fmt.Println("Fetching Etcd IPs from BOSH...")
	boshVMs, err := c.BoshClient.GetDeploymentVMs(deployment)
	if err != nil {
		errorPrint(err, w)
		return
	}
	etcdVMs := bosh.FindVMs(boshVMs, fmt.Sprintf("^%s*", deployconfig.EtcdJobName))
	fmt.Println("Found Etcd VMs")
	leaderList = make(map[string]map[bool]int)
	for _, etcdVM := range etcdVMs {
		etcdConfig := &etcd.Config{
			EtcdIP:       etcdVM.IPs[0],
			HTTPClient:   c.EtcdHTTPClient,
			EtcdProtocol: etcdProtocol,
		}
		etcdClient := etcd.NewClient(etcdConfig)
		leader, followers, err := etcdClient.GetLeaderStats()
		if err != nil {
			errorPrint(err, w)
			return
		}
		leaderInfo = make(map[bool]int)
		leaderInfo[leader] = followers
		leaderList[etcdVM.IPs[0]] = leaderInfo
	}
	for _, leaderItem := range leaderList {
		for leader, followers := range leaderItem {
			if leader == true {
				leaderCount++
				if followers != (len(etcdVMs) - 1) {
					httpResponseMessage = `{"healthy": false, "message": "Incorrect number of followers"}`
				}
			}
		}
	}
	if leaderCount > 1 {
		fmt.Println("More than one etcd leader detected, number of leaders: ", leaderCount)
		httpResponseMessage = `{"healthy": false, "message": "Too many leaders"}`
	} else if leaderCount == 0 {
		fmt.Println("Not enough etcd leaders detected, number of leaders: ", leaderCount)
		httpResponseMessage = `{"healthy": false, "message": "Not enough leaders"}`
	} else if httpResponseMessage == "" {
		httpResponseMessage = `{"healthy": true, "message": "Everything is healthy"}`
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, httpResponseMessage)
}

func errorPrint(err error, w http.ResponseWriter) {
	if err != nil {
		fmt.Println("An error occured:")
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
