package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Configuration contains all of the config for the validation checks.
type Configuration struct {
	Resources    Resources    `json:"resources"`
	HealthChecks HealthChecks `json:"healthChecks"`
	Images       Images       `json:"images"`
	Networking   Networking   `json:"networking"`
	Security     Security     `json:"security"`
}

// Severity represents the severity of action to take (Ignore, Warning, Error).
type Severity string

// Resources contains config for resource requests and limits.
type Resources struct {
	CPURequests    Resource `json:"cpuRequests"`
	CPULimits      Resource `json:"cpuLimits"`
	MemoryRequests Resource `json:"memoryRequests"`
	MemoryLimits   Resource `json:"memoryLimits"`
}

// Resource contains config for requests or limits for a specific resource.
type Resource struct {
	Absent  Severity      `json:"absent"`
	Warning ResourceRange `json:"warning"`
	Error   ResourceRange `json:"error"`
}

// ResourceRange can contain below and above conditions for validation.
type ResourceRange struct {
	Below *resource.Quantity `json:"below"`
	Above *resource.Quantity `json:"above"`
}

// HealthChecks contains config for readiness and liveness probes.
type HealthChecks struct {
	ReadinessProbeMissing Severity `json:"readinessProbeMissing"`
	LivenessProbeMissing  Severity `json:"livenessProbeMissing"`
}

// Images contains the config for images.
type Images struct {
	TagNotSpecified     Severity     `json:"tagNotSpecified"`
	PullPolicyNotAlways Severity     `json:"pullPolicyNotAlways"`
	Repositories        Repositories `json:"repositories"`
}

// Repositories provides lists of patterns to match or avoid in image tags.
type Repositories struct {
	Error   WhitelistBlacklist `json:"error"`
	Warning WhitelistBlacklist `json:"warning"`
}

// WhitelistBlacklist can contain a whitelist or blacklist.
type WhitelistBlacklist struct {
	Whitelist []string `json:"whitelist"`
	Blacklist []string `json:"blacklist"`
}

// Networking contains the config for networking validations.
type Networking struct {
	HostAliasSet   Severity `json:"hostAliasSet"`
	HostIPCSet     Severity `json:"hostIPCSet"`
	HostNetworkSet Severity `json:"hostNetworkSet"`
	HostPIDSet     Severity `json:"hostPIDSet"`
	HostPortSet    Severity `json:"hostPortSet"`
}

// Security contains the config for security validations.
type Security struct {
	runAsNonRoot              Severity             `json:"runAsNonRoot"`
	runAsPriviliged           Severity             `json:"runAsPriviliged"`
	notReadOnlyRootFileSystem Severity             `json:"notReadOnlyRootFileSystem"`
	capabilities              SecurityCapabilities `json:"capabilities"`
}

// SecurityCapabilities contains the config for security capabilities validations.
type SecurityCapabilities struct {
	Error   WhitelistBlacklist `json:"error"`
	Warning WhitelistBlacklist `json:"warning"`
}

// ParseFile parses config from a file.
func ParseFile(path string) (Configuration, error) {
	rawBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return Configuration{}, err
	}
	return Parse(rawBytes)
}

// Parse parses config from a byte array.
func Parse(rawBytes []byte) (Configuration, error) {
	reader := bytes.NewReader(rawBytes)
	conf := Configuration{}
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(&conf); err != nil {
			if err == io.EOF {
				return conf, nil
			}
			return conf, fmt.Errorf("Decoding config failed: %v", err)
		}
	}
}
