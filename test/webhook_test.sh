#!/bin/bash
#Testing to ensure that the webhook starts up, allows a correct deployment to pass,
#and prevents a incorrectly formatted deployment. 
set -e
#sed is replacing the polaris version with this commit sha so we are testing exactly this verison.
sed -ri "s|'(quay.io/reactiveops/polaris:).+'|'\1${CIRCLE_SHA1}'|" ./deploy/webhook.yaml

kubectl apply -f ./deploy/webhook.yaml &> /dev/null
timeout=100
while kubectl apply -f test/failing_test.deployment.yaml &> /dev/null; do
  echo "Waiting for webhook to start..."
  if [ $timeout -eq 0 ]; then
    echo "Timed out while waiting for webhook to start"
    exit 1
  fi
  timeout=$((timeout-1))
  sleep 1
  kubectl delete nginx-deployment
done
echo "Webhook started!"

#Webhook started, setting all tests as passed initially.
ALL_TESTS_PASSED=1

if ! kubectl apply -f test/passing_test.deployment.yaml &> /dev/null; then
    ALL_TESTS_PASSED=0
    echo "Test Failed: Polaris prevented a deployment with no configuration issues." 
fi

if kubectl apply -f test/failing_test.deployment.yaml; then
    ALL_TESTS_PASSED=0
    echo "Test Failed: Polaris should have prevented this deployment due to configuration issues."
fi

#Verify that all the tests passed.
if [ $ALL_TESTS_PASSED -eq 1 ]; then
    echo "Tests Passed."
else
    echo "Tests Failed."
    exit 1
fi
