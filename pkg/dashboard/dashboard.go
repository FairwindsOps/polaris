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
	results := []validator.Results{}
	var clientset = kube.CreateClientset()
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
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
	var clientset = kube.CreateClientset()
	deploys, err := clientset.AppsV1().Deployments("").List(metav1.ListOptions{})
	if err != nil {
		return DashboardData{}, err
	}

	dashboardData := DashboardData{
		ClusterSummary: &validator.ResultSummary{
			Successes: 0,
			Warnings:  4,
			Failures:  0,
		},
		NamespacedResults: map[string]*validator.NamespacedResult{},
	}

	for _, deploy := range deploys.Items {
		validationFailures := validator.ValidateDeploys(c, &deploy)
		deployResult := validator.ResourceResult{
			Name: deploy.Name,
			Type: "Deployment",
			Summary: &validator.ResultSummary{
				Successes: 0,
				Warnings:  2,
				Failures:  0,
			},
			ContainerResults: []validator.ContainerResult{},
		}

		if dashboardData.NamespacedResults[deploy.Namespace] == nil {
			dashboardData.NamespacedResults[deploy.Namespace] = &validator.NamespacedResult{
				Results: []validator.ResourceResult{},
				Summary: &validator.ResultSummary{
					Successes: 0,
					Warnings:  3,
					Failures:  0,
				},
			}
		}

		containerValidations := append(validationFailures.InitContainerValidations, validationFailures.ContainerValidations...)
		for _, containerValidation := range containerValidations {
			containerResult := validator.ContainerResult{
				Name:     containerValidation.Container.Name,
				Messages: []validator.ResultMessage{},
			}
			for _, success := range containerValidation.Successes {
				dashboardData.ClusterSummary.Successes++
				dashboardData.NamespacedResults[deploy.Namespace].Summary.Successes++
				deployResult.Summary.Successes++
				containerResult.Messages = append(containerResult.Messages, success)
			}
			for _, failure := range containerValidation.Failures {
				dashboardData.ClusterSummary.Failures++
				dashboardData.NamespacedResults[deploy.Namespace].Summary.Failures++
				deployResult.Summary.Failures++
				containerResult.Messages = append(containerResult.Messages, failure)
			}
			deployResult.ContainerResults = append(deployResult.ContainerResults, containerResult)
		}

		dashboardData.NamespacedResults[deploy.Namespace].Results = append(
			dashboardData.NamespacedResults[deploy.Namespace].Results, deployResult)
	}

	return dashboardData, nil
}
