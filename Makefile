helm-to-k8s:
	helm template deploy/helm/polaris/ --name polaris --namespace polaris --set templateOnly=true > deploy/dashboard.yaml
	helm template deploy/helm/polaris/ --name polaris --namespace polaris --set templateOnly=true --set webhook.enable=true --set dashboard.enable=false > deploy/webhook.yaml
