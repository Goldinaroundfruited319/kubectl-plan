package risk

type Rule struct {
	ID     string
	Name   string
	Weight int
	Phase  int
}

var RuleRegistry = []Rule{
	{ID: "namespace_criticality", Name: "Namespace Criticality", Weight: 20, Phase: 1},
	{ID: "direct_dependents", Name: "Direct Dependents", Weight: 30, Phase: 1},
	{ID: "indirect_dependents", Name: "Indirect Dependents", Weight: 15, Phase: 1},
	{ID: "ingress_exposed", Name: "Ingress Exposed", Weight: 10, Phase: 1},
	{ID: "cross_namespace_impact", Name: "Cross-Namespace Impact", Weight: 10, Phase: 1},
	{ID: "has_pdb", Name: "PodDisruptionBudget present", Weight: 10, Phase: 1},
	{ID: "has_hpa", Name: "HPA configured (may auto-recover)", Weight: 5, Phase: 1},
}
