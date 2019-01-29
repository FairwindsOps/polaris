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
	NamespacedResults map[string]*validator.NamespacedResult
}

var tmpl = template.Must(template.ParseFiles("pkg/dashboard/templates/charts.gohtml"))

func Render(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	dashboardData, err := getDashboardData(c)
	if err != nil {
		http.Error(w, "Error Fetching Deploys", 500)
		return
	}

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
		result := validator.ValidatePods(c, &pod.Spec)
		results = append(results, result)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func getDashboardData(c conf.Configuration) (DashboardData, error) {
	deploys, err := kube.AppsV1API.Deployments("").List(metav1.ListOptions{})
	if err != nil {
		return DashboardData{}, err
	}

	log.Println("deploys =======>", len(deploys.Items))

	dashboardData := DashboardData{
		ClusterSummary: &validator.ResultSummary{
			Successes: 46,
			Warnings:  8,
			Failures:  5,
		},
	}

	for _, deploy := range deploys.Items {
		validationFailures := validator.ValidateDeploys(c, &deploy)
		resResult := validator.ResourceResult{
			Name: deploy.Name,
			Type: "Deployment",
		}

		for _, containerValidation := range validationFailures.InitContainerValidations {
			for _, failure := range containerValidation.Failures {
				dashboardData.ClusterSummary.Failures++
				// rr := *resResult.Summary
				// rr.Failures++
				resResult.Messages = append(resResult.Messages, validator.ResultMessage{
					Message: failure.Reason(),
					Type:    "failure",
				})
			}
		}

		for _, containerValidation := range validationFailures.ContainerValidations {
			for _, failure := range containerValidation.Failures {
				dashboardData.ClusterSummary.Failures++
				// rr := *resResult.Summary
				// rr.Failures++
				resResult.Messages = append(resResult.Messages, validator.ResultMessage{
					Message: failure.Reason(),
					Type:    "failure",
				})
			}
		}

		log.Println("adding results to =======>", deploy.Namespace)
		rr := []validator.ResourceResult{}
		rr = append(rr, resResult)
		nsRes := validator.NamespacedResult{Results: rr}

		ns := map[string]*validator.NamespacedResult{
			deploy.Namespace: &nsRes,
		}
		dashboardData.NamespacedResults = ns

		// dashboardData.NamespacedResults[deploy.Namespace].Results = append(dashboardData.NamespacedResults[deploy.Namespace].Results, resResult)
		log.Println("adding results to =======>", len(dashboardData.NamespacedResults[deploy.Namespace].Results))
	}

	return dashboardData, nil

	// return DashboardData{
	// 	ClusterSummary: &validator.ResultSummary{
	// 		Successes: 46,
	// 		Warnings:  8,
	// 		Failures:  5,
	// 	},
	// 	NamespacedResults: {
	// 		"kube-system": {
	// 			Summary: &validator.ResultSummary{
	// 				Successes: 7,
	// 				Warnings:  3,
	// 				Failures:  2,
	// 			},
	// 			Results: []validator.ResourceResult{{
	// 				Name: "tiller",
	// 				Type: "Deployment",
	// 				Summary: &validator.ResultSummary{
	// 					Successes: 7,
	// 					Warnings:  3,
	// 					Failures:  2,
	// 				},
	// 				Messages: []validator.ResultMessage{{
	// 					Message: "Image Tag Specified",
	// 					Type:    "success",
	// 				}, {
	// 					Message: "Liveness Probe Specified",
	// 					Type:    "success",
	// 				}, {
	// 					Message: "Readiness Probe Specified",
	// 					Type:    "success",
	// 				}, {
	// 					Message: "Container Running As Root",
	// 					Type:    "warning",
	// 				}, {
	// 					Message: "Resource requests are not set",
	// 					Type:    "failure",
	// 				}},
	// 			}},
	// 		}
	// 	}},
	// }, nil
}
