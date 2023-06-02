package insights

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fairwindsops/polaris/pkg/auth"
	"github.com/sirupsen/logrus"
)

type cluster struct {
	Name         string `json:"Name"`
	AuthToken    string `json:"AuthToken"`
	Organization string `json:"Organization"`
	Status       string `json:"Status"`
}

type reportJob struct {
	ID            int    `json:"id"`
	Status        string `json:"status"`
	TimeTakenInMs int    `json:"timeTaken"`
}

type Client interface {
	upsertCluster(clusterName string) (*cluster, error)
	sendReport(cluster cluster, reportType, reportVersion string, payload []byte) (*reportJob, error)
	getReportJob(clusterName string, reportJobID int) (*reportJob, error)
}

type HTTPClient struct {
	insightsHost string
	auth         auth.Host
}

func NewHTTPClient(host string, auth auth.Host) Client {
	return HTTPClient{host, auth}
}

func (ic HTTPClient) upsertCluster(clusterName string) (*cluster, error) {
	clusterURL := fmt.Sprintf("%s/v0/organizations/%s/clusters/%s?showToken=true", ic.insightsHost, ic.auth.Organization, clusterName)
	req, err := http.NewRequest("GET", clusterURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for fetching cluster: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ic.auth.Token)
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
		var c cluster
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
	req.Header.Set("Authorization", "Bearer "+ic.auth.Token)
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
		return nil, fmt.Errorf("creating cluster, expected 200 OK received %s: %v", resp.Status, string(body))
	}

	var c cluster
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling response body: %w", err)
	}

	logrus.Infof("cluster %q created...", clusterName)
	return &c, nil
}

func (ic HTTPClient) sendReport(cluster cluster, reportType, reportVersion string, payload []byte) (*reportJob, error) {
	uploadReportURL := fmt.Sprintf("%s/v0/organizations/%s/clusters/%s/data/%s", ic.insightsHost, ic.auth.Organization, cluster.Name, reportType)
	req, err := http.NewRequest("POST", uploadReportURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("building request for output: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cluster.AuthToken)
	req.Header.Set("X-Fairwinds-Report-Version", reportVersion)
	req.Header.Set("X-Fairwinds-Report-Priority", "4") // should have higher priority than the default 5
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
		return nil, fmt.Errorf("sending %s report, expected 200 OK received %s: %v", reportType, resp.Status, string(body))
	}

	var rj reportJob
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

func (ic HTTPClient) getReportJob(clusterName string, reportJobID int) (*reportJob, error) {
	reportJobsURL := fmt.Sprintf("%s/v0/organizations/%s/clusters/%s/report-jobs/%d", ic.insightsHost, ic.auth.Organization, clusterName, reportJobID)
	req, err := http.NewRequest("GET", reportJobsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for fetching report-job: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ic.auth.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request fetching report-job: %w", err)
	}
	defer resp.Body.Close()

	if !isSuccessful2XX(resp.StatusCode) {
		return nil, fmt.Errorf("fetching report-job, expected 200 OK received %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	var rj reportJob
	err = json.Unmarshal(body, &rj)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling response body: %w", err)
	}
	return &rj, nil
}
