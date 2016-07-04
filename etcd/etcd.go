package etcd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Config - used for configration of Client
type Config struct {
	EtcdIP     string
	HTTPClient *http.Client
}

// Client - used to communicate with Etcd
type Client struct {
	Config *Config
}

type etcdLeader struct {
	Message   string              `json:"message"`
	Leader    string              `json:"leader"`
	Followers map[string]struct{} `json:"followers"`
}

// NewClient - returns a new client
func NewClient(config *Config) *Client {
	return &Client{Config: config}
}

// GetLeaderStats - returns leader true/false and count of followers
func (c *Client) GetLeaderStats() (bool, int, error) {
	var etcdLeader etcdLeader
	resp, err := c.Config.HTTPClient.Get(fmt.Sprintf("http://%s:4001/v2/stats/leader", c.Config.EtcdIP))
	defer resp.Body.Close()
	if err != nil {
		return false, 0, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &etcdLeader)
	if err != nil {
		return false, 0, err
	}
	if etcdLeader.Leader != "" {
		return true, len(etcdLeader.Followers), nil
	}
	return false, 0, nil
}
