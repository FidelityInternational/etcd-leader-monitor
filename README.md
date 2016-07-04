# etcd-leader-monitor

[![codecov.io](https://codecov.io/github/FidelityInternational/etcd-leader-monitor/coverage.svg?branch=master)](https://codecov.io/github/FidelityInternational/etcd-leader-monitor?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/FidelityInternational/etcd-leader-monitor)](https://goreportcard.com/report/github.com/FidelityInternational/etcd-leader-monitor)
[![Build Status](https://travis-ci.org/FidelityInternational/etcd-leader-monitor.svg?branch=master)](https://travis-ci.org/FidelityInternational/etcd-leader-monitor)

An application deployable to CF that checks the health state of etcd clusters.

Occasionally on CF etcd clusters can fragment, I.E having many nodes thinking they are leader or leaders not having the correct number of followers, this project is aimed at detecting when this occured so that you can monitor.

### Operation:
If all etcd nodes are responding with the correct information, this application will return a json response of: `{"healthy": true, "message": "Everything is healthy"}`
If the incorrect number of etcd leaders or followers are found an appropriate response will be returned. Examples are:
`{"healthy": false, "message": "Incorrect number of followers"}`
`{"healthy": false, "message": "Too many leaders"}`
`{"healthy": false, "message": "Not enough leaders"}`
These JSON responses are intended to make it easy to integrate with a health monitoring dashboard to continously display the health of an etcd cluster.

### Prereqs:
- This application communicates directly with bosh on port 25555 to get a list of etcd machine IPs
- This application makes http requests directly to the etcd nodes to find the etcd leader status.
- Cloudfoundry container security groups are applied on a per-space basis.
- You will need to ensure that your CF security-group rules permit communcation to bosh on port 25555 and all etcd vms on port 4001 for this applicaiton to function correctly

### Deployment

```
cf target -o <my_org> -s <my_space>
cf push --no-start
cf set-env etcd-leader-monitor BOSH_USERNAME <BOSH_DIRECTOR_USERNAME>
cf set-env etcd-leader-monitor BOSH_PASSWORD <BOSH_DIRECTOR_PASSWORD>
cf set-env etcd-leader-monitor BOSH_URI <BOSH_DIRECTOR_PRIVATE_IP>
cf set-env etcd-leader-monitor BOSH_PORT 25555
cf start etcd-leader-monitor
```

### Testing

`go test -v ./...`

This application has been tested with go version 1.6 and version 1.7.7 of the CF Go buildpack
