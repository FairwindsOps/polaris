package dashboard

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// TemplateData is passed to the dashboard HTML template
type TemplateData struct {
	AuditData validator.AuditData
	JSON      template.JS
}

// MainHandler gets template data and renders the dashboard with it.
func MainHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration, kubeAPI *kube.API) {
	auditData, err := validator.RunAudit(c, kubeAPI)
	if err != nil {
		fmt.Printf("Error getting audit data %v \n", err)
		http.Error(w, "Error running audit", 500)
		return
	}
	jsonData, err := json.Marshal(auditData)
	if err != nil {
		http.Error(w, "Error serializing audit data", 500)
		return
	}
	templateData := TemplateData{
		AuditData: auditData,
		JSON:      template.JS(jsonData),
	}
	tmpl, err := template.New(TemplateName).Funcs(template.FuncMap{
		"getWarningWidth": func(rs validator.ResultSummary, fullWidth int) uint {
			return uint(float64(rs.Successes+rs.Warnings) / float64(rs.Successes+rs.Warnings+rs.Errors) * float64(fullWidth))
		},
		"getSuccessWidth": func(rs validator.ResultSummary, fullWidth int) uint {
			return uint(float64(rs.Successes) / float64(rs.Successes+rs.Warnings+rs.Errors) * float64(fullWidth))
		},
		"getGrade": func(rs validator.ResultSummary) string {
			total := (rs.Successes * 2) + rs.Warnings + (rs.Errors * 2)
			score := uint((float64(rs.Successes*2) / float64(total)) * 100)
			if score >= 97 {
				return "A+"
			} else if score >= 93 {
				return "A"
			} else if score >= 90 {
				return "A-"
			} else if score >= 87 {
				return "B+"
			} else if score >= 83 {
				return "B"
			} else if score >= 80 {
				return "B-"
			} else if score >= 77 {
				return "C+"
			} else if score >= 73 {
				return "C"
			} else if score >= 70 {
				return "C-"
			} else if score >= 67 {
				return "D+"
			} else if score >= 63 {
				return "D"
			} else if score >= 60 {
				return "D-"
			} else {
				return "F"
			}
		},
		"getScore": func(rs validator.ResultSummary) uint {
			total := (rs.Successes * 2) + rs.Warnings + (rs.Errors * 2)
			return uint((float64(rs.Successes*2) / float64(total)) * 100)
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

	buf := &bytes.Buffer{}
	err = template.Must(tmpl.Clone()).Execute(buf, templateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		buf.WriteTo(w)
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
