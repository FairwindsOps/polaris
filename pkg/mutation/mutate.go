package mutation

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	jsonpatchV5 "github.com/evanphx/json-patch/v5"
	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"gomodules.xyz/jsonpatch/v2"
	yaml "gopkg.in/yaml.v3"
)

// ApplyAllMutations applies available mutation to a single resource
func ApplyAllMutations(manifest string, mutations []jsonpatch.Operation) (string, error) {
	var mutated string
	var doc yaml.Node
	err := yaml.Unmarshal([]byte(manifest), &doc)
	if err != nil {
		return mutated, err
	}

	for _, patch := range mutations {
		tag, value, kind := getValueTagAndKind(patch.Value)
		switch patch.Operation {
		case "add", "replace":
			fmt.Println(patch.Path)
			var newNode = yaml.Node{
				Kind:  kind,
				Tag:   tag,
				Value: value,
			}
			err = addOrReplaceValue(&doc, patch.Path, &newNode)
			if err != nil {
				return mutated, err
			}
		case "remove":
			// ignore error if the value specified does not exists
			_ = removeNodes(&doc, patch.Path)
		}
	}

	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err = enc.Encode(&doc)
	if err != nil {
		return mutated, err
	}
	err = enc.Close()
	if err != nil {
		return mutated, err
	}

	return buf.String(), nil
}

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
		EnsurePathExistsOnAdd:    true,
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

func findNodes(node *yaml.Node, selectors []string) ([]*yaml.Node, error) {
	var nodes []*yaml.Node
	currentSelector := selectors[0]
	// array[N] or array[*] selectors.
	if i := strings.LastIndex(currentSelector, "["); i > 0 && strings.HasSuffix(currentSelector, "]") {
		arrayIndex := currentSelector[i+1 : len(currentSelector)-1]
		currentSelector = currentSelector[:i]

		index, err := strconv.Atoi(arrayIndex)
		if err != nil {
			if arrayIndex == "*" {
				index = -1
			} else {
				return nil, errors.Wrapf(err, "can't parse array index from %v[%v]", currentSelector, arrayIndex)
			}
		} else if index < 0 {
			return nil, errors.Wrapf(err, "array index can't be negative %v[%v]", currentSelector, arrayIndex)
		}

		// Go into array node(s).
		arrayNodes, err := findNodes(node, []string{currentSelector})
		if err != nil {
			return nil, errors.Errorf("can't find %v", currentSelector)
		}
		for _, arrayNode := range arrayNodes {
			if arrayNode.Kind != yaml.SequenceNode {
				return nil, errors.Errorf("%v is not an array", currentSelector)
			}
			if index >= len(arrayNode.Content) {
				return nil, errors.Errorf("%v array doesn't have index %v", currentSelector, index)
			}

			var visitArrayNodes []*yaml.Node
			if index >= 0 { // array[N]
				visitArrayNodes = []*yaml.Node{arrayNode.Content[index]}
			} else { // array[*]
				visitArrayNodes = arrayNode.Content
			}

			for i, node := range visitArrayNodes {
				if len(selectors) == 1 {
					// Last selector, use this as final node.
					nodes = append(nodes, node)
				} else {
					// Go deeper into a specific array.
					deeperNodes, err := findNodes(node, selectors[1:])
					if err != nil {
						return nil, errors.Wrapf(err, "failed to go deeper into %v[%v]", currentSelector, i)
					}
					nodes = append(nodes, deeperNodes...)
				}
			}
		}
		return nodes, nil
	}
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			// Does the current key match the selector?
			if node.Content[i].Value == currentSelector {
				// Found last key, return its value.
				isLastSelector := len(selectors) == 1
				if !isLastSelector {
					// Match the rest of the selector path, ie. go deeper
					// in to the value node.
					return findNodes(node.Content[i+1], selectors[1:])
				}
				return []*yaml.Node{node.Content[i+1]}, nil
			}
		}
	case yaml.ScalarNode:
		// Overwrite any existing nodes.
		node.Kind = yaml.MappingNode
		node.Tag = "!!map"
		node.Value = ""
	case yaml.SequenceNode:
		return nil, errors.Errorf("parent node is array, use [*] or [0]..[%v] instead of .%v to access its item(s) first", len(node.Content)-1, currentSelector)

	default:
		return nil, errors.Errorf("parent node is of unknown kind %v", node.Kind)
	}
	// Create the rest of the selector path.
	for _, selector := range selectors {
		var newNode = yaml.Node{
			Content: []*yaml.Node{
				{
					Kind:  yaml.ScalarNode,
					Tag:   "!!str",
					Value: selector,
				},
				{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
				},
			},
		}
		node.Content = append(node.Content, newNode.Content...)
		node = newNode.Content[len(newNode.Content)-1]
	}

	return []*yaml.Node{node}, nil
}

func addOrReplaceValue(node *yaml.Node, path string, value *yaml.Node) error {
	splits := strings.Split(path, "/")
	nodes, err := findNodes(node.Content[0], splits)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if node.Kind == yaml.ScalarNode {
			// Overwrite an existing scalar value with a new value (whatever kind).
			*node = *value
		} else if node.Kind == yaml.MappingNode && value.Kind == yaml.MappingNode {
			// Append new values onto an existing map node.
			node.Content = append(value.Content, node.Content...)
		} else if node.Kind == yaml.MappingNode && node.Content == nil {
			// Overwrite a new map node we created in findNodes(), as confirmed
			// by the nil check (the node.Content wouldn't be nil otherwise).
			*node = *value
		} else if node.Kind == yaml.SequenceNode && value.Kind == yaml.SequenceNode {
			// Append new values onto an existing array node.
			node.Content = append(value.Content, node.Content...)
		} else {
			return errors.Errorf("can't overwrite %v value (line: %v, column: %v) with %v value", node.Tag, node.Line, node.Column, value.Tag)
		}
	}

	return nil
}

func getValueTagAndKind(valueInterface interface{}) (tag, value string, kind yaml.Kind) {
	switch v := valueInterface.(type) {
	case int:
		return "!!int", strconv.Itoa(v), yaml.ScalarNode
	case float64:
		return "!!float", fmt.Sprintf("%f", v), yaml.ScalarNode
	case string:
		// v is a string here, so e.g. v + " Yeah!" is possible.
		return "!!str", v, yaml.ScalarNode
	default:
		return "!!map", fmt.Sprintf("%v", v), yaml.MappingNode
	}
}

func removeNodes(doc *yaml.Node, path string) error {
	selectors := strings.Split(path, "/")

	err := removeMatchingNode(doc.Content[0], selectors)
	if err != nil {
		return errors.Wrapf(err, "failed to match %q", path)
	}

	return nil
}

func removeMatchingNode(node *yaml.Node, selectors []string) error {
	currentSelector := selectors[0]
	lastSelector := len(selectors) == 1

	// Iterate over the keys (the slice is key/value pairs).
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == currentSelector {
			// Key matches the selector.
			if !lastSelector {
				// Try to match the rest of the selector path in the value.
				return removeMatchingNode(node.Content[i+1], selectors[1:])
			}

			node.Content[i] = nil   // Delete key.
			node.Content[i+1] = nil // Delete value.
			node.Content = append(node.Content[:i], node.Content[i+2:]...)
			return nil
		}
	}

	return errors.Errorf("can't find %q", strings.Join(selectors, "."))
}
