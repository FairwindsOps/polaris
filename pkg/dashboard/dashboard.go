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

type DashboardData struct {
	ClusterSummary    *validator.ResultSummary
	NamespacedResults validator.NamespacedResults
}

var tmpl = template.Must(template.ParseFiles("pkg/dashboard/templates/dashboard.gohtml"))

func Render(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	dashboardData, err := getDashboardData(c)
	if err != nil {
		http.Error(w, "Error Fetching Deploys", 500)
		return
	}

	tmpl.Execute(w, dashboardData)
}

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

	var clusterSuccesses uint
	var clusterFailures uint
	var clusterWarnings uint

	// Aggregate all summary counts to get a clusterwide count.
	for _, nsRes := range nsResults {
		for _, rr := range nsRes.Results {
			clusterSuccesses += rr.Summary.Successes
			clusterFailures += rr.Summary.Failures
			clusterWarnings += rr.Summary.Warnings
		}
	}

	dashboardData := DashboardData{
		ClusterSummary: &validator.ResultSummary{
			Successes: clusterSuccesses,
			Warnings:  clusterWarnings,
			Failures:  clusterFailures,
		},
		NamespacedResults: nsResults,
	}

	return dashboardData, nil
}
