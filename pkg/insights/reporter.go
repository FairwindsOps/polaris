package insights

import (
	"encoding/json"
	"fmt"
	"time"

	workloads "github.com/fairwindsops/insights-plugins/plugins/workloads/pkg"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/sirupsen/logrus"
)

type insightsReporter struct {
	client Client
}

func NewInsightsReporter(client Client) *insightsReporter {
	return &insightsReporter{
		client: client,
	}
}

type WorkloadsReport struct {
	Version string
	Payload workloads.ClusterWorkloadReport
}

type PolarisReport struct {
	Version string
	Payload validator.AuditData
}

// ReportAuditToFairwindsInsights report audit to insights
// 1 - check if cluster exists, otherwise create it
// 2 - send workload report
// 3 - send polaris report
// 4 - checks if report job is completed for 3 minutes
// 5 - display link to Fairwinds Insights
func (ir insightsReporter) ReportAuditToFairwindsInsights(clusterName string, wr WorkloadsReport, pr PolarisReport) error {
	cluster, err := ir.client.upsertCluster(clusterName)
	if err != nil {
		return err
	}
	workloadsPayload, err := json.MarshalIndent(wr.Payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}
	_, err = ir.client.sendReport(*cluster, "workloads", wr.Version, workloadsPayload)
	if err != nil {
		return err
	}

	polarisPayload, err := json.MarshalIndent(pr.Payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}
	reportJob, err := ir.client.sendReport(*cluster, "polaris", pr.Version, polarisPayload)
	if err != nil {
		return err
	}

	success, err := verifyReportJobCompletion(&ir, cluster.Name, reportJob.ID)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("timed out waiting for report job to complete")
	}
	return nil
}

// verifyReportJobCompletion checks Insights for reportJob completion (timeout after 3 minutes)
func verifyReportJobCompletion(ir *insightsReporter, clusterName string, reportJobID int) (bool, error) {
	defer func() { fmt.Println() }()
	logrus.Println("Processing (this usually takes 1-3 minutes)...")
	for i := 0; i < 60; i++ {
		reportJob, err := ir.client.getReportJob(clusterName, reportJobID)
		if err != nil {
			return false, err
		}
		if reportJob.Status == "completed" {
			return true, nil
		}
		fmt.Print(".")
		time.Sleep(3 * time.Second)
	}
	return false, nil
}
