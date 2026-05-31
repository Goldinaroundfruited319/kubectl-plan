package risk

import (
	"fmt"
	"math"
	"strings"

	"github.com/samaasi/kubectl-plan/internal/criticality"
	"github.com/samaasi/kubectl-plan/internal/dependency"
)

type Scorer struct {
	profile *criticality.ProfileConfig
}

func NewScorer(profile *criticality.ProfileConfig) *Scorer {
	return &Scorer{profile: profile}
}

func (s *Scorer) Score(graph *dependency.DependencyGraph) (RiskScore, UncertaintyScore) {
	var contributors []Contributor
	var activeWeightSum int
	var weightedValueSum float64

	ns := graph.Target.Namespace
	critLevel, critMult := criticality.EvaluateNamespace(ns, s.profile)

	directCount := 0
	indirectCount := 0
	namespaces := make(map[string]bool)
	hasIngress := false
	hasPDB := false
	hasHPA := false

	namespaces[graph.Target.Namespace] = true

	for _, node := range graph.Nodes {
		if node.Kind == graph.Target.Kind && node.Name == graph.Target.Name && node.Namespace == graph.Target.Namespace {
			continue
		}
		if node.Kind == "Pod" {
			continue
		}

		namespaces[node.Namespace] = true

		if node.Kind == "Ingress" {
			hasIngress = true
		} else if node.Kind == "PodDisruptionBudget" {
			hasPDB = true
		} else if node.Kind == "HorizontalPodAutoscaler" {
			hasHPA = true
		}
	}

	for _, edge := range graph.Edges {
		if edge.Relationship == dependency.RelOwns {
			continue
		}
		if edge.Depth == 1 {
			directCount++
		} else if edge.Depth > 1 {
			indirectCount++
		}
	}

	for _, rule := range RuleRegistry {
		if rule.Phase > 1 {
			continue
		}

		var value float64
		var details string

		switch rule.ID {
		case "namespace_criticality":
			value = critMult
			details = fmt.Sprintf("criticality: %s", strings.ToUpper(critLevel))
		case "direct_dependents":
			value = math.Min(float64(directCount)/5.0, 1.0)
			details = fmt.Sprintf("%d confirmed direct consumers", directCount)
		case "indirect_dependents":
			value = math.Min(float64(indirectCount)/10.0, 1.0)
			details = fmt.Sprintf("%d confirmed indirect consumers", indirectCount)
		case "ingress_exposed":
			if hasIngress {
				value = 1.0
				details = "Ingress routes traffic externally"
			} else {
				value = 0.0
				details = "No Ingress route detected"
			}
		case "cross_namespace_impact":
			if len(namespaces) > 1 {
				value = 1.0
				details = fmt.Sprintf("Blast radius spans %d namespaces", len(namespaces))
			} else {
				value = 0.0
				details = "Contained within single namespace"
			}
		case "has_pdb":
			if hasPDB {
				value = 1.0
				details = "PDB constrains rollouts/scaling"
			} else {
				value = 0.0
				details = "No PDB present"
			}
		case "has_hpa":
			if hasHPA {
				value = 0.5
				details = "HPA present, may auto-recover"
			} else {
				value = 0.0
				details = "No HPA active"
			}
		}

		contrib := value * float64(rule.Weight)
		contributors = append(contributors, Contributor{
			RuleID:       rule.ID,
			Name:         rule.Name,
			Weight:       rule.Weight,
			Value:        value,
			Contribution: contrib / 100.0 * 10.0,
			Details:      details,
		})

		activeWeightSum += rule.Weight
		weightedValueSum += contrib
	}

	finalScore := 0.0
	if activeWeightSum > 0 {
		finalScore = (weightedValueSum / float64(activeWeightSum)) * 10.0
	}

	finalScore = math.Round(finalScore*10) / 10

	var level RiskLevel
	switch {
	case finalScore <= 3.0:
		level = RiskLow
	case finalScore <= 6.0:
		level = RiskMedium
	case finalScore <= 8.5:
		level = RiskHigh
	default:
		level = RiskCritical
	}

	riskScore := RiskScore{
		Score:        finalScore,
		Level:        level,
		Contributors: contributors,
	}

	uncLevel := UncertaintyHigh
	uncScore := 0.75
	reasons := []string{"No Prometheus available; performing topology-only analysis"}

	hasEnvVarOnly := true
	for _, edge := range graph.Edges {
		if edge.Relationship != dependency.RelEnvRef && edge.Relationship != dependency.RelOwns {
			hasEnvVarOnly = false
			break
		}
	}
	if hasEnvVarOnly && len(graph.Edges) > 0 {
		uncLevel = UncertaintyVeryHigh
		uncScore = 0.95
		reasons = append(reasons, "Inferences are based primarily on environment variable string matches")
	}

	uncertainty := UncertaintyScore{
		Level:   uncLevel,
		Reasons: reasons,
		Score:   uncScore,
	}

	return riskScore, uncertainty
}
