package dashboard

import (
	"encoding/json"
	"html/template"
	"net/http"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
	"github.com/reactiveops/fairwinds/pkg/validator"
)

const (
	// TemplateName references the dashboard template to use
	TemplateName = "dashboard.gohtml"

	// TemplateFile references the path of the dashboard template to use
	TemplateFile = "pkg/dashboard/templates/" + TemplateName
)

// TemplateData represents data in a format that's template friendly.
type TemplateData struct {
	ClusterSummary    *validator.ResultSummary
	NamespacedResults validator.NamespacedResults
}

// MainHandler gets template data and renders the dashboard with it.
func MainHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration, kubeAPI *kube.API) {
	templateData, err := getTemplateData(c, kubeAPI)
	if err != nil {
		http.Error(w, "Error Fetching Deployments", 500)
		return
	}
	tmpl, err := template.New(TemplateName).Funcs(template.FuncMap{
		"getWarningWidth": func(rs validator.ResultSummary, fullWidth int) uint {
			return uint(float64(rs.Successes+rs.Warnings) / float64(rs.Successes+rs.Warnings+rs.Errors) * float64(fullWidth))
		},
		"getSuccessWidth": func(rs validator.ResultSummary, fullWidth int) uint {
			return uint(float64(rs.Successes) / float64(rs.Successes+rs.Warnings+rs.Errors) * float64(fullWidth))
		},
		"getIcon": func(rm validator.ResultMessage) string {
			switch rm.Type {
			case "success":
				return "fas fa-check"
			case "warning":
				return "fas fa-exclamation"
			default:
				return "fas fa-times"
			}
		},
	}).ParseFiles(TemplateFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = template.Must(tmpl.Clone()).Execute(w, templateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// EndpointHandler gets template data and renders json with it.
func EndpointHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration, kubeAPI *kube.API) {
	templateData, err := getTemplateData(c, kubeAPI)
	if err != nil {
		http.Error(w, "Error Fetching Deployments", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templateData)
}

func getTemplateData(config conf.Configuration, kubeAPI *kube.API) (TemplateData, error) {

	// TODO: Once we are validating more than deployments,
	// we will need to merge the namespaceResults that get returned
	// from each validation.
	nsResults, err := validator.ValidateDeploys(config, kubeAPI)
	if err != nil {
		return TemplateData{}, err
	}

	var clusterSuccesses, clusterErrors, clusterWarnings uint

	// Aggregate all summary counts to get a clusterwide count.
	for _, nsRes := range nsResults {
		for _, rr := range nsRes.Results {
			clusterErrors += rr.Summary.Errors
			clusterWarnings += rr.Summary.Warnings
			clusterSuccesses += rr.Summary.Successes
		}
	}

	templateData := TemplateData{
		ClusterSummary: &validator.ResultSummary{
			Errors:    clusterErrors,
			Warnings:  clusterWarnings,
			Successes: clusterSuccesses,
		},
		NamespacedResults: nsResults,
	}

	return templateData, nil
}
