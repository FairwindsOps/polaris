package messages

const (
	// CPURequestsLabel label
	CPURequestsLabel = "CPU requests"
	// CPULimitsLabel label
	CPULimitsLabel = "CPU limits"
	// MemoryRequestsLabel label
	MemoryRequestsLabel = "Memory requests"
	// MemoryLimitsLabel label
	MemoryLimitsLabel = "Memory limits"

	// CPURequestsMissing message
	CPURequestsMissing = "CPU requests should be set"
	// CPULimitsMissing message
	CPULimitsMissing = "CPU limits should be set"
	// MemoryRequestsMissing message
	MemoryRequestsMissing = "Memory requests should be set"
	// MemoryLimitsMissing message
	MemoryLimitsMissing = "Memory limits should be set"
	// ResourceAmountTooHigh message
	ResourceAmountTooHigh = "%s should be lower than %s"
	// ResourceAmountTooLow message
	ResourceAmountTooLow = "%s should be higher than %s"
	// ResourceAmountCorrect message
	ResourceAmountCorrect = "%s are within the expected range"
	// ReadinessProbeNotConfigured message
	ReadinessProbeNotConfigured = "Readiness probe should be configured"
	// ReadinessProbeConfigured message
	ReadinessProbeConfigured = "Readiness probe configured"
	// LivenessProbeNotConfigured message
	LivenessProbeNotConfigured = "Liveness probe should be configured"
	// LivenessProbeConfigured message
	LivenessProbeConfigured = "Liveness probe is configured"
	// ImageTagNotSpecified message
	ImageTagNotSpecified = "Image tag should be specified"
	// ImageTagSpecified message
	ImageTagSpecified = "Image tag is specified"
	// HostPortSet message
	HostPortSet = "Host port should not be configured"
	// HostPortNotSet message
	HostPortNotSet = "Host port is not configured"
	// RunAsRootEnabled message
	RunAsRootEnabled = "Should not be running as root"
	// RunAsRootDisabled message
	RunAsRootDisabled = "Not running as root"
	// RunAsPrivilegedEnabled message
	RunAsPrivilegedEnabled = "Should not be running as privileged"
	// RunAsPrivilegedDisabled message
	RunAsPrivilegedDisabled = "Not running as privileged"
	// ReadOnlyFilesystem message
	ReadOnlyFilesystem = "Filesystem is read only"
	// WritableFilesystem message
	WritableFilesystem = "Filesystem should be read only"
	// PrivilegeEscalationAllowed message
	PrivilegeEscalationAllowed = "Privilege escalation should not be allowed"
	// PrivilegeEscalationDisallowed message
	PrivilegeEscalationDisallowed = "Privilege escalation not allowed"
	// HasBadSecurityCapabilities message
	HasBadSecurityCapabilities = "The following security capabilities should not be added: %v"
	// SecurityCapabilitiesOK message
	SecurityCapabilitiesOK = "Security capabilities are within the configured limits"
	// SecurityCapabilitiesNotDropped message
	SecurityCapabilitiesNotDropped = "The following security capabilities should be dropped: %v"

	// HostAliasConfigured message
	HostAliasConfigured = "Host alias should not be configured"
	// HostAliasNotConfigured message
	HostAliasNotConfigured = "Host alias is not configured"
	// HostIPCConfigured message
	HostIPCConfigured = "Host IPC should not be configured"
	// HostIPCNotConfigured message
	HostIPCNotConfigured = "Host IPC is not configured"
	// HostPIDConfigured message
	HostPIDConfigured = "Host PID should not be configured"
	// HostPIDNotConfigured message
	HostPIDNotConfigured = "Host PID is not configured"
	// HostNetworkConfigured message
	HostNetworkConfigured = "Host network should not be configured"
	// HostNetworkNotConfigured message
	HostNetworkNotConfigured = "Host network is not configured"
)
