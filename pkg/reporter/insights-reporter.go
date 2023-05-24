package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	workloads "github.com/fairwindsops/insights-plugins/plugins/workloads/pkg"
	"github.com/fairwindsops/polaris/pkg/auth"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/sirupsen/logrus"
)

type InsightsReporter struct {
	insightsURL      string
	auth             auth.Host
	polarisVersion   string
	workloadsVersion string
}

func NewInsightsReporter(insightsURL string, auth auth.Host, polarisVersion, workloadsVersion string) *InsightsReporter {
	return &InsightsReporter{
		insightsURL:      insightsURL,
		auth:             auth,
		polarisVersion:   polarisVersion,
		workloadsVersion: workloadsVersion,
	}
}

type insightsCluster struct {
	Name         string `json:"Name" yaml:"Name"`
	AuthToken    string `json:"AuthToken" yaml:"AuthToken"`
	Organization string `json:"Organization" yaml:"Organization"`
	Status       string `json:"Status" yaml:"Status"`
}

// ReportAuditToFairwindsInsights report audit to insights
// 1 - check if cluster exists, otherwise create it
// 2 - send workload report
// 3 - send polaris report
// 4 - results?!
func (ir InsightsReporter) ReportAuditToFairwindsInsights(clusterName string, k8sResources workloads.ClusterWorkloadReport, auditData validator.AuditData) error {
	// 1
	cluster, err := ir.upsertCluster(clusterName)
	if err != nil {
		return err
	}
	// 2
	workloadsPayload, err := json.MarshalIndent(k8sResources, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}
	err = ir.sendReport(cluster, "workloads", ir.workloadsVersion, workloadsPayload)
	if err != nil {
		return err
	}
	// 3
	polarisPayload, err := json.MarshalIndent(auditData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}
	err = ir.sendReport(cluster, "polaris", ir.polarisVersion, polarisPayload)
	if err != nil {
		return err
	}
	return nil
}

func (ir InsightsReporter) upsertCluster(clusterName string) (*insightsCluster, error) {
	clusterURL := fmt.Sprintf("%s/v0/organizations/%s/clusters/%s?showToken=true", ir.insightsURL, ir.auth.Organization, clusterName)
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

func (ir InsightsReporter) sendReport(cluster *insightsCluster, reportType, reportVersion string, payload []byte) error {
	uploadReportURL := fmt.Sprintf("%s/v0/organizations/%s/clusters/%s/data/%s", ir.insightsURL, ir.auth.Organization, cluster.Name, reportType)
	req, err := http.NewRequest("POST", uploadReportURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("building request for output: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cluster.AuthToken)
	req.Header.Set("X-Fairwinds-Report-Version", reportVersion)

	// TODO: Vitor - ?!
	req.Header.Set("X-Fairwinds-Agent-Version", "")
	req.Header.Set("X-Fairwinds-Agent-Chart-Version", "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("making request for output: %w", err)
	}
	defer resp.Body.Close()
	if !isSuccessful2XX(resp.StatusCode) {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("sending %s report, expected 2xx received %d: %v", reportType, resp.StatusCode, string(body))
	}
	logrus.Infof("%s report sent to fairwinds insights", reportType)
	return nil
}

func isSuccessful2XX(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
