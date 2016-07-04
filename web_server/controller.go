package webServer

import (
	"fmt"
	"github.com/FidelityInternational/etcd-leader-monitor/bosh"
	"github.com/FidelityInternational/etcd-leader-monitor/etcd"
	"net/http"
)

// Controller struct
type Controller struct {
	BoshClient     *bosh.Client
	EtcdHTTPClient *http.Client
}

// CreateController - returns a populated controller object
func CreateController(boshClient *bosh.Client, etcdHTTPClient *http.Client) *Controller {
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
	)

	fmt.Println("Checking Leaders...")
	fmt.Println("Fetching Etcd IPs from BOSH...")
	deployment, err := c.BoshClient.SearchDeployment("^cf-.+")
	if err != nil {
		fmt.Println("An error occured:")
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("Found deployment: ", deployment)
	boshVMs, err := c.BoshClient.GetEtcdVMs(deployment)
	if err != nil {
		fmt.Println("An error occured:")
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("Found Etcd VMs")
	etcdIPs := boshVMs.GetAllIPs()
	fmt.Println("Found VM IPs", etcdIPs)
	leaderList = make(map[string]map[bool]int)
	for _, etcdIP := range etcdIPs {
		etcdConfig := &etcd.Config{
			EtcdIP:     etcdIP,
			HTTPClient: c.EtcdHTTPClient,
		}
		etcdClient := etcd.NewClient(etcdConfig)
		leader, followers, err := etcdClient.GetLeaderStats()
		if err != nil {
			fmt.Println("An error occured:")
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		leaderInfo = make(map[bool]int)
		leaderInfo[leader] = followers
		leaderList[etcdIP] = leaderInfo
	}
	for _, leaderItem := range leaderList {
		for leader, followers := range leaderItem {
			if leader == true {
				leaderCount++
				if followers != (len(etcdIPs) - 1) {
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
