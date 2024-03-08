package fix

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/mutation"
	"github.com/fairwindsops/polaris/pkg/validator"
)

const templateLineMarker = "# POLARIS_FIX_TMPL"
const templateOpenMarker = "POLARIS_OPEN_TMPL"
const templateCloseMarker = "POLARIS_CLOSE_TMPL"

var ErrFilesPathRequired = errors.New("files-path flag is required")

func Execute(config config.Configuration, filesPath string, isTemplate bool, checksToFix ...string) error {
	if filesPath == "" {
		return ErrFilesPathRequired
	}

	var yamlFiles []string
	fileInfo, err := os.Stat(filesPath)
	if err != nil {
		return fmt.Errorf("error getting file info: %v", err)
	}
	if fileInfo.IsDir() {
		baseDir := filesPath
		if !strings.HasSuffix(filesPath, "/") {
			baseDir = baseDir + "/"
		}
		yamlFiles, err = getYamlFiles(baseDir)
		if err != nil {
			return fmt.Errorf("error getting yaml files from directory: %v", err)
		}
	} else {
		yamlFiles = append(yamlFiles, filesPath)
	}

	if len(checksToFix) > 0 {
		if len(checksToFix) == 1 && checksToFix[0] == "all" {
			allchecks := []string{}
			for key := range config.Checks {
				allchecks = append(allchecks, key)
			}
			config.Mutations = allchecks
		} else if len(checksToFix) == 0 && checksToFix[0] == "none" {
			config.Mutations = nil
		} else {
			config.Mutations = checksToFix
		}
	}

	for _, fullFilePath := range yamlFiles {
		yamlContent, err := os.ReadFile(fullFilePath)
		if err != nil {
			return fmt.Errorf("error reading file with file path %s: %v", fullFilePath, err)
		}

		if isTemplate {
			yamlContent = []byte(detemplate(string(yamlContent)))
		}
		kubeResources, err := kube.CreateResourceProviderFromYaml(string(yamlContent))
		if err != nil {
			return fmt.Errorf("error creating resource provider from yaml: %v", err)
		}
		results, err := validator.ApplyAllSchemaChecksToResourceProvider(&config, kubeResources)
		if err != nil {
			return fmt.Errorf("error applying schema check to the resources %s: %v", fullFilePath, err)
		}
		allMutations := mutation.GetMutationsFromResults(results)

		updatedYamlContent := ""
		if len(allMutations) > 0 {
			for _, resources := range kubeResources.Resources {
				for _, resource := range resources {
					key := fmt.Sprintf("%s/%s/%s", resource.Kind, resource.Resource.GetName(), resource.Resource.GetNamespace())
					mutations := allMutations[key]
					mutatedYamlContent, err := mutation.ApplyAllMutations(string(resource.OriginalObjectYAML), mutations)
					if err != nil {
						return fmt.Errorf("error applying schema mutations to the resource %s: %v", key, err)
					}
					if updatedYamlContent != "" {
						updatedYamlContent += "\n---\n"
					}
					updatedYamlContent += mutatedYamlContent
				}
			}
		}

		if isTemplate {
			updatedYamlContent = retemplate(updatedYamlContent)
		}

		if updatedYamlContent != "" {
			err = os.WriteFile(fullFilePath, []byte(updatedYamlContent), 0644)
			if err != nil {
				return fmt.Errorf("error writing output to file: %v", err)
			}
		}
	}

	return nil
}

func detemplate(content string) string {
	lines := strings.Split(content, "\n")
	for idx, line := range lines {
		lines[idx] = detemplateLine(line)
	}
	return strings.Join(lines, "\n")
}

func retemplate(content string) string {
	lines := strings.Split(content, "\n")
	for idx, line := range lines {
		lines[idx] = retemplateLine(line)
	}
	return strings.Join(lines, "\n")
}

func detemplateLine(line string) string {
	if !strings.HasPrefix(strings.TrimSpace(line), "{{") {
		line = strings.ReplaceAll(line, "{", templateOpenMarker)
		line = strings.ReplaceAll(line, "}", templateCloseMarker)
		return line
	}
	tmplStart := strings.Index(line, "{{")
	newLine := line[:tmplStart] + templateLineMarker + line[tmplStart:]
	return newLine
}

func retemplateLine(line string) string {
	if !strings.Contains(line, templateLineMarker) {
		line = strings.ReplaceAll(line, templateOpenMarker, "{")
		line = strings.ReplaceAll(line, templateCloseMarker, "}")
		return line
	}
	return strings.Replace(line, templateLineMarker, "", 1)
}

func getYamlFiles(rootpath string) ([]string, error) {
	var list []string
	err := filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			list = append(list, path)
		}
		return nil
	})
	return list, err
}
