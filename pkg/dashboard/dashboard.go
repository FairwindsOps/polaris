package dashboard

import (
	"encoding/json"
	"html/template"
	"net/http"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/validator"
)

// TemplateData represents data in a format that's template friendly.
type TemplateData struct {
	ClusterSummary    *validator.ResultSummary
	NamespacedResults validator.NamespacedResults
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

	// TODO: Once we are validating more than deployments,
	// we will need to merge the namespaceResults that get returned
	// from each validation.
	nsResults, err := validator.ValidateDeploys(c)
	if err != nil {
		return TemplateData{}, err
	}

	var clusterSuccesses, clusterFailures, clusterWarnings uint

	// Aggregate all summary counts to get a clusterwide count.
	for _, nsRes := range nsResults {
		for _, rr := range nsRes.Results {
			clusterFailures += rr.Summary.Failures
			clusterWarnings += rr.Summary.Warnings
			clusterSuccesses += rr.Summary.Successes
		}
	}

	templateData := TemplateData{
		ClusterSummary: &validator.ResultSummary{
			Failures:  clusterFailures,
			Warnings:  clusterWarnings,
			Successes: clusterSuccesses,
		},
		NamespacedResults: nsResults,
	}

	return templateData, nil
}
