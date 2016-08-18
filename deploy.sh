#! /bin/bash

set -u
set -e

trap "rm security_group.json" EXIT

echo "Logging into CF..."
cf api https://api."$CF_SYS_DOMAIN" --skip-ssl-validation
cf auth "$CF_DEPLOY_USERNAME" "$CF_DEPLOY_PASSWORD"
echo "Creating Org etcd-leader-monitor..."
cf create-org etcd-leader-monitor
echo "Targetting Org etcd-leader-monitor..."
cf target -o etcd-leader-monitor
echo "Creating Space etcd-leader-monitor..."
cf create-space etcd-leader-monitor
echo "Targetting Space etcd-leader-monitor..."
cf target -s etcd-leader-monitor
echo "Setting up security groups..."
cat > security_group.json << EOF
[
  {
    "destination": "$CF_NETWORK",
    "protocol": "tcp",
    "ports": "25555, 4001, 8443"
  }
]
EOF
if cf create-security-group etcd-leader-monitor security_group.json | grep -q "already exists"; then
  cf update-security-group etcd-leader-monitor security_group.json
fi
cf bind-security-group etcd-leader-monitor etcd-leader-monitor etcd-leader-monitor
echo "Deploying apps..."

if [[ "$(cf app etcd-leader-monitor) || true)" == *"FAILED"* ]] ; then
  echo "Zero downtime deploying etcd-leader-monitor..."
  domain=$(cf app etcd-leader-monitor | grep urls | cut -d":" -f2 | xargs | cut -d"." -f 2-)
  cf push etcd-leader-monitor-green -f manifest.yml -n etcd-leader-monitor-green --no-start
  cf set-env etcd-leader-monitor-green BOSH_USERNAME "$BOSH_USERNAME"
  cf set-env etcd-leader-monitor-green BOSH_PASSWORD "$BOSH_PASSWORD"
  cf set-env etcd-leader-monitor-green BOSH_URI "$BOSH_URI"
  cf set-env etcd-leader-monitor-green BOSH_PORT 25555
  cf start etcd-leader-monitor-green
  cf map-route etcd-leader-monitor-green "$domain" -n etcd-leader-monitor
  cf delete etcd-leader-monitor -f
  cf rename etcd-leader-monitor-green etcd-leader-monitor
  cf unmap-route etcd-leader-monitor "$domain" -n etcd-leader-monitor-green
else
  cf push --no-start
  cf set-env etcd-leader-monitor BOSH_USERNAME "$BOSH_USERNAME"
  cf set-env etcd-leader-monitor BOSH_PASSWORD "$BOSH_PASSWORD"
  cf set-env etcd-leader-monitor BOSH_URI "$BOSH_URI"
  cf set-env etcd-leader-monitor BOSH_PORT 25555
  cf start etcd-leader-monitor
fi
