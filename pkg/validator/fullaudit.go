package validator

import (
	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// FairwindsOutputVersion is the version of the current output structure
	FairwindsOutputVersion = "0.0"
)

// ClusterSummary contains Fairwinds results as well as some high-level stats
type ClusterSummary struct {
	Results    ResultSummary
	Version    string
	Nodes      int
	Pods       int
	Namespaces int
}

// AuditData contains all the data from a full Fairwinds audit
type AuditData struct {
	FairwindsOutputVersion string
	ClusterSummary         ClusterSummary
	NamespacedResults      NamespacedResults
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

	clusterResults := ResultSummary{}

	// Aggregate all summary counts to get a clusterwide count.
	for _, nsRes := range nsResults {
		for _, rr := range nsRes.Results {
			clusterResults.appendResults(*rr.Summary)
		}
	}

	kubeVersion, err := kubeAPI.Clientset.Discovery().ServerVersion()
	if err != nil {
		return AuditData{}, err
	}

	listOpts := metav1.ListOptions{}
	nodes, err := kubeAPI.Clientset.CoreV1().Nodes().List(listOpts)
	if err != nil {
		return AuditData{}, err
	}
	namespaces, err := kubeAPI.Clientset.CoreV1().Namespaces().List(listOpts)
	if err != nil {
		return AuditData{}, err
	}
	numPods := 0
	for _, ns := range namespaces.Items {
		pods, err := kubeAPI.Clientset.CoreV1().Pods(ns.Name).List(listOpts)
		if err != nil {
			return AuditData{}, err
		}
		numPods += len(pods.Items)
	}

	auditData := AuditData{
		FairwindsOutputVersion: FairwindsOutputVersion,
		ClusterSummary: ClusterSummary{
			Version:    kubeVersion.Major + "." + kubeVersion.Minor,
			Nodes:      len(nodes.Items),
			Pods:       numPods,
			Namespaces: len(namespaces.Items),
			Results:    clusterResults,
		},
		NamespacedResults: nsResults,
	}
	return auditData, nil
}
