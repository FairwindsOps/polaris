#!/bin/bash
set -euo pipefail

docker exec e2e-command-runner mkdir -p /tmp/test-results

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

export PATH="$(pwd)/bin-kind:${PATH}"
kind version

docker load --input "$tar"
docker tag "us-docker.pkg.dev/fairwinds-ops/oss/polaris:${CI_SHA1}-amd64" \
  "us-docker.pkg.dev/fairwinds-ops/oss/polaris:${CI_SHA1}"
kind load docker-image --name e2e "us-docker.pkg.dev/fairwinds-ops/oss/polaris:${CI_SHA1}"

docker cp . e2e-command-runner:/polaris
