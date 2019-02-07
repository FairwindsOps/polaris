package dashboard

import (
	"encoding/json"
	"html/template"
	"net/http"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
	"github.com/reactiveops/fairwinds/pkg/validator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TemplateData represents data in a format that's template friendly.
type TemplateData struct {
	ClusterSummary    *validator.ResultSummary
	NamespacedResults map[string]*validator.NamespacedResult
}

var tmpl = template.Must(template.ParseFiles("pkg/dashboard/templates/dashboard.gohtml"))

// MainHandler gets template data and renders the dashboard with it.
func MainHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	templateData, err := getTemplateData(c)
	if err != nil {
		http.Error(w, "Error Fetching Deployments", 500)
		return
	}

	tmpl.Execute(w, templateData)
}

// EndpointHandler gets template data and renders json with it.
func EndpointHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	templateData, err := getTemplateData(c)
	if err != nil {
		http.Error(w, "Error Fetching Deployments", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templateData)
}

func getTemplateData(c conf.Configuration) (TemplateData, error) {
	var clientset = kube.CreateClientset()
	deploys, err := clientset.AppsV1().Deployments("").List(metav1.ListOptions{})
	if err != nil {
		return TemplateData{}, err
	}

	templateData := TemplateData{
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

		if templateData.NamespacedResults[deploy.Namespace] == nil {
			templateData.NamespacedResults[deploy.Namespace] = &validator.NamespacedResult{
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
				templateData.ClusterSummary.Successes++
				templateData.NamespacedResults[deploy.Namespace].Summary.Successes++
				deployResult.Summary.Successes++
				containerResult.Messages = append(containerResult.Messages, success)
			}
			for _, failure := range containerValidation.Failures {
				templateData.ClusterSummary.Failures++
				templateData.NamespacedResults[deploy.Namespace].Summary.Failures++
				deployResult.Summary.Failures++
				containerResult.Messages = append(containerResult.Messages, failure)
			}
			deployResult.ContainerResults = append(deployResult.ContainerResults, containerResult)
		}

		templateData.NamespacedResults[deploy.Namespace].Results = append(
			templateData.NamespacedResults[deploy.Namespace].Results, deployResult)
	}

	return templateData, nil
}
