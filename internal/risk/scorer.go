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
	ns := graph.Target.Namespace
	critLevel, critMult := criticality.EvaluateNamespace(ns, s.profile)

	directCount, indirectCount, hasIngress, hasPDB, hasHPA, namespaces := graphStats(graph)

	var contributors []Contributor
	var activeWeightSum int
	var weightedValueSum float64

	for _, rule := range RuleRegistry {
		if rule.Phase > 1 {
			continue
		}
		value, details := evaluateRule(rule.ID, critLevel, critMult, directCount, indirectCount,
			hasIngress, hasPDB, hasHPA, len(namespaces))

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

	return RiskScore{
		Score:        finalScore,
		Level:        scoreToLevel(finalScore),
		Contributors: contributors,
	}, computeUncertainty(graph)
}

func graphStats(graph *dependency.DependencyGraph) (direct, indirect int, hasIngress, hasPDB, hasHPA bool, namespaces map[string]bool) {
	namespaces = map[string]bool{graph.Target.Namespace: true}

	for _, node := range graph.Nodes {
		if node.Kind == graph.Target.Kind && node.Name == graph.Target.Name && node.Namespace == graph.Target.Namespace {
			continue
		}
		if node.Kind == "Pod" {
			continue
		}
		namespaces[node.Namespace] = true
		switch node.Kind {
		case "Ingress":
			hasIngress = true
		case "PodDisruptionBudget":
			hasPDB = true
		case "HorizontalPodAutoscaler":
			hasHPA = true
		}
	}

	for _, edge := range graph.Edges {
		if edge.Relationship == dependency.RelOwns {
			continue
		}
		if edge.Depth == 1 {
			direct++
		} else if edge.Depth > 1 {
			indirect++
		}
	}
	return
}

func evaluateRule(id, critLevel string, critMult float64, direct, indirect int, hasIngress, hasPDB, hasHPA bool, namespaceCount int) (float64, string) {
	switch id {
	case "namespace_criticality":
		return critMult, fmt.Sprintf("criticality: %s", strings.ToUpper(critLevel))
	case "direct_dependents":
		return math.Min(float64(direct)/5.0, 1.0), fmt.Sprintf("%d confirmed direct consumers", direct)
	case "indirect_dependents":
		return math.Min(float64(indirect)/10.0, 1.0), fmt.Sprintf("%d confirmed indirect consumers", indirect)
	case "ingress_exposed":
		if hasIngress {
			return 1.0, "Ingress routes traffic externally"
		}
		return 0.0, "No Ingress route detected"
	case "cross_namespace_impact":
		if namespaceCount > 1 {
			return 1.0, fmt.Sprintf("Blast radius spans %d namespaces", namespaceCount)
		}
		return 0.0, "Contained within single namespace"
	case "has_pdb":
		if hasPDB {
			return 1.0, "PDB constrains rollouts/scaling"
		}
		return 0.0, "No PDB present"
	case "has_hpa":
		if hasHPA {
			return 0.5, "HPA present, may auto-recover"
		}
		return 0.0, "No HPA active"
	}
	return 0.0, ""
}

func scoreToLevel(score float64) RiskLevel {
	switch {
	case score <= 3.0:
		return RiskLow
	case score <= 6.0:
		return RiskMedium
	case score <= 8.5:
		return RiskHigh
	default:
		return RiskCritical
	}
}

func computeUncertainty(graph *dependency.DependencyGraph) UncertaintyScore {
	level := UncertaintyHigh
	score := 0.75
	reasons := []string{"No Prometheus available; performing topology-only analysis"}

	hasEnvVarOnly := true
	for _, edge := range graph.Edges {
		if edge.Relationship != dependency.RelEnvRef && edge.Relationship != dependency.RelOwns {
			hasEnvVarOnly = false
			break
		}
	}
	if hasEnvVarOnly && len(graph.Edges) > 0 {
		level = UncertaintyVeryHigh
		score = 0.95
		reasons = append(reasons, "Inferences are based primarily on environment variable string matches")
	}

	return UncertaintyScore{Level: level, Reasons: reasons, Score: score}
}
