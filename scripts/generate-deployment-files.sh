# /bin/bash

set -e

helm template polaris $CHARTS_DIR/stable/polaris/ \
  --namespace polaris \
  --set templateOnly=true \
  > deploy/dashboard.yaml

helm template polaris $CHARTS_DIR/stable/polaris/ \
  --namespace polaris \
  --set templateOnly=true \
  --set webhook.enable=true \
  --set dashboard.enable=false \
  > deploy/webhook.yaml
