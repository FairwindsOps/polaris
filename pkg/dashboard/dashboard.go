package dashboard

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
	"github.com/reactiveops/fairwinds/pkg/validator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DashboardData stores validation results organized by namespace and also
// tracks the total cluster count of failed/successed validation checks.
type DashboardData struct {
	ClusterSummary    *validator.ResultSummary
	NamespacedResults validator.NamespacedResults
}

var tmpl = template.Must(template.ParseFiles("pkg/dashboard/templates/dashboard.gohtml"))

// Render populates the dashboard template with validation data.
func Render(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	dashboardData, err := getDashboardData(c)
	if err != nil {
		http.Error(w, "Error Fetching Deploys", 500)
		return
	}

	tmpl.Execute(w, dashboardData)
}

// RenderJSON returns pod validation data in JSON format.
func RenderJSON(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	var clientset = kube.CreateClientset()
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		http.Error(w, "Error Fetching Pods", 500)
		return
	}
	log.Println("pods count:", len(pods.Items))
	res := []validator.ResourceResult{}
	for _, pod := range pods.Items {
		resResult := validator.ValidatePod(c, &pod.Spec)
		resResult.Name = pod.Name
		res = append(res, resResult)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func getDashboardData(c conf.Configuration) (DashboardData, error) {
	// TODO: Once we are validating more than deployments,
	// we will need to merge the namespaceResults that get returned
	// from each validation.
	nsResults, _ := validator.ValidateDeploys(c)

	var clusterSuccesses, clusterFailures, clusterWarnings uint

	// Aggregate all summary counts to get a clusterwide count.
	for _, nsRes := range nsResults {
		for _, rr := range nsRes.Results {
			clusterFailures += rr.Summary.Failures
			clusterWarnings += rr.Summary.Warnings
			clusterSuccesses += rr.Summary.Successes
		}
	}

	dashboardData := DashboardData{
		ClusterSummary: &validator.ResultSummary{
			Failures:  clusterFailures,
			Warnings:  clusterWarnings,
			Successes: clusterSuccesses,
		},
		NamespacedResults: nsResults,
	}

	return dashboardData, nil
}
