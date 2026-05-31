package dependency

func NewDependencyGraph(target Node) *DependencyGraph {
	return &DependencyGraph{
		Target: target,
		Nodes:  make(map[string]Node),
		Edges:  []Edge{},
	}
}

func (dg *DependencyGraph) AddNode(n Node) {
	key := NodeKey(n.Namespace, n.Kind, n.Name)
	dg.Nodes[key] = n
}

func (dg *DependencyGraph) AddEdge(e Edge) {
	dg.Edges = append(dg.Edges, e)
}

func NodeKey(namespace, kind, name string) string {
	return namespace + "/" + kind + "/" + name
}
