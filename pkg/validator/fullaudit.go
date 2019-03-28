package validator

import (
	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
)

// AuditData contains all the data from a full Fairwinds audit
type AuditData struct {
	ClusterSummary    ResultSummary
	NamespacedResults NamespacedResults
}

// RunAudit runs a full Fairwinds audit and returns an AuditData object
func RunAudit(config conf.Configuration, kubeAPI *kube.API) (AuditData, error) {
	// TODO: Validate StatefulSets, DaemonSets, Cron jobs
	// in addition to deployments

	// TODO: Once we are validating more than deployments,
	// we will need to merge the namespaceResults that get returned
	// from each validation.
	nsResults, err := ValidateDeploys(config, kubeAPI)
	if err != nil {
		return AuditData{}, err
	}

	var clusterSuccesses, clusterErrors, clusterWarnings uint

	// Aggregate all summary counts to get a clusterwide count.
	for _, nsRes := range nsResults {
		for _, rr := range nsRes.Results {
			clusterErrors += rr.Summary.Errors
			clusterWarnings += rr.Summary.Warnings
			clusterSuccesses += rr.Summary.Successes
		}
	}

	auditData := AuditData{
		ClusterSummary: ResultSummary{
			Errors:  clusterErrors,
			Warnings:  clusterWarnings,
			Successes: clusterSuccesses,
		},
		NamespacedResults: nsResults,
	}
	return auditData, nil
}
