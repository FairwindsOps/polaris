package validator

import (
	"time"

	conf "github.com/reactiveops/polaris/pkg/config"
	"github.com/reactiveops/polaris/pkg/kube"
)

const (
	// PolarisOutputVersion is the version of the current output structure
	PolarisOutputVersion = "0.2"
)

// ClusterSummary contains Polaris results as well as some high-level stats
type ClusterSummary struct {
	Results     ResultSummary
	Version     string
	Nodes       int
	Pods        int
	Namespaces  int
	Deployments int
	Score       uint
}

// AuditData contains all the data from a full Polaris audit
type AuditData struct {
	PolarisOutputVersion string
	AuditTime            string
	SourceType           string
	SourceName           string
	DisplayName          string
	ClusterSummary       ClusterSummary
	NamespacedResults    NamespacedResults
}

// RunAudit runs a full Polaris audit and returns an AuditData object
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
			Version:     kubeResources.ServerVersion,
			Nodes:       len(kubeResources.Nodes),
			Pods:        len(kubeResources.Pods),
			Namespaces:  len(kubeResources.Namespaces),
			Deployments: len(kubeResources.Deployments),
			Results:     clusterResults,
			Score:       clusterResults.Totals.GetScore(),
		},
		NamespacedResults: nsResults,
	}
	return auditData, nil
}
