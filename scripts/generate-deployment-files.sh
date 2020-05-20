# /bin/bash

set -eo pipefail

helm template $CHARTS_DIR/stable/polaris/ \
  --name polaris --namespace polaris \
  --set templateOnly=true \
  > deploy/dashboard.yaml

helm template $CHARTS_DIR/stable/polaris/ \
  --name polaris --namespace polaris \
  --set templateOnly=true \
  --set webhook.enable=true \
  --set dashboard.enable=false \
  > deploy/webhook.yaml
