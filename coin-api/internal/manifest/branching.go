package manifest

// BranchingBundle is materialized from a pinned branching-model component version.
type BranchingBundle struct {
	Name    string
	Version string
	Rules   map[string]any
}
