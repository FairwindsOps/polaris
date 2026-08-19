#!/bin/bash
set -euo pipefail

if [[ -n "${CIRCLE_PR_NUMBER:-}" ]]; then
  echo "Skipping Kubernetes tests for forked PR"
  exit 0
fi

cd /polaris
./test/webhook_test.sh
./test/kube_dashboard_test.sh
