# /bin/bash

set -eo pipefail

helm template $CHARTS_DIR/stable/polaris/ \
  --name polaris --namespace polaris \
  --set templateOnly=true \
  --set config="$(cat ./examples/config.yaml)" \
  > deploy/dashboard.yaml

helm template $CHARTS_DIR/stable/polaris/ \
  --name polaris --namespace polaris \
  --set templateOnly=true \
  --set webhook.enable=true \
  --set dashboard.enable=false \
  --set config="$(cat ./examples/config.yaml)" \
  > deploy/webhook.yaml
