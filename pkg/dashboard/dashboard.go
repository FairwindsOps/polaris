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
	NamespacedResults []validator.NamespacedResult
}

var tmpl = template.Must(template.ParseFiles("pkg/dashboard/templates/charts.gohtml"))

func Render(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	dashboardData := getDashboardData()
	tmpl.Execute(w, dashboardData)
}

func RenderJSON(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	results := []validator.Results{}
	pods, err := kube.CoreV1API.Pods("").List(metav1.ListOptions{})
	if err != nil {
		http.Error(w, "Error Fetching Pods", 500)
		return
	}
	log.Println("pods count:", len(pods.Items))
	for _, pod := range pods.Items {
		result := validator.ValidatePods(c, &pod.Spec, validator.Results{})
		results = append(results, result)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func getDashboardData() DashboardData {
	return DashboardData{
		ClusterSummary: &validator.ResultSummary{
			Successes: 46,
			Warnings:  8,
			Failures:  5,
		},
		NamespacedResults: []validator.NamespacedResult{{
			Namespace: "kube-system",
			Results: []validator.ResourceResult{{
				Name: "tiller",
				Type: "Deployment",
				Summary: &validator.ResultSummary{
					Successes: 7,
					Warnings:  3,
					Failures:  2,
				},
				Messages: []validator.ResultMessage{{
					Message: "Image Tag Specified",
					Type:    "success",
				}, {
					Message: "Liveness Probe Specified",
					Type:    "success",
				}, {
					Message: "Readiness Probe Specified",
					Type:    "success",
				}, {
					Message: "Container Running As Root",
					Type:    "warning",
				}, {
					Message: "Resource requests are not set",
					Type:    "failure",
				}},
			}},
		}},
	}
}
