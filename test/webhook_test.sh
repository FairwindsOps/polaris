#!/bin/bash

kubectl apply  &> /dev/null -f correctconfig.yaml
if [ $? -eq 0 ]; then
    VAR1="pass"
    echo pass 
else
    echo "Test Failed: " 
    kubectl apply -f correctconfig.yaml
fi

kubectl apply &> /dev/null -f incorrectconfig.yaml
if [ $? -ne 0 ]; then
    VAR2="pass"
    echo pass 
else
    echo "Test Failed: Polaris should have prevented this deployment due to configuration problems."
fi

if [ $VAR1 = "pass" -a $VAR2 = "pass" ]; then
    echo "Tests Passed."
else
    echo "Tests Failed"
fi
