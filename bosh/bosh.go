package bosh

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/FidelityInternational/virgil/utility"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Config - used for configration of Client
type Config struct {
	Username          string
	Password          string
	BoshURI           string
	SkipSSLValidation bool
	Port              string
	HTTPClient        *http.Client
	InsecureTransport bool
}

// Client - used to communicate with BOSH
type Client struct {
	Config *Config
}

type request struct {
	method   string
	url      string
	params   url.Values
	body     io.Reader
	obj      interface{}
	username string
	password string
}

// Deployments - A collection of deployments in the director
type Deployments []*Deployment

// Deployment - A single deployment in the director
type Deployment struct {
	Name      string
	Releases  []NameVersion
	Stemcells []NameVersion
}

// NameVersion - A reusable structure for names and versions
type NameVersion struct {
	Name    string
	Version string
}

// DeploymentVMs - A collection of Deployment VMs
type DeploymentVMs []*DeploymentVM

// DeploymentVM - A Bosh VM struct
type DeploymentVM struct {
	JobName string   `json:"job_name"`
	Index   int      `json:"index"`
	VMCid   string   `json:"vm_cid"`
	AgentID string   `json:"agent_id"`
	IPs     []string `json:"ips"`
}

// TaskState - A Bosh task state struct
type TaskState struct {
	ID    int    `json:"id"`
	State string `json:"state"`
}

// NewClient - returns a new client
func NewClient(config *Config) *Client {
	const defaultRedirectLimit = 30
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}

	if config.SkipSSLValidation == true {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		config.HTTPClient = &http.Client{Transport: tr}
	}

	config.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) > defaultRedirectLimit {
			return fmt.Errorf("%d consecutive requests(redirects)", len(via))
		}
		if len(via) == 0 {
			return nil
		}

		for key, val := range via[0].Header {
			req.Header[key] = val
		}
		req.Header["Referer"] = []string{""}
		return nil
	}

	return &Client{Config: config}
}

// SearchDeployment - Returns the first depoyment name matching regex
func (c *Client) SearchDeployment(deploymentRegex string) (string, error) {
	var deployments Deployments
	r := c.newRequest("GET", "/deployments")
	resp, err := c.doRequest(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &deployments)
	if err != nil {
		return "", err
	}
	for _, deployment := range deployments {
		matched, _ := regexp.MatchString(deploymentRegex, deployment.Name)
		if matched {
			return deployment.Name, nil
		}
	}
	return "", fmt.Errorf("No deployment was found")
}

// GetEtcdVMs - Returns an array of Deployment VMs for etcd for a deployment
func (c *Client) GetEtcdVMs(deploymentName string) (DeploymentVMs, error) {
	var taskState TaskState

	requestURL := fmt.Sprintf("/deployments/%s/vms?format=full", deploymentName)
	r := c.newRequest("GET", requestURL)
	resp, err := c.doRequest(r)
	if err != nil {
		return DeploymentVMs{}, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &taskState)
	for taskState.State != "done" {
		taskState = c.isTaskComplete(taskState)
		time.Sleep(1)
	}
	body, err = c.getTaskOutput(taskState)
	if err != nil {
		return DeploymentVMs{}, err
	}
	allDeploymentVMs, err := splitDeployment(string(body))
	if err != nil {
		return DeploymentVMs{}, err
	}
	deploymentVMs := sanitiseEtcdVMs(allDeploymentVMs)
	return deploymentVMs, nil
}

// GetAllIPs - Returns an array unique IP addresses for the Deployment VMs
func (deploymentVms DeploymentVMs) GetAllIPs() []string {
	var ips []string
	for _, deployment := range deploymentVms {
		ips = append(ips, deployment.IPs...)
	}
	utility.RemoveDuplicates(&ips)
	return ips
}

func (c *Client) getTaskOutput(taskState TaskState) ([]byte, error) {
	r := c.newRequest("GET", fmt.Sprintf("/tasks/%d/output?type=result", taskState.ID))
	resp, err := c.doRequest(r)
	defer resp.Body.Close()
	if err != nil {
		return []byte{}, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func (c *Client) newRequest(method, path string) *request {
	var transport string
	transport = "https"
	if c.Config.InsecureTransport {
		transport = "http"
	}
	r := &request{
		method:   method,
		url:      transport + "://" + c.Config.BoshURI + ":" + c.Config.Port + path,
		params:   make(map[string][]string),
		username: c.Config.Username,
		password: c.Config.Password,
	}
	return r
}

func (c *Client) doRequest(r *request) (*http.Response, error) {
	req, err := r.toHTTP()
	if err != nil {
		return nil, err
	}
	resp, err := c.Config.HTTPClient.Do(req)
	return resp, err
}

func (r *request) toHTTP() (*http.Request, error) {
	// Create the HTTP request
	httpRequest, err := http.NewRequest(r.method, r.url, r.body)
	httpRequest.SetBasicAuth(r.username, r.password)
	return httpRequest, err
}

func (c *Client) isTaskComplete(taskState TaskState) TaskState {
	r := c.newRequest("GET", fmt.Sprintf("/tasks/%d", taskState.ID))
	resp, err := c.doRequest(r)
	if err != nil {
		return taskState
	}
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json.Unmarshal(body, &taskState)
	return taskState
}

func splitDeployment(deploymentStrings string) (DeploymentVMs, error) {
	var allDeploymentVMs DeploymentVMs
	splitDeployments := strings.Split(deploymentStrings, "\n")
	for _, deployment := range splitDeployments {
		if deployment == "" {
			continue
		}
		var vmDeployment DeploymentVM
		err := json.Unmarshal([]byte(deployment), &vmDeployment)
		if err != nil {
			return DeploymentVMs{}, err
		}
		allDeploymentVMs = append(allDeploymentVMs, &vmDeployment)
	}
	return allDeploymentVMs, nil
}

func sanitiseEtcdVMs(allDeploymentVMs DeploymentVMs) DeploymentVMs {
	var deploymentVMs DeploymentVMs
	for _, deploymentVM := range allDeploymentVMs {
		matched, _ := regexp.MatchString("^etcd_server.+", deploymentVM.JobName)
		if matched {
			deploymentVMs = append(deploymentVMs, deploymentVM)
		}
	}
	return deploymentVMs
}
