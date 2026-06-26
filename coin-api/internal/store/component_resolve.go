package store

// ComponentResolveMode controls which component_versions.status values are visible during resolve.
type ComponentResolveMode string

const (
	// ComponentResolveStable — published components only (stable channel, agent/executor slots).
	ComponentResolveStable ComponentResolveMode = "stable"
	// ComponentResolveDraft — published + draft (GP draft edit, canary channel resolve).
	ComponentResolveDraft ComponentResolveMode = "draft"
)

// GPResolveOptions configures GP release loading for resolve.
type GPResolveOptions struct {
	AllowDraftGP  bool
	ComponentMode ComponentResolveMode
}

func allowedComponentStatuses(mode ComponentResolveMode) []string {
	switch mode {
	case ComponentResolveDraft:
		return []string{"published", "draft"}
	default:
		return []string{"published"}
	}
}

func componentStatusAllowed(status string, mode ComponentResolveMode) bool {
	for _, allowed := range allowedComponentStatuses(mode) {
		if status == allowed {
			return true
		}
	}
	// Legacy rows until migration 029 is applied.
	if status == "canary" {
		return mode == ComponentResolveDraft
	}
	return false
}

// componentResolveModeForGPDraftEdit selects valid component statuses when creating/updating GP drafts.
func componentResolveModeForGPDraftEdit(componentType string) ComponentResolveMode {
	if componentType == "agent" || componentType == "executor" {
		return ComponentResolveStable
	}
	return ComponentResolveDraft
}

// componentResolveModeForGPPromote requires all composition pins to be published.
func componentResolveModeForGPPromote(string) ComponentResolveMode {
	return ComponentResolveStable
}
