package dependency

type DependencyGraph struct {
	Target Node
	Nodes  map[string]Node
	Edges  []Edge
}

type Node struct {
	Kind      string
	Name      string
	Namespace string
	Labels    map[string]string
	Metadata  map[string]string
}

type Edge struct {
	From         string
	To           string
	Relationship RelationshipType
	Depth        int
	Evidence     []Evidence
	Confidence   float64
	Uncertainty  float64
}

type RelationshipType string

const (
	RelOwns          RelationshipType = "OWNS"
	RelSelects       RelationshipType = "SELECTS"
	RelRoutes        RelationshipType = "ROUTES"
	RelEnvRef        RelationshipType = "ENV_REF"
	RelVolumeRef     RelationshipType = "VOLUME_REF"
	RelNetworkPolicy RelationshipType = "NETWORK_POLICY"
	RelCronDepends   RelationshipType = "CRON_DEPENDS"
	RelTraffic       RelationshipType = "TRAFFIC"
)
