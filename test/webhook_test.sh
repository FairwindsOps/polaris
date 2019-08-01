#!/bin/bash
sed -ri "s|'(quay.io/reactiveops/polaris:).+'|'\1${CIRCLE_SHA1}'|" ./deploy/webhook.yaml
kubectl apply &> /dev/null -f ./deploy/webhook.yaml
sleep 10

kubectl apply &> /dev/null -f test/correctconfig.yaml
if [ $? -eq 0 ]; then
    VAR1="pass"
    echo pass 
else
    echo "Test Failed: Polaris prevented a deployment with no configuration issues." 
fi

kubectl apply &> /dev/null -f test/incorrectconfig.yaml
if [ $? -ne 0 ]; then
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
