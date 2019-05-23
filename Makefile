helm-to-k8s:
	helm repo add reactiveops-stable https://charts.reactiveops.com/stable
	# TODO: once we're on Helm 3, we can template using remote repos
	helm fetch --untar --untardir ./polaris-helm 'reactiveops-stable/polaris'
	helm template ./polaris-helm/polaris --name polaris --namespace polaris --set templateOnly=true > deploy/dashboard.yaml
	helm template ./polaris-helm/polaris --name polaris --namespace polaris --set templateOnly=true --set webhook.enable=true --set dashboard.enable=false > deploy/webhook.yaml
	rm -r ./polaris-helm
