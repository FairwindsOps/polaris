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

// MainHandler gets template data and renders the dashboard with it.
func MainHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration, kubeAPI *kube.API) {
	templateData, err := validator.RunAudit(c, kubeAPI)
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
	templateData, err := validator.RunAudit(c, kubeAPI)
	if err != nil {
		http.Error(w, "Error Fetching Deployments", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templateData)
}
