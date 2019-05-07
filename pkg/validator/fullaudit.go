package validator

import (
	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
)

const (
	// FairwindsOutputVersion is the version of the current output structure
	FairwindsOutputVersion = "0.0"
)

// ClusterSummary contains Fairwinds results as well as some high-level stats
type ClusterSummary struct {
	Results     ResultSummary
	Version     string
	Nodes       int
	Pods        int
	Namespaces  int
	Deployments int
}

// AuditData contains all the data from a full Fairwinds audit
type AuditData struct {
	FairwindsOutputVersion string
	ClusterSummary         ClusterSummary
	NamespacedResults      NamespacedResults
}

// RunAudit runs a full Fairwinds audit and returns an AuditData object
func RunAudit(config conf.Configuration, kubeResources *kube.ResourceProvider) (AuditData, error) {
	// TODO: Validate StatefulSets, DaemonSets, Cron jobs
	// in addition to deployments

	// TODO: Once we are validating more than deployments,
	// we will need to merge the namespaceResults that get returned
	// from each validation.
	nsResults, err := ValidateDeployments(config, kubeResources)
	if err != nil {
		return AuditData{}, err
	}

	clusterResults := ResultSummary{}

	// Aggregate all summary counts to get a clusterwide count.
	for _, nsRes := range nsResults {
		for _, dr := range nsRes.DeploymentResults {
			clusterResults.appendResults(*dr.PodResult.Summary)
		}
	}

	auditData := AuditData{
		FairwindsOutputVersion: FairwindsOutputVersion,
		ClusterSummary: ClusterSummary{
			Version:     kubeResources.ServerVersion,
			Nodes:       len(kubeResources.Nodes),
			Pods:        len(kubeResources.Pods),
			Namespaces:  len(kubeResources.Namespaces),
			Deployments: len(kubeResources.Deployments),
			Results:     clusterResults,
		},
		NamespacedResults: nsResults,
	}
	return auditData, nil
}
