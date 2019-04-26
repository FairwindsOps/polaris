helm-to-k8s:
	helm template deploy/helm/fairwinds/ --name fairwinds --namespace fairwinds --set templateOnly=true > deploy/dashboard.yaml
	helm template deploy/helm/fairwinds/ --name fairwinds --namespace fairwinds --set templateOnly=true --set webhook.enable=true dashboard.enable=false > deploy/webhook.yaml
