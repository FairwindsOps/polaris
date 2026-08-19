#!/bin/bash
set -euo pipefail

mkdir -p /tmp/test-results

if [[ -n "${CIRCLE_PR_NUMBER:-}" ]]; then
  echo "Skipping Kubernetes tests for forked PR"
  exit 0
fi

cd /polaris

helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --version v1.12.1 \
  --set installCRDs=true \
  --wait \
  --create-namespace

./test/webhook_test.sh
./test/kube_dashboard_test.sh
