package store

// ComponentResolveMode controls which component_versions.status values are visible during resolve.
type ComponentResolveMode string

const (
	// ComponentResolveStable — product CI on stable channel: published only.
	ComponentResolveStable ComponentResolveMode = "stable"
	// ComponentResolveCanary — canary channel / pilot projects: published + canary.
	ComponentResolveCanary ComponentResolveMode = "canary"
	// ComponentResolveAdmin — admin GP draft preview: published + canary + draft.
	ComponentResolveAdmin ComponentResolveMode = "admin"
)

// GPResolveOptions configures GP release loading for resolve.
type GPResolveOptions struct {
	AllowDraftGP  bool
	ComponentMode ComponentResolveMode
}

func allowedComponentStatuses(mode ComponentResolveMode) []string {
	switch mode {
	case ComponentResolveCanary:
		return []string{"published", "canary"}
	case ComponentResolveAdmin:
		return []string{"published", "canary", "draft"}
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
	return false
}
