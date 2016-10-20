#! /bin/bash

set -e

trap "rm security_group.json" EXIT

setEnvs () {
  cf set-env "$1" BOSH_USERNAME "${BOSH_USERNAME}"
  cf set-env "$1" BOSH_PASSWORD "${BOSH_PASSWORD}"
  cf set-env "$1" BOSH_URI "${BOSH_URI}"
  cf set-env "$1" BOSH_PORT 25555
  if [ -n "${CF_DEPLOYMENT_NAME}" ]; then
    cf set-env "$1" CF_DEPLOYMENT_NAME "${CF_DEPLOYMENT_NAME}"
  fi
  if [ -n "${ETCD_JOB_NAME}" ]; then
    cf set-env "$1" ETCD_JOB_NAME "${ETCD_JOB_NAME}"
  fi
  if [ -n "${SSL_ENABLED}" ]; then
    cf set-env "$1" SSL_ENABLED "${SSL_ENABLED}"
  fi
  if [ -n "${SKIP_SSL_VERIFICATION}" ]; then
    cf set-env "$1" SKIP_SSL_VERIFICATION "${SKIP_SSL_VERIFICATION}"
  fi
}

echo "Logging into CF..."
cf api https://api."${CF_SYS_DOMAIN}" --skip-ssl-validation
cf auth "${CF_DEPLOY_USERNAME}" "${CF_DEPLOY_PASSWORD}"
echo "Creating Org etcd-leader-monitor..."
cf create-org etcd-leader-monitor
echo "Targetting Org etcd-leader-monitor..."
cf target -o etcd-leader-monitor
echo "Creating Space etcd-leader-monitor..."
cf create-space etcd-leader-monitor
echo "Targetting Space etcd-leader-monitor..."
cf target -s etcd-leader-monitor
echo "Setting up security groups..."
if [ -z "${CF_NETWORKS}" ]; then
# Disable this shellcheck warning as we know how to spell correctly.
# shellcheck disable=SC2153
CF_NETWORKS="${CF_NETWORK}"
fi

echo "[" > security_group.json
last_subnet=$(echo "${CF_NETWORKS}" | awk -F, '{print $NF}')
for subnet in ${CF_NETWORKS//,/ } ;
do
cat >> security_group.json << EOF
  {
    "destination": "$subnet",
    "protocol": "tcp",
    "ports": "25555, 4001, 8443"
EOF
if [ "${subnet}" != "${last_subnet}" ];
then
  echo "  }," >> security_group.json
else
  echo "  }" >> security_group.json
fi

done
echo "]" >> security_group.json

if cf create-security-group etcd-leader-monitor security_group.json | grep -q "already exists"; then
  cf update-security-group etcd-leader-monitor security_group.json
fi
cf bind-security-group etcd-leader-monitor etcd-leader-monitor etcd-leader-monitor
echo "Deploying apps..."

if [[ "$(cf app etcd-leader-monitor) || true)" == *"FAILED"* ]] ; then
  cf push "${APP_NAME:-etcd-leader-monitor}" --no-start
  setEnvs "${APP_NAME:-etcd-leader-monitor}"
  cf start "${APP_NAME:-etcd-leader-monitor}"
else
  echo "Zero downtime deploying etcd-leader-monitor..."
  domain=$(cf app etcd-leader-monitor | grep urls | cut -d":" -f2 | xargs | cut -d"." -f 2-)
  cf push "${APP_NAME:-etcd-leader-monitor}-green" -f manifest.yml -n "${APP_NAME:-etcd-leader-monitor}-green" --no-start
  setEnvs "${APP_NAME:-etcd-leader-monitor}-green"
  cf start "${APP_NAME:-etcd-leader-monitor}-green"
  cf map-route "${APP_NAME:-etcd-leader-monitor}-green" "${domain}" -n "${APP_NAME:-etcd-leader-monitor}"
  cf delete "${APP_NAME:-etcd-leader-monitor}" -f
  cf rename "${APP_NAME:-etcd-leader-monitor}-green" "${APP_NAME:-etcd-leader-monitor}"
  cf unmap-route "${APP_NAME:-etcd-leader-monitor}" "${domain}" -n "${APP_NAME:-etcd-leader-monitor}-green"
fi
