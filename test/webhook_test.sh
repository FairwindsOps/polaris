#!/bin/bash
set -e

#sed is replacing the polaris version with this commit sha so we are testing exactly this verison.
sed -ri "s|'(quay.io/reactiveops/polaris:).+'|'\1${CIRCLE_SHA1}'|" ./deploy/webhook.yaml

# Testing to ensure that the webhook starts up, allows a correct deployment to pass,
# and prevents a incorrectly formatted deployment. 
function check_webhook_is_ready() {
    # Get the epoch time in one minute from now
    local timeout_epoch

    # Reset another 2 minutes to wait for webhook
    timeout_epoch=$(date -d "+2 minutes" +%s)

    # loop until this fails (desired condition is we cannot apply this yaml doc, which means the webhook is working
    echo "Waiting for webhook to be ready"
    while ! kubectl get pods -n polaris | grep -E "webhook.*1/1.*Running"; do
        check_timeout "${timeout_epoch}"
        echo -n "."
    done

    echo "Webhook started!"
}

# Check if timeout is hit and exit if it is
function check_timeout() {
    local timeout_epoch="${1}"
    if [[ "$(date +%s)" -ge "${timeout_epoch}" ]]; then
        echo -e "Timeout hit waiting for readiness: exiting"
        grab_logs
        clean_up
        exit 1
    fi

}

# Clean up all your stuff
function clean_up() {
    # Clean up files you've installed (helps with local testing)
    for filename in test/*yaml; do
        # || true to avoid issues when we cannot delete
        kubectl delete -f $filename &>/dev/null ||true
    done
    # Uninstall webhook and webhook config
    kubectl delete validatingwebhookconfigurations polaris-webhook --wait=false &>/dev/null
    kubectl -n polaris delete deploy -l app=polaris --wait=false &>/dev/null
}

function grab_logs() {
    kubectl -n polaris get pods -oyaml -l app=polaris
    kubectl -n polaris describe pods -l app=polaris
    kubectl -n polaris logs -l app=polaris
}

# Install the webhook 
kubectl apply -f ./deploy/webhook.yaml &> /dev/null


# wait for the webhook to come online
check_webhook_is_ready


# Webhook started, setting all tests as passed initially.
ALL_TESTS_PASSED=1

# Run tests against correctly configured objects
for filename in test/passing_test.*.yaml; do
    echo $filename
    if ! kubectl apply -f $filename &> /dev/null; then
        ALL_TESTS_PASSED=0
        echo "Test Failed: Polaris prevented a deployment with no configuration issues." 
    fi
done

# Run tests against incorrectly configured objects
for filename in test/failing_test.*.yaml; do
    echo $filename
    if kubectl apply -f $filename &> /dev/null; then
        ALL_TESTS_PASSED=0
        echo "Test Failed: Polaris should have prevented this deployment due to configuration issues."
    fi
done

clean_up

#Verify that all the tests passed.
if [ $ALL_TESTS_PASSED -eq 1 ]; then
    echo "Tests Passed."
else
    echo "Tests Failed."
    exit 1
fi
