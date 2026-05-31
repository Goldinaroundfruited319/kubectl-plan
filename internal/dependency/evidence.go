package dependency

type Evidence struct {
	Type        EvidenceType
	Source      EvidenceSource
	Description string
	Confidence  float64
	RawValue    string
}

type EvidenceType string

const (
	EvidenceLabelSelector  EvidenceType = "label_selector"
	EvidenceIngressBackend EvidenceType = "ingress_backend"
	EvidenceEnvVar         EvidenceType = "env_var"
	EvidencePrometheus     EvidenceType = "prometheus_traffic"
	EvidenceOwnerRef       EvidenceType = "owner_reference"
	EvidenceNetworkPolicy  EvidenceType = "network_policy"
	EvidenceVolumeMount    EvidenceType = "volume_mount"
	EvidenceDNSPattern     EvidenceType = "dns_pattern"
)

type EvidenceSource string

const (
	SourceK8sAPI     EvidenceSource = "kubernetes_api"
	SourcePrometheus EvidenceSource = "prometheus"
)
