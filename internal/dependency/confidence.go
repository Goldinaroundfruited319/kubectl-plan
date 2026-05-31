package dependency

func GetEvidenceConfidence(t EvidenceType) float64 {
	switch t {
	case EvidenceOwnerRef:
		return 1.00
	case EvidenceLabelSelector:
		return 0.95
	case EvidenceIngressBackend:
		return 0.95
	case EvidenceNetworkPolicy:
		return 0.80
	case EvidenceEnvVar:
		return 0.70
	case EvidenceDNSPattern:
		return 0.65
	case EvidenceVolumeMount:
		return 0.60
	case EvidencePrometheus:
		return 0.99
	default:
		return 0.50
	}
}
