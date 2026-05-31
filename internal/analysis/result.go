package analysis

import (
	"github.com/samaasi/kubectl-plan/internal/dependency"
	"github.com/samaasi/kubectl-plan/internal/risk"
)

type AnalysisResult struct {
	Action              string
	Target              dependency.Node
	Risk                risk.RiskScore
	Confidence          OverallConfidence
	Uncertainty         risk.UncertaintyScore
	Graph               dependency.DependencyGraph
	CrossNamespace      bool
	ClusterUID          string
	ClusterVersion      ClusterVersion
	DataSources         DataSources
	AutoRecordedOutcome interface{}
}

type OverallConfidence struct {
	Overall float64
	Sources []string
}

type ClusterVersion struct {
	Major string
	Minor string
}

type DataSources struct {
	PrometheusAvailable bool
	ServiceMeshDetected bool
	K8sAPIAvailable     bool
}
