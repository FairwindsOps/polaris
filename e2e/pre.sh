#!/bin/bash
set -euo pipefail

if [ -z "${CI_SHA1:-}" ]; then
  echo "CI_SHA1 not set"
  exit 1
fi

echo "CI_SHA1: ${CI_SHA1}"

tar="/tmp/workspace/docker_save/polaris_${CI_SHA1}-amd64.tar"
if [ ! -f "$tar" ]; then
  echo "Missing snapshot image at ${tar}"
  exit 1
fi

docker load --input "$tar"
docker tag "us-docker.pkg.dev/fairwinds-ops/oss/polaris:${CI_SHA1}-amd64" \
  "us-docker.pkg.dev/fairwinds-ops/oss/polaris:${CI_SHA1}"
kind load docker-image --name e2e "us-docker.pkg.dev/fairwinds-ops/oss/polaris:${CI_SHA1}"

helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --version v1.12.1 \
  --set installCRDs=true \
  --wait \
  --create-namespace

docker cp . e2e-command-runner:/polaris
