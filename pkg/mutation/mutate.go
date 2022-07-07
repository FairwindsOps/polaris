package mutation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	jsonpatchV5 "github.com/evanphx/json-patch/v5"
	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/thoas/go-funk"
	"gomodules.xyz/jsonpatch/v2"
)

// ApplyAllSchemaMutations applies available mutation to a single resource
func ApplyAllSchemaMutations(conf *config.Configuration, resourceProvider *kube.ResourceProvider, resource kube.GenericResource, mutations []jsonpatch.Operation) (kube.GenericResource, error) {
	resByte := resource.OriginalObjectJSON
	var jsonByte []byte
	mutationByte, err := json.Marshal(mutations)
	if err != nil {
		return resource, err
	}

	patch, err := jsonpatchV5.DecodePatch(mutationByte)
	if err != nil {
		return resource, err
	}
	jsonByte, err = patch.ApplyWithOptions(resByte, &jsonpatchV5.ApplyOptions{
		AllowMissingPathOnRemove: true,
		EnsurePathExistsOnAdd: true,
	})
	if err != nil {
		return resource, err
	}
	mutated, err := kube.NewGenericResourceFromBytes(jsonByte)
	if err != nil {
		return resource, err
	}

	return mutated, nil
}

// GetMutationsAndCommentsFromResults returns all mutations from results
func GetMutationsAndCommentsFromResults(results []validator.Result) ([]config.MutationComment, map[string][]jsonpatch.Operation) {
	allMutationsFromResults := make(map[string][]jsonpatch.Operation)
	comments := []config.MutationComment{}
	for _, result := range results {
		key := fmt.Sprintf("%s/%s/%s", result.Kind, result.Name, result.Namespace)

		mutations, resultsComments := GetMutationsAndCommentsFromResult(&result)
		allMutationsFromResults[key] = mutations
		comments = append(comments, resultsComments...)

	}
	return comments, allMutationsFromResults
}

// GetMutationsAndCommentsFromResult returns all mutations from single result
func GetMutationsAndCommentsFromResult(result *validator.Result) ([]jsonpatch.Operation, []config.MutationComment) {
	mutations := []jsonpatch.Operation{}
	comments := []config.MutationComment{}
	for _, resultMessage := range result.Results {
		if len(resultMessage.Mutations) > 0 {
			mutations = append(mutations, resultMessage.Mutations...)
		}
		if len(resultMessage.Comments) > 0 {
			comments = append(comments, resultMessage.Comments...)
		}
	}

	for _, resultMessage := range result.PodResult.Results {
		if len(resultMessage.Mutations) > 0 {
			mutations = append(mutations, resultMessage.Mutations...)
		}
		if len(resultMessage.Comments) > 0 {
			comments = append(comments, resultMessage.Comments...)
		}
	}

	for _, containerResult := range result.PodResult.ContainerResults {
		for _, resultMessage := range containerResult.Results {
			if len(resultMessage.Mutations) > 0 {
				mutations = append(mutations, resultMessage.Mutations...)
			}
			if len(resultMessage.Comments) > 0 {
				comments = append(comments, resultMessage.Comments...)
			}
		}
	}

	return mutations, comments
}

// UpdateMutatedContentWithComments Updates mutated object with comments
func UpdateMutatedContentWithComments(yamlContent string, comments []config.MutationComment) string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(yamlContent))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	commentMap := funk.Map(comments, func(c config.MutationComment) (string, string) {
		return c.Find, c.Comment
	}).(map[string]string)
	fileContent := ""
	for _, line := range lines {
		comment, ok := commentMap[strings.TrimSpace(line)]
		if ok {
			line += (" #" + comment)
		}
		fileContent += line
		fileContent += "\n"
	}
	return fileContent
}
