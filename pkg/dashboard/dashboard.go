// Copyright 2019 ReactiveOps
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		"getWarningWidth": getWarningWidth,
		"getSuccessWidth": getSuccessWidth,
		"getWeatherIcon":  getWeatherIcon,
		"getWeatherText":  getWeatherText,
		"getGrade":        getGrade,
		"getScore":        getScore,
		"getIcon":         getIcon,
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
