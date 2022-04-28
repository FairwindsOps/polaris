function check_dashboard_is_ready() {
    local timeout_epoch
    timeout_epoch=$(date -d "+2 minutes" +%s)
    echo "Waiting for dashboard to be ready"
    while ! kubectl get pods -n polaris | grep -E "dashboard.*1/1.*Running"; do
        check_timeout "${timeout_epoch}"
        echo -n "."
    done

    echo "Dashboard Running!"
}

function check_timeout() {
    local timeout_epoch="${1}"
    if [[ "$(date +%s)" -ge "${timeout_epoch}" ]]; then
        echo -e "Timeout hit waiting for readiness: exiting"
        grab_logs
        clean_up
        exit 1
    fi
}

helm repo add fairwinds-stable https://charts.fairwinds.com/stable
helm install polaris fairwinds-stable/polaris --namespace polaris --create-namespace \
  --set image.tag=$CI_SHA1

check_dashboard_is_ready

kubectl port-forward --namespace polaris svc/polaris-dashboard 3000:80 &
sleep 30
curl -f http://localhost:3000 > /dev/null
curl -f http://localhost:3000/health > /dev/null
curl -f http://localhost:3000/favicon.ico > /dev/null
curl -f http://localhost:3000/static/css/main.css > /dev/null
curl -f http://localhost:3000/results.json > /dev/null
curl -f http://localhost:3000/details/security > /dev/null
