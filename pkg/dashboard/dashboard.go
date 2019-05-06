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
	"html/template"
	"net/http"

	packr "github.com/gobuffalo/packr/v2"
	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
	"github.com/reactiveops/fairwinds/pkg/validator"
	"github.com/sirupsen/logrus"
)

const (
	// MainTemplateName is the main template
	MainTemplateName = "main.gohtml"
	// HeaderTemplateName contains the navbar
	HeaderTemplateName = "header.gohtml"
	// PreambleTemplateName contains an empty preamble that can be overridden
	PreambleTemplateName = "preamble.gohtml"
	// DashboardTemplateName contains the content of the dashboard
	DashboardTemplateName = "dashboard.gohtml"
	// FooterTemplateName contains the footer
	FooterTemplateName = "footer.gohtml"
)

var (
	templateBox = (*packr.Box)(nil)
	assetBox    = (*packr.Box)(nil)
)

// GetAssetBox returns a binary-friendly set of assets packaged from disk
func GetAssetBox() *packr.Box {
	if assetBox == (*packr.Box)(nil) {
		assetBox = packr.New("Assets", "assets")
	}
	return assetBox
}

// GetTemplateBox returns a binary-friendly set of templates for rendering the dash
func GetTemplateBox() *packr.Box {
	if templateBox == (*packr.Box)(nil) {
		templateBox = packr.New("Templates", "templates")
	}
	return templateBox
}

// TemplateData is passed to the dashboard HTML template
type TemplateData struct {
	AuditData validator.AuditData
	JSON      template.JS
}

// GetBaseTemplate puts together the dashboard template. Individual pieces can be overridden before rendering.
func GetBaseTemplate(name string) (*template.Template, error) {
	tmpl := template.New(name).Funcs(template.FuncMap{
		"getWarningWidth": getWarningWidth,
		"getSuccessWidth": getSuccessWidth,
		"getWeatherIcon":  getWeatherIcon,
		"getWeatherText":  getWeatherText,
		"getGrade":        getGrade,
		"getScore":        getScore,
		"getIcon":         getIcon,
	})

	templateBox := GetTemplateBox()
	templateFileNames := []string{
		DashboardTemplateName,
		HeaderTemplateName,
		PreambleTemplateName,
		FooterTemplateName,
		MainTemplateName,
	}
	for _, fname := range templateFileNames {
		templateFile, err := templateBox.Find(fname)
		if err != nil {
			return nil, err
		}

		tmpl, err = tmpl.Parse(string(templateFile))
		if err != nil {
			return nil, err
		}
	}
	return tmpl, nil
}

// MainHandler gets template data and renders the dashboard with it.
func MainHandler(w http.ResponseWriter, r *http.Request, auditData validator.AuditData) {
	jsonData, err := json.Marshal(auditData)

	if err != nil {
		http.Error(w, "Error serializing audit data", 500)
		return
	}

	templateData := TemplateData{
		AuditData: auditData,
		JSON:      template.JS(jsonData),
	}
	tmpl, err := GetBaseTemplate("main")
	if err != nil {
		logrus.Printf("Error getting template data %v", err)
		http.Error(w, "Error getting template data", 500)
		return
	}

	buf := &bytes.Buffer{}
	err = tmpl.ExecuteTemplate(buf, "main", templateData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
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
