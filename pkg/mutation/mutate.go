package mutation

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"gomodules.xyz/jsonpatch/v2"
	yaml "gopkg.in/yaml.v3"
)

const (
	strTag  = "!!str"
	seqTag  = "!!seq"
	mapTag  = "!!map"
	intTag  = "!!int"
	boolTag = "!!bool"
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
		splits := getSplitFromPath(patch.Path)
		tag, value, kind, content := getValueTagAndKind(patch.Value)
		switch patch.Operation {
		case "add", "replace":
			var newNode = yaml.Node{
				Kind:    kind,
				Tag:     tag,
				Value:   value,
				Content: content,
			}
			err = addOrReplaceValue(&doc, splits, &newNode)
			if err != nil {
				return mutated, err
			}
		case "remove":
			// ignore error if the value specified does not exists
			_ = removeNodes(&doc, splits)
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

	if result.PodResult != nil {
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

func createPathAndFindNodes(node *yaml.Node, selectors []string, create bool) ([]*yaml.Node, error) {
	var nodes []*yaml.Node
	currentSelector := selectors[0]
	isLastSelector := len(selectors) == 1
	// array[N] or array[*] selectors.
	if i := strings.LastIndex(currentSelector, "["); i > 0 && strings.HasSuffix(currentSelector, "]") {
		arrayIndex := currentSelector[i+1 : len(currentSelector)-1]
		currentSelector = currentSelector[:i]
		if checkIfNodeExistedInContent(node.Content, currentSelector) || !create {
			return findArrayNodes(selectors, currentSelector, node, nodes, arrayIndex, create)
		}
		index, err := strconv.Atoi(arrayIndex)
		if err != nil {
			if arrayIndex != "-" {
				return nil, errors.Wrapf(err, "can't parse array index from %v/%v/", currentSelector, arrayIndex)
			}
			// if index provided is greater than or less than 0 for an empty array should throw an exception
		} else if index != 0 {
			return nil, errors.Errorf("array index (%s) does not exists because array (%s) does not exists", arrayIndex, currentSelector)
		}
		// default to zero since no node is present.
		selectorsToCreateNodes := []string{currentSelector, "0"}
		if len(selectors) > 1 {
			selectorsToCreateNodes = append(selectorsToCreateNodes, selectors[1:]...)
			return createNonExistingPath(selectorsToCreateNodes, node), nil
		}
	}
	if currentSelector == "-" {
		if isLastSelector {
			return []*yaml.Node{node}, nil
		}
		_, err := createPathAndFindNodes(node, selectors[1:], create)
		if err != nil {
			return nil, err
		}
	}
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			// Does the current key match the selector?
			if node.Content[i].Value == currentSelector {
				// Found last key, return its value.
				if !isLastSelector {
					// Match the rest of the selector path, ie. go deeper
					// in to the value node.
					return createPathAndFindNodes(node.Content[i+1], selectors[1:], create)
				}
				return []*yaml.Node{node.Content[i+1]}, nil
			}
		}
	case yaml.ScalarNode:
		// Overwrite any existing nodes.
		node.Kind = yaml.MappingNode
		node.Tag = mapTag
		node.Value = ""
	case yaml.SequenceNode:
		return nil, errors.Errorf("parent node is array, use /*/ or /0/../%v/ instead of .%v to access its item(s) first", len(node.Content)-1, currentSelector)

	default:
		return nil, errors.Errorf("parent node is of unknown kind %v", node.Kind)
	}
	if !create {
		return nil, errors.Errorf("failed to match %s", strings.Join(selectors, "/"))
	}

	return createNonExistingPath(selectors, node), nil
}

func addOrReplaceValue(node *yaml.Node, splits []string, value *yaml.Node) error {
	if len(node.Content) == 0 {
		return errors.New("No content in node")
	}
	nodes, err := createPathAndFindNodes(node.Content[0], splits, true)
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
			// Overwrite a new map node we created in createPathAndFindNodes(), as confirmed
			// by the nil check (the node.Content wouldn't be nil otherwise).
			*node = *value
		} else if node.Kind == yaml.SequenceNode && value.Kind == yaml.SequenceNode {
			// Append new values onto an existing array node.
			node.Content = append(node.Content, value.Content...)
		} else if node.Kind == yaml.SequenceNode && value.Kind == yaml.ScalarNode {
			// Append new value onto an existing array node.
			node.Content = append(node.Content, value)
		} else {
			return errors.Errorf("can't overwrite %v value (line: %v, column: %v) with %v value", node.Tag, node.Line, node.Column, value.Tag)
		}
	}

	return nil
}

func getValueTagAndKind(valueInterface interface{}) (tag, value string, kind yaml.Kind, contents []*yaml.Node) {
	var content []*yaml.Node
	switch v := valueInterface.(type) {
	case int:
		return intTag, strconv.Itoa(v), yaml.ScalarNode, content
	case float64:
		return intTag, strconv.Itoa(int(v)), yaml.ScalarNode, content
	case string:
		if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") && len(v) > 2 {
			itemStr := v[1 : len(v)-1]
			items := strings.Split(itemStr, ",")
			content = getSequenceNodeContent(items)
			return seqTag, "", yaml.SequenceNode, content
		}
		return strTag, v, yaml.ScalarNode, content
	case bool:
		return boolTag, fmt.Sprintf("%t", v), yaml.ScalarNode, content
	default:
		return mapTag, fmt.Sprintf("%v", v), yaml.MappingNode, content
	}
}

func removeNodes(doc *yaml.Node, selectors []string) error {
	err := removeMatchingNode(doc.Content[0], selectors)
	if err != nil {
		return errors.Wrapf(err, "failed to match %q", strings.Join(selectors, "/"))
	}

	return nil
}

func removeMatchingNode(node *yaml.Node, selectors []string) error {
	currentSelector := selectors[0]
	lastSelector := len(selectors) == 1
	if i := strings.LastIndex(currentSelector, "["); i > 0 && strings.HasSuffix(currentSelector, "]") {
		arrayIndex := currentSelector[i+1 : len(currentSelector)-1]
		currentSelector = currentSelector[:i]

		index, err := strconv.Atoi(arrayIndex)
		if err != nil {
			if arrayIndex == "*" {
				index = -1
			} else {
				return errors.Wrapf(err, "can't parse array index from %v/%v/", currentSelector, arrayIndex)
			}
		} else if index < 0 {
			return errors.Wrapf(err, "array index can't be negative %v/%v/", currentSelector, arrayIndex)
		}
		// Go into array node(s).
		arrayNodes, err := createPathAndFindNodes(node, []string{currentSelector}, false)
		if err != nil {
			return errors.Errorf("can't find %v", currentSelector)
		}
		for _, arrayNode := range arrayNodes {
			if arrayNode.Kind != yaml.SequenceNode {
				return errors.Errorf("%v is not an array", currentSelector)
			}
			if index >= len(arrayNode.Content) {
				return errors.Errorf("%v array doesn't have index %v", currentSelector, index)
			}

			var visitArrayNodes []*yaml.Node
			if index >= 0 { // array[N]
				visitArrayNodes = []*yaml.Node{arrayNode.Content[index]}
			} else { // array[*]
				visitArrayNodes = arrayNode.Content
			}
			for _, node := range visitArrayNodes {
				lastSelector := len(selectors) == 1
				if !lastSelector {
					removeMatchingNode(node, selectors[1:])
				}
			}
		}
	}

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

func getSplitFromPath(path string) []string {
	var digitStarCheck = regexp.MustCompile(`^[0-9*]+$`)
	splits := strings.Split(path, "/")
	var formatedSplit []string
	for _, key := range splits {
		if key == "" {
			continue
		}
		if digitStarCheck.MatchString(key) {
			lastElementIdx := len(formatedSplit) - 1
			lastElement := formatedSplit[lastElementIdx]
			lastElement = fmt.Sprintf("%s[%s]", lastElement, key)
			formatedSplit[lastElementIdx] = lastElement
			continue
		}
		formatedSplit = append(formatedSplit, key)
	}
	return formatedSplit
}

func findArrayNodes(selectors []string, currentSelector string, node *yaml.Node, nodes []*yaml.Node, arrayIndex string, create bool) ([]*yaml.Node, error) {

	index, err := strconv.Atoi(arrayIndex)
	if err != nil {
		if arrayIndex == "*" {
			index = -1
		} else {
			return nil, errors.Wrapf(err, "can't parse array index from %v/%v/", currentSelector, arrayIndex)
		}
	} else if index < 0 {
		return nil, errors.Wrapf(err, "array index can't be negative %v/%v/", currentSelector, arrayIndex)
	}
	// Go into array node(s).
	arrayNodes, err := createPathAndFindNodes(node, []string{currentSelector}, create)
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
				deeperNodes, err := createPathAndFindNodes(node, selectors[1:], create)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to go deeper into %v/%v/", currentSelector, i)
				}
				nodes = append(nodes, deeperNodes...)
			}
		}
	}
	return nodes, nil
}

func getSequenceNodeContent(items []string) []*yaml.Node {
	var content []*yaml.Node
	for _, item := range items {
		tag, value, kind, _ := getValueTagAndKind(item)
		var newNode = yaml.Node{
			Kind:  kind,
			Tag:   tag,
			Value: value,
		}
		content = append(content, &newNode)
	}
	return content
}

func checkIfNodeExistedInContent(nodes []*yaml.Node, currentSelector string) bool {
	for i := 0; i < len(nodes); i += 2 {
		// Does the current key match the selector?
		if nodes[i].Value == currentSelector {
			return true
		}
	}

	return false
}

func createNonExistingPath(selectors []string, node *yaml.Node) []*yaml.Node {
	var digitDashCheck = regexp.MustCompile(`^[0-9-]+$`)
	// Create the rest of the selector path.
	for idx, selector := range selectors {
		if digitDashCheck.MatchString(selector) {
			continue
		}
		kind, tag := yaml.MappingNode, mapTag
		// if the next selector is "-" then current selector is sequence/slice/array
		if idx < len(selectors)-1 && digitDashCheck.MatchString(selectors[idx+1]) {
			kind, tag = yaml.SequenceNode, seqTag
		}
		var newNode = yaml.Node{
			Content: []*yaml.Node{
				{
					Kind:  yaml.ScalarNode,
					Tag:   strTag,
					Value: selector,
				},
				{
					Kind: kind,
					Tag:  tag,
				},
			},
		}
		// if previous node is array/sequenceNode append a node rather than appending contents
		if node.Kind == yaml.SequenceNode {
			newNode.Kind = kind
			node.Content = append(node.Content, &newNode)
		} else {
			node.Content = append(node.Content, newNode.Content...)
		}
		node = newNode.Content[len(newNode.Content)-1]
	}
	return []*yaml.Node{node}

}
