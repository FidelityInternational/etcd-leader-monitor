# etcd-leader-monitor

[![codecov.io](https://codecov.io/github/FidelityInternational/etcd-leader-monitor/coverage.svg?branch=master)](https://codecov.io/github/FidelityInternational/etcd-leader-monitor?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/FidelityInternational/etcd-leader-monitor)](https://goreportcard.com/report/github.com/FidelityInternational/etcd-leader-monitor)
[![Build Status](https://travis-ci.org/FidelityInternational/etcd-leader-monitor.svg?branch=master)](https://travis-ci.org/FidelityInternational/etcd-leader-monitor)

An application deployable to CF that checks the health state of etcd clusters.

Occasionally on CF etcd clusters can fragment, I.E having many nodes thinking they are leader or leaders not having the correct number of followers, this project is aimed at detecting when this occured so that you can monitor.

### Operation:
If all etcd nodes are responding with the correct information, this application will return a json response of:

`{"healthy": true, "message": "Everything is healthy"}`

If the incorrect number of etcd leaders or followers are found an appropriate response will be returned. Examples are:

`{"healthy": false, "message": "Incorrect number of followers"}`

`{"healthy": false, "message": "Too many leaders"}`

`{"healthy": false, "message": "Not enough leaders"}`

These JSON responses are intended to make it easy to integrate with a health monitoring dashboard to continously display the health of an etcd cluster.

### Prereqs:
- This application communicates directly with bosh on port 25555 (and 8443 to use UAA) to get a list of etcd machine IPs
- This application makes http requests directly to the etcd nodes to find the etcd leader status.
- Cloudfoundry container security groups are applied on a per-space basis.
- You will need to ensure that your CF security-group rules permit communcation to bosh on port 25555 and 8443 and all etcd vms on port 4001 for this applicaiton to function correctly
- By default the application expects its cloudfoundry deployment name to start with `cf-` and etcd job name with `etcd_server`, for custom config set environment variables as described in below manual deployment steps.
- By default the application will connect to the etcd servers using http. If you wish to use SSL (TLS) then set the `SSL_ENABLED` environment variable to `true`.

**Note**: When `SSL_ENABLED=true` has been set you may get certificate mismatch errors as the applcication will connect using the IP address rather than DNS name. For these use cases also set `SKIP_SSL_VERIFICATION=true`

### Deployment

#### Manual deployment

```
cf target -o <my_org> -s <my_space>
cf push --no-start
cf set-env etcd-leader-monitor BOSH_USERNAME <BOSH_USERNAME>
cf set-env etcd-leader-monitor BOSH_PASSWORD <BOSH_PASSWORD>
cf set-env etcd-leader-monitor BOSH_URI <https://10.0.0.6:25555>
cf set-env etcd-leader-monitor CF_DEPLOYMENT_NAME <CF_DEPLOYMENT_NAME>
cf set-env etcd-leader-monitor ETCD_JOB_NAME <ETCD_JOB_NAME>
cf set-env etcd-leader-monitor SKIP_SSL_VERIFICATION <true|false> \
cf set-env etcd-leader-monitor SSL_ENABLED <true|false> \
cf start etcd-leader-monitor
```

#### Automated zero-downtime deployment

```
BOSH_USERNAME=<BOSH_USERNAME> \
BOSH_PASSWORD=<BOSH_PASSWORD>\
BOSH_URI=<https://10.0.0.6:25555> \
CF_SYS_DOMAIN=<system.example.com> \
CF_DEPLOY_USERNAME=<CF_USERNAME> \
CF_DEPLOY_PASSWORD=<CF_PASSWORD> \
CF_NETWORKS=<10.0.0.0/24,11.0.0.0/24> \
CF_DEPLOYMENT_NAME=<CF_DEPLOYMENT_NAME> \
ETCD_JOB_NAME=<ETCD_JOB_NAME> \
SKIP_SSL_VERIFICATION=<true|false> \
SSL_ENABLED=<true|false> \
APP_NAME=<ETCD_LEADER_MONITOR_APP_NAME> \
./deploy.sh
```

##### Example

```
BOSH_USERNAME=director \
BOSH_PASSWORD=123456789abcdef \
BOSH_URI=https://10.0.0.6:25555 \
CF_SYS_DOMAIN=system.example.cf.com \
CF_DEPLOY_USERNAME=cf_admin \
CF_DEPLOY_PASSWORD=123456789abcdef \
CF_NETWORKS=10.0.0.0/24,11.0.0.0/24 \
CF_DEPLOYMENT_NAME='cf-' \
ETCD_JOB_NAME=diego_database \
SKIP_SSL_VERIFICATION=true \
SSL_ENABLED=true \
APP_NAME=my_etcd_leader_monitor_app \
./deploy.sh
```

### Testing

`go test -v ./...`

#### Smoke Tests

```
APP_URL=<etcd-leader-monitor.apps.example.com> \
./smoke-test.sh
```

This application has been tested with go version 1.6 and version 1.7.7 of the CF Go buildpack
