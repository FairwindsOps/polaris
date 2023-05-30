package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	workloads "github.com/fairwindsops/insights-plugins/plugins/workloads/pkg"
	"github.com/fairwindsops/polaris/pkg/auth"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/sirupsen/logrus"
)

type insightsReporter struct {
	insightsHost string
	auth         auth.Host
}

func NewInsightsReporter(insightsHost string, auth auth.Host) *insightsReporter {
	return &insightsReporter{
		insightsHost: insightsHost,
		auth:         auth,
	}
}

type insightsCluster struct {
	Name         string `json:"Name"`
	AuthToken    string `json:"AuthToken"`
	Organization string `json:"Organization"`
	Status       string `json:"Status"`
}

type insightsReportJob struct {
	ID            int    `json:"id"`
	Status        string `json:"status"`
	TimeTakenInMs int    `json:"timeTaken"`
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
	cluster, err := ir.upsertCluster(clusterName)
	if err != nil {
		return err
	}
	logrus.Infof("Uploading to Fairwinds Insights organization '%s/%s'...", ir.auth.Organization, clusterName)

	workloadsPayload, err := json.MarshalIndent(wr.Payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}
	_, err = ir.sendReport(cluster, "workloads", wr.Version, workloadsPayload)
	if err != nil {
		return err
	}

	polarisPayload, err := json.MarshalIndent(pr.Payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}
	reportJob, err := ir.sendReport(cluster, "polaris", pr.Version, polarisPayload)
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
	logrus.Println("Success! You can see your results at:")
	logrus.Printf("%s/orgs/%s/clusters/%s\n", ir.insightsHost, ir.auth.Organization, cluster.Name)
	return nil
}

// verifyReportJobCompletion checks Insights for reportJob completion (timeout after 3 minutes)
func verifyReportJobCompletion(ir *insightsReporter, clusterName string, reportJobID int) (bool, error) {
	defer func() { fmt.Println() }()
	logrus.Println("Processing (this usually takes 1-3 minutes)...")
	for i := 0; i < 60; i++ {
		reportJob, err := ir.getReportJob(clusterName, reportJobID)
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

func (ir insightsReporter) upsertCluster(clusterName string) (*insightsCluster, error) {
	clusterURL := fmt.Sprintf("%s/v0/organizations/%s/clusters/%s?showToken=true", ir.insightsHost, ir.auth.Organization, clusterName)
	req, err := http.NewRequest("GET", clusterURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for fetching cluster: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ir.auth.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request fetching cluster: %w", err)
	}
	defer resp.Body.Close()

	if isSuccessful2XX(resp.StatusCode) {
		// cluster already created
		logrus.Infof("cluster %q found...", clusterName)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}
		var c insightsCluster
		err = json.Unmarshal(body, &c)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling response body: %w", err)
		}
		return &c, nil
	} else {
		logrus.Warnf("not able to fetch cluster, expected 2xx - received %d, will try to create", resp.StatusCode)
	}

	logrus.Infof("cluster %q not found... creating..", clusterName)

	req, err = http.NewRequest("POST", clusterURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for creating cluster: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ir.auth.Token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request for creating cluster: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if !isSuccessful2XX(resp.StatusCode) {
		return nil, fmt.Errorf("creating cluster, expected 2xx received %d: %v", resp.StatusCode, string(body))
	}

	var c insightsCluster
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling response body: %w", err)
	}

	logrus.Infof("cluster %q created...", clusterName)
	return &c, nil
}

func (ir insightsReporter) sendReport(cluster *insightsCluster, reportType, reportVersion string, payload []byte) (*insightsReportJob, error) {
	uploadReportURL := fmt.Sprintf("%s/v0/organizations/%s/clusters/%s/data/%s", ir.insightsHost, ir.auth.Organization, cluster.Name, reportType)
	req, err := http.NewRequest("POST", uploadReportURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("building request for output: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cluster.AuthToken)
	req.Header.Set("X-Fairwinds-Report-Version", reportVersion)
	req.Header.Set("X-Fairwinds-Report-Priority", "4") // should have higher priority than the default 5

	// TODO: Vitor - ?!
	req.Header.Set("X-Fairwinds-Agent-Version", "")
	req.Header.Set("X-Fairwinds-Agent-Chart-Version", "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request for output: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	if !isSuccessful2XX(resp.StatusCode) {
		return nil, fmt.Errorf("sending %s report, expected 2xx received %d: %v", reportType, resp.StatusCode, string(body))
	}

	var rj insightsReportJob
	err = json.Unmarshal(body, &rj)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling response body: %w", err)
	}
	logrus.Debugf("%s report sent to fairwinds insights", reportType)

	return &rj, nil
}

func isSuccessful2XX(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func (ir insightsReporter) getReportJob(clusterName string, reportJobID int) (*insightsReportJob, error) {
	reportJobsURL := fmt.Sprintf("%s/v0/organizations/%s/clusters/%s/report-jobs/%d", ir.insightsHost, ir.auth.Organization, clusterName, reportJobID)
	req, err := http.NewRequest("GET", reportJobsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for fetching report-job: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ir.auth.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request fetching report-job: %w", err)
	}
	defer resp.Body.Close()

	if !isSuccessful2XX(resp.StatusCode) {
		return nil, fmt.Errorf("fetching report-job, expected 2xx received %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	var rj insightsReportJob
	err = json.Unmarshal(body, &rj)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling response body: %w", err)
	}
	return &rj, nil
}
