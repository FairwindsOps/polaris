package messages

const (
	// CategoryHealthChecks category
	CategoryHealthChecks = "Health Checks"
	// CategorySecurity category
	CategorySecurity = "Security"
	// CategoryNetworking category
	CategoryNetworking = "Networking"
	// CategoryResources category
	CategoryResources = "Resources"
	// CategoryImages category
	CategoryImages = "Images"

	// CPURequestsLabel label
	CPURequestsLabel = "CPU requests"
	// CPULimitsLabel label
	CPULimitsLabel = "CPU limits"
	// MemoryRequestsLabel label
	MemoryRequestsLabel = "Memory requests"
	// MemoryLimitsLabel label
	MemoryLimitsLabel = "Memory limits"

	// CPURequestsFailure message
	CPURequestsFailure = "CPU requests should be set"
	// CPULimitsFailure message
	CPULimitsFailure = "CPU limits should be set"
	// MemoryRequestsFailure message
	MemoryRequestsFailure = "Memory requests should be set"
	// MemoryLimitsFailure message
	MemoryLimitsFailure = "Memory limits should be set"
	// ResourceAmountTooHighFailure message
	ResourceAmountTooHighFailure = "%s should be lower than %s"
	// ResourceAmountTooLowFailure message
	ResourceAmountTooLowFailure = "%s should be higher than %s"
	// ResourceAmountSuccess message
	ResourceAmountSuccess = "%s are within the expected range"
	// ReadinessProbeFailure message
	ReadinessProbeFailure = "Readiness probe should be configured"
	// ReadinessProbeSuccess message
	ReadinessProbeSuccess = "Readiness probe configured"
	// LivenessProbeFailure message
	LivenessProbeFailure = "Liveness probe should be configured"
	// LivenessProbeSuccess message
	LivenessProbeSuccess = "Liveness probe is configured"
	// ImageTagFailure message
	ImageTagFailure = "Image tag should be specified"
	// ImageTagSuccess message
	ImageTagSuccess = "Image tag is specified"
	// HostPortFailure message
	HostPortFailure = "Host port should not be configured"
	// HostPortSuccess message
	HostPortSuccess = "Host port is not configured"
	// RunAsRootFailure message
	RunAsRootFailure = "Should not be running as root"
	// RunAsRootSuccess message
	RunAsRootSuccess = "Not running as root"
	// RunAsPrivilegedFailure message
	RunAsPrivilegedFailure = "Should not be running as privileged"
	// RunAsPrivilegedSuccess message
	RunAsPrivilegedSuccess = "Not running as privileged"
	// ReadOnlyFilesystemSuccess message
	ReadOnlyFilesystemSuccess = "Filesystem is read only"
	// ReadOnlyFilesystemFailure message
	ReadOnlyFilesystemFailure = "Filesystem should be read only"
	// PrivilegeEscalationFailure message
	PrivilegeEscalationFailure = "Privilege escalation should not be allowed"
	// PrivilegeEscalationSuccess message
	PrivilegeEscalationSuccess = "Privilege escalation not allowed"
	// SecurityCapabilitiesAddedFailure message
	SecurityCapabilitiesAddedFailure = "The following security capabilities should not be added: %v"
	// SecurityCapabilitiesNotDroppedFailure message
	SecurityCapabilitiesNotDroppedFailure = "The following security capabilities should be dropped: %v"
	// SecurityCapabilitiesSuccess message
	SecurityCapabilitiesSuccess = "Security capabilities are within the configured limits"

	// HostAliasFailure message
	HostAliasFailure = "Host alias should not be configured"
	// HostAliasSuccess message
	HostAliasSuccess = "Host alias is not configured"
	// HostIPCFailure message
	HostIPCFailure = "Host IPC should not be configured"
	// HostIPCSuccess message
	HostIPCSuccess = "Host IPC is not configured"
	// HostPIDFailure message
	HostPIDFailure = "Host PID should not be configured"
	// HostPIDSuccess message
	HostPIDSuccess = "Host PID is not configured"
	// HostNetworkFailure message
	HostNetworkFailure = "Host network should not be configured"
	// HostNetworkSuccess message
	HostNetworkSuccess = "Host network is not configured"
)
