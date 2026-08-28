package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/qri-io/jsonschema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	gatewayAPIGroup  = "gateway.networking.k8s.io"
	kgatewayAPIGroup = "gateway.kgateway.dev"
)

type objectReference struct {
	group       string
	kind        string
	name        string
	namespace   string
	sectionName string
	port        int64
}

func init() {
	registerCustomChecks("httpRouteInsecureListener", httpRouteInsecureListener)
	registerCustomChecks("gatewayCrossNamespaceCertificateRef", gatewayCrossNamespaceCertificateRef)
	registerCustomChecks("httpRouteCrossNamespaceBackendRef", httpRouteCrossNamespaceBackendRef)
	registerCustomChecks("httpRouteBackendTLSMissing", httpRouteBackendTLSMissing)
}

func httpRouteInsecureListener(test schemaTestCase) (bool, []jsonschema.KeyError, error) {
	if isFullHTTPSRedirect(test.Resource.Resource.Object) || test.ResourceProvider == nil {
		return true, nil, nil
	}

	routeNamespace := test.Resource.ObjectMeta.GetNamespace()
	for _, parent := range referencesAt(test.Resource.Resource.Object, "spec", "parentRefs") {
		parent = withDefaults(parent, gatewayAPIGroup, "Gateway", routeNamespace)
		if parent.group != gatewayAPIGroup || parent.kind != "Gateway" {
			continue
		}

		gateway := findResource(test.ResourceProvider.Resources[gatewayAPIGroup+"/Gateway"], parent.namespace, parent.name)
		if gateway == nil {
			continue
		}
		listeners := nestedSlice(gateway.Resource.Object, "spec", "listeners")
		for _, rawListener := range listeners {
			listener, ok := rawListener.(map[string]any)
			if !ok || (parent.sectionName != "" && stringValue(listener["name"]) != parent.sectionName) {
				continue
			}
			if stringValue(listener["protocol"]) == "HTTP" && listenerAcceptsHTTPRoute(listener, parent.namespace, test.Resource, test.ResourceProvider) {
				return gatewayFailure("spec.parentRefs", fmt.Sprintf("HTTPRoute references HTTP listener %q on Gateway %s/%s without a full HTTPS redirect", stringValue(listener["name"]), parent.namespace, parent.name))
			}
		}
	}
	return true, nil, nil
}

func gatewayCrossNamespaceCertificateRef(test schemaTestCase) (bool, []jsonschema.KeyError, error) {
	if test.ResourceProvider == nil {
		return true, nil, nil
	}

	sourceNamespace := test.Resource.ObjectMeta.GetNamespace()
	listeners := nestedSlice(test.Resource.Resource.Object, "spec", "listeners")
	for _, rawListener := range listeners {
		listener, ok := rawListener.(map[string]any)
		if !ok {
			continue
		}
		for _, ref := range referencesAt(listener, "tls", "certificateRefs") {
			ref = withDefaults(ref, "", "Secret", sourceNamespace)
			if ref.namespace != sourceNamespace && !hasReferenceGrant(test.ResourceProvider, sourceNamespace, "Gateway", ref) {
				return gatewayFailure("spec.listeners.tls.certificateRefs", fmt.Sprintf("Gateway %s/%s references %s %s/%s without a matching ReferenceGrant", sourceNamespace, test.Resource.ObjectMeta.GetName(), ref.kind, ref.namespace, ref.name))
			}
		}
	}
	return true, nil, nil
}

func httpRouteCrossNamespaceBackendRef(test schemaTestCase) (bool, []jsonschema.KeyError, error) {
	if test.ResourceProvider == nil {
		return true, nil, nil
	}

	sourceNamespace := test.Resource.ObjectMeta.GetNamespace()
	for _, ref := range httpRouteBackendRefs(test.Resource.Resource.Object) {
		ref = withDefaults(ref, "", "Service", sourceNamespace)
		if ref.namespace != sourceNamespace && !hasReferenceGrant(test.ResourceProvider, sourceNamespace, "HTTPRoute", ref) {
			return gatewayFailure("spec.rules.backendRefs", fmt.Sprintf("HTTPRoute %s/%s references %s %s/%s without a matching ReferenceGrant", sourceNamespace, test.Resource.ObjectMeta.GetName(), ref.kind, ref.namespace, ref.name))
		}
	}
	return true, nil, nil
}

func httpRouteBackendTLSMissing(test schemaTestCase) (bool, []jsonschema.KeyError, error) {
	if test.ResourceProvider == nil {
		return true, nil, nil
	}

	routeNamespace := test.Resource.ObjectMeta.GetNamespace()
	for _, ref := range httpRouteBackendRefs(test.Resource.Resource.Object) {
		ref = withDefaults(ref, "", "Service", routeNamespace)
		if !backendUsesTLS(test.ResourceProvider, ref) {
			continue
		}
		if hasBackendTLSPolicy(test.ResourceProvider, ref) || hasKgatewayBackendTLSPolicy(test.ResourceProvider, ref) {
			continue
		}
		return gatewayFailure("spec.rules.backendRefs", fmt.Sprintf("HTTPRoute backend %s %s/%s appears to use TLS but has no BackendTLSPolicy or kgateway BackendConfigPolicy", ref.kind, ref.namespace, ref.name))
	}
	return true, nil, nil
}

func isFullHTTPSRedirect(object map[string]any) bool {
	rules := nestedSlice(object, "spec", "rules")
	if len(rules) == 0 {
		return false
	}
	for _, rawRule := range rules {
		rule, ok := rawRule.(map[string]any)
		if !ok || len(referencesAt(rule, "backendRefs")) > 0 || !ruleMatchesAllTraffic(rule) || !hasHTTPSRedirect(rule) {
			return false
		}
	}
	return true
}

func ruleMatchesAllTraffic(rule map[string]any) bool {
	matches := nestedSlice(rule, "matches")
	if len(matches) == 0 {
		return true
	}
	for _, rawMatch := range matches {
		match, ok := rawMatch.(map[string]any)
		if !ok || len(match) != 1 {
			continue
		}
		path, ok := match["path"].(map[string]any)
		if ok && (stringValue(path["type"]) == "" || stringValue(path["type"]) == "PathPrefix") && stringValue(path["value"]) == "/" {
			return true
		}
	}
	return false
}

func hasHTTPSRedirect(rule map[string]any) bool {
	filters := nestedSlice(rule, "filters")
	for _, rawFilter := range filters {
		filter, ok := rawFilter.(map[string]any)
		if !ok || stringValue(filter["type"]) != "RequestRedirect" {
			continue
		}
		redirect, ok := filter["requestRedirect"].(map[string]any)
		if ok && strings.EqualFold(stringValue(redirect["scheme"]), "https") {
			return true
		}
	}
	return false
}

func listenerAcceptsHTTPRoute(listener map[string]any, gatewayNamespace string, route kube.GenericResource, provider *kube.ResourceProvider) bool {
	if !listenerHostnameIntersectsRoute(listener, route.Resource.Object) {
		return false
	}

	allowedRoutes, ok := listener["allowedRoutes"].(map[string]any)
	if !ok {
		return route.ObjectMeta.GetNamespace() == gatewayNamespace
	}
	if kinds := nestedSlice(allowedRoutes, "kinds"); len(kinds) > 0 {
		allowsHTTPRoute := false
		for _, rawKind := range kinds {
			kind, ok := rawKind.(map[string]any)
			if ok && withDefaultString(stringValue(kind["group"]), gatewayAPIGroup) == gatewayAPIGroup && stringValue(kind["kind"]) == "HTTPRoute" {
				allowsHTTPRoute = true
				break
			}
		}
		if !allowsHTTPRoute {
			return false
		}
	}

	namespaces, ok := allowedRoutes["namespaces"].(map[string]any)
	if !ok || stringValue(namespaces["from"]) == "" || stringValue(namespaces["from"]) == "Same" {
		return route.ObjectMeta.GetNamespace() == gatewayNamespace
	}
	if stringValue(namespaces["from"]) == "All" {
		return true
	}
	if stringValue(namespaces["from"]) != "Selector" {
		return false
	}
	selectorMap, ok := namespaces["selector"].(map[string]any)
	if !ok {
		return false
	}
	selector := &metav1.LabelSelector{}
	selectorJSON, err := json.Marshal(selectorMap)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(selectorJSON, selector); err != nil {
		return false
	}
	compiled, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return false
	}
	for _, namespace := range provider.Namespaces {
		if namespace.Name == route.ObjectMeta.GetNamespace() {
			return compiled.Matches(labels.Set(namespace.Labels))
		}
	}
	return false
}

func listenerHostnameIntersectsRoute(listener, route map[string]any) bool {
	listenerHostname := stringValue(listener["hostname"])
	routeHostnames := nestedSlice(route, "spec", "hostnames")
	if len(routeHostnames) == 0 || listenerHostname == "" {
		return true
	}
	for _, routeHostname := range routeHostnames {
		if hostnamesIntersect(listenerHostname, stringValue(routeHostname)) {
			return true
		}
	}
	return false
}

func hostnamesIntersect(left, right string) bool {
	if left == "" || right == "" || left == "*" || right == "*" || strings.EqualFold(left, right) {
		return true
	}
	leftSuffix, leftWildcard := strings.CutPrefix(strings.ToLower(left), "*.")
	rightSuffix, rightWildcard := strings.CutPrefix(strings.ToLower(right), "*.")
	switch {
	case leftWildcard && rightWildcard:
		return leftSuffix == rightSuffix || strings.HasSuffix(leftSuffix, "."+rightSuffix) || strings.HasSuffix(rightSuffix, "."+leftSuffix)
	case leftWildcard:
		return strings.HasSuffix(strings.ToLower(right), "."+leftSuffix)
	case rightWildcard:
		return strings.HasSuffix(strings.ToLower(left), "."+rightSuffix)
	default:
		return false
	}
}

func httpRouteBackendRefs(object map[string]any) []objectReference {
	var refs []objectReference
	rules := nestedSlice(object, "spec", "rules")
	for _, rawRule := range rules {
		rule, ok := rawRule.(map[string]any)
		if !ok {
			continue
		}
		refs = append(refs, referencesAt(rule, "backendRefs")...)
		filters := nestedSlice(rule, "filters")
		for _, rawFilter := range filters {
			filter, ok := rawFilter.(map[string]any)
			if !ok || stringValue(filter["type"]) != "RequestMirror" {
				continue
			}
			mirror, ok := filter["requestMirror"].(map[string]any)
			if !ok {
				continue
			}
			if backend, ok := mirror["backendRef"].(map[string]any); ok {
				refs = append(refs, referenceFromMap(backend))
			}
		}
	}
	return refs
}

func referencesAt(object map[string]any, fields ...string) []objectReference {
	items := nestedSlice(object, fields...)
	if len(items) == 0 {
		return nil
	}
	refs := make([]objectReference, 0, len(items))
	for _, item := range items {
		if ref, ok := item.(map[string]any); ok {
			refs = append(refs, referenceFromMap(ref))
		}
	}
	return refs
}

func referenceFromMap(ref map[string]any) objectReference {
	return objectReference{
		group:       stringValue(ref["group"]),
		kind:        stringValue(ref["kind"]),
		name:        stringValue(ref["name"]),
		namespace:   stringValue(ref["namespace"]),
		sectionName: stringValue(ref["sectionName"]),
		port:        int64Value(ref["port"]),
	}
}

func withDefaults(ref objectReference, group, kind, namespace string) objectReference {
	if ref.group == "" {
		ref.group = group
	}
	if ref.kind == "" {
		ref.kind = kind
	}
	if ref.namespace == "" {
		ref.namespace = namespace
	}
	return ref
}

func withDefaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func hasReferenceGrant(provider *kube.ResourceProvider, sourceNamespace, sourceKind string, target objectReference) bool {
	for _, grant := range provider.Resources[gatewayAPIGroup+"/ReferenceGrant"] {
		if grant.ObjectMeta.GetNamespace() != target.namespace {
			continue
		}
		fromMatches := false
		for _, from := range referencesAt(grant.Resource.Object, "spec", "from") {
			if from.group == gatewayAPIGroup && from.kind == sourceKind && from.namespace == sourceNamespace {
				fromMatches = true
				break
			}
		}
		if !fromMatches {
			continue
		}
		for _, to := range referencesAt(grant.Resource.Object, "spec", "to") {
			if to.group == target.group && to.kind == target.kind && (to.name == "" || to.name == target.name) {
				return true
			}
		}
	}
	return false
}

func backendUsesTLS(provider *kube.ResourceProvider, ref objectReference) bool {
	// ponytail: infer TLS from conventional ports and backend metadata; replace
	// this with controller status or an implementation graph when Polaris has one.
	if ref.port == 443 || ref.port == 8443 {
		return true
	}
	groupKind := ref.kind
	if ref.group != "" {
		groupKind = ref.group + "/" + ref.kind
	}
	backend := findResource(provider.Resources[groupKind], ref.namespace, ref.name)
	if backend == nil {
		return false
	}
	if ref.group == "" && ref.kind == "Service" {
		ports := nestedSlice(backend.Resource.Object, "spec", "ports")
		for _, rawPort := range ports {
			port, ok := rawPort.(map[string]any)
			if !ok || (ref.port != 0 && int64Value(port["port"]) != ref.port) {
				continue
			}
			name := strings.ToLower(stringValue(port["name"]))
			appProtocol := strings.ToLower(stringValue(port["appProtocol"]))
			if name == "https" || strings.HasPrefix(name, "https-") || appProtocol == "https" || strings.HasSuffix(appProtocol, "/https") {
				return true
			}
		}
	}
	if ref.group == kgatewayAPIGroup && ref.kind == "Backend" {
		hosts := nestedSlice(backend.Resource.Object, "spec", "static", "hosts")
		for _, rawHost := range hosts {
			host, ok := rawHost.(map[string]any)
			if ok && (int64Value(host["port"]) == 443 || int64Value(host["port"]) == 8443) {
				return true
			}
		}
	}
	return false
}

func hasBackendTLSPolicy(provider *kube.ResourceProvider, ref objectReference) bool {
	if ref.group != "" || ref.kind != "Service" {
		return false
	}
	for _, policy := range provider.Resources[gatewayAPIGroup+"/BackendTLSPolicy"] {
		if policy.ObjectMeta.GetNamespace() == ref.namespace && policyTargets(policy, ref) {
			return true
		}
	}
	return false
}

func hasKgatewayBackendTLSPolicy(provider *kube.ResourceProvider, ref objectReference) bool {
	for _, policy := range provider.Resources[kgatewayAPIGroup+"/BackendConfigPolicy"] {
		if policy.ObjectMeta.GetNamespace() != ref.namespace {
			continue
		}
		if _, found := nestedValue(policy.Resource.Object, "spec", "tls"); found && policyTargets(policy, ref) {
			return true
		}
	}
	return false
}

func policyTargets(policy kube.GenericResource, target objectReference) bool {
	for _, ref := range referencesAt(policy.Resource.Object, "spec", "targetRefs") {
		ref = withDefaults(ref, "", "Service", policy.ObjectMeta.GetNamespace())
		if ref.group == target.group && ref.kind == target.kind && ref.name == target.name {
			return true
		}
	}
	return false
}

func findResource(resources []kube.GenericResource, namespace, name string) *kube.GenericResource {
	for i := range resources {
		if resources[i].ObjectMeta.GetNamespace() == namespace && resources[i].ObjectMeta.GetName() == name {
			return &resources[i]
		}
	}
	return nil
}

func stringValue(value any) string {
	valueString, _ := value.(string)
	return valueString
}

func int64Value(value any) int64 {
	switch number := value.(type) {
	case int:
		return int64(number)
	case int32:
		return int64(number)
	case int64:
		return number
	case float64:
		return int64(number)
	default:
		return 0
	}
}

func nestedSlice(object map[string]any, fields ...string) []any {
	value, found := nestedValue(object, fields...)
	if !found {
		return nil
	}
	items, _ := value.([]any)
	return items
}

func nestedValue(object map[string]any, fields ...string) (any, bool) {
	var current any = object
	for _, field := range fields {
		currentMap, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = currentMap[field]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func gatewayFailure(path, message string) (bool, []jsonschema.KeyError, error) {
	return false, []jsonschema.KeyError{{
		PropertyPath: path,
		Message:      message,
	}}, nil
}
