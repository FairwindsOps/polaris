package validator

import (
	"time"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
)

// RunAudit runs a full Polaris audit and returns an AuditData object
func RunAudit(config conf.Configuration, kubeResources *kube.ResourceProvider) (AuditData, error) {
	nsResults := NamespacedResults{}
	ValidateControllers(config, kubeResources, &nsResults)

	clusterResults := ResultSummary{}

	// Aggregate all summary counts to get a clusterwide count.
	for _, result := range nsResults.GetAllControllerResults() {
		clusterResults.appendResults(*result.PodResult.Summary)
	}

	displayName := config.DisplayName
	if displayName == "" {
		displayName = kubeResources.SourceName
	}

	auditData := AuditData{
		PolarisOutputVersion: PolarisOutputVersion,
		AuditTime:            kubeResources.CreationTime.Format(time.RFC3339),
		SourceType:           kubeResources.SourceType,
		SourceName:           kubeResources.SourceName,
		DisplayName:          displayName,
		ClusterSummary: ClusterSummary{
			Version:                kubeResources.ServerVersion,
			Nodes:                  len(kubeResources.Nodes),
			Pods:                   len(kubeResources.Pods),
			Namespaces:             len(kubeResources.Namespaces),
			Deployments:            len(kubeResources.Deployments),
			StatefulSets:           len(kubeResources.StatefulSets),
			DaemonSets:             len(kubeResources.DaemonSets),
			Jobs:                   len(kubeResources.Jobs),
			CronJobs:               len(kubeResources.CronJobs),
			ReplicationControllers: len(kubeResources.ReplicationControllers),
			Results:                clusterResults,
			Score:                  clusterResults.Totals.GetScore(),
		},
		NamespacedResults: nsResults,
	}
	return auditData, nil
}
