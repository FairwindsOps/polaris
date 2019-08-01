#!/bin/bash
set -x
sed -ri "s|'(quay.io/reactiveops/polaris:).+'|'\1${CIRCLE_SHA1}'|" ./deploy/webhook.yaml
kubectl apply -f ./deploy/webhook.yaml
sleep 20
kubectl apply -f test/correctconfig.yaml
status=$?
sleep 20 


if [ status -eq 0 ]; then
    VAR1="pass"
    echo pass 
else
    echo "Test Failed: Polaris prevented a deployment with no configuration issues." 
fi

kubectl apply -f test/incorrectconfig.yaml
status=$?
sleep 20
if [ status -ne 0 ]; then
    VAR2="pass"
    echo pass 
else
    echo "Test Failed: Polaris should have prevented this deployment due to configuration problems."
fi

if [ "$VAR1" == "pass" -a "$VAR2" == "pass" ]; then
    echo "Tests Passed."
else
    echo "Tests Failed"
    exit 1
fi
