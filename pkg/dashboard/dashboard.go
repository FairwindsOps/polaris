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
	conf "github.com/reactiveops/polaris/pkg/config"
	"github.com/reactiveops/polaris/pkg/kube"
	"github.com/reactiveops/polaris/pkg/validator"
	"github.com/sirupsen/logrus"
	"gitlab.com/golang-commonmark/markdown"
)

const (
	// MainTemplateName is the main template
	MainTemplateName = "main.gohtml"
	// HeadTemplateName contains styles and meta info
	HeadTemplateName = "head.gohtml"
	// NavbarTemplateName contains the navbar
	NavbarTemplateName = "navbar.gohtml"
	// PreambleTemplateName contains an empty preamble that can be overridden
	PreambleTemplateName = "preamble.gohtml"
	// DashboardTemplateName contains the content of the dashboard
	DashboardTemplateName = "dashboard.gohtml"
	// FooterTemplateName contains the footer
	FooterTemplateName = "footer.gohtml"
	// CheckDetailsTemplateName is a page for rendering details about a given check
	CheckDetailsTemplateName = "check-details.gohtml"
)

var (
	templateBox = (*packr.Box)(nil)
	assetBox    = (*packr.Box)(nil)
	markdownBox = (*packr.Box)(nil)
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

// GetMarkdownBox returns a binary-friendly set of markdown files with error details
func GetMarkdownBox() *packr.Box {
	if markdownBox == (*packr.Box)(nil) {
		markdownBox = packr.New("Markdown", "../../docs")
	}
	return markdownBox
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
		"getIcon":         getIcon,
		"getCategoryLink": getCategoryLink,
		"getCategoryInfo": getCategoryInfo,
	})

	templateFileNames := []string{
		DashboardTemplateName,
		HeadTemplateName,
		NavbarTemplateName,
		PreambleTemplateName,
		FooterTemplateName,
		MainTemplateName,
	}
	return parseTemplateFiles(tmpl, templateFileNames)
}

func parseTemplateFiles(tmpl *template.Template, templateFileNames []string) (*template.Template, error) {
	templateBox := GetTemplateBox()
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

func writeTemplate(tmpl *template.Template, data *TemplateData, w http.ResponseWriter) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
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
	writeTemplate(tmpl, &templateData, w)
}

// EndpointHandler gets template data and renders json with it.
func EndpointHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration, kubeResources *kube.ResourceProvider) {
	templateData, err := validator.RunAudit(c, kubeResources)
	if err != nil {
		http.Error(w, "Error Fetching Deployments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(templateData)
}

// DetailsHandler returns details for a given error type
func DetailsHandler(w http.ResponseWriter, r *http.Request, category string) {
	box := GetMarkdownBox()
	contents, err := box.Find(category + ".md")
	if err != nil {
		http.Error(w, "Error details not found for category "+category, http.StatusNotFound)
		return
	}
	md := markdown.New(markdown.XHTMLOutput(true))
	detailsHTML := "{{ define \"details\" }}" + md.RenderToString(contents) + "{{ end }}"

	templateFileNames := []string{
		HeadTemplateName,
		NavbarTemplateName,
		CheckDetailsTemplateName,
		FooterTemplateName,
	}
	tmpl := template.New("check-details")
	tmpl, err = parseTemplateFiles(tmpl, templateFileNames)
	if err != nil {
		logrus.Printf("Error getting template data %v", err)
		http.Error(w, "Error getting template data", 500)
		return
	}
	tmpl.Parse(detailsHTML)
	writeTemplate(tmpl, nil, w)
}
