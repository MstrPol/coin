package branching

import (
	"fmt"
	"os"
	"strings"
)

// ShouldPublish evaluates manifest branching publish policy for the current git context.
func ShouldPublish(m *Model, g GitContext) (bool, string) {
	if override := strings.TrimSpace(os.Getenv("COIN_PUBLISH_OVERRIDE")); override == "true" {
		return true, "COIN_PUBLISH_OVERRIDE=true"
	}
	if err := m.validateConfigured(); err != nil {
		return false, err.Error()
	}
	switch m.Publish.When {
	case "never":
		return false, "publish.when=never"
	case "always":
		return true, "publish.when=always"
	case "branch":
		if strings.TrimSpace(g.Branch) == m.Publish.Branch {
			return true, fmt.Sprintf("branch=%s", g.Branch)
		}
		return false, fmt.Sprintf("branch %q != publish.branch %q", g.Branch, m.Publish.Branch)
	case "tag":
		tag := strings.TrimSpace(g.TagName)
		if tag == "" {
			return false, "no tag on current commit"
		}
		if m.isSemverTagModel() {
			v := stripTagPrefix(tag, m.TagPrefix)
			if semverTagPattern.MatchString(v) {
				return true, fmt.Sprintf("semver tag %s", tag)
			}
			return false, fmt.Sprintf("tag %q is not a semver release tag", tag)
		}
		v := coinVersionFromTag(tag, m.TagPrefix)
		if v == "" {
			return false, fmt.Sprintf("tag %q does not match versioning format", tag)
		}
		if m.Qualifiers.RCEnabled && strings.Contains(v, "-rc-") {
			if m.Qualifiers.RCReleaseBranchesOnly && !strings.HasPrefix(g.Branch, "release/") {
				return false, "rc tag requires release/* branch"
			}
			return true, fmt.Sprintf("rc tag %s", tag)
		}
		return false, "publish requires rc tag (trunk-based policy)"
	default:
		return false, fmt.Sprintf("unsupported publish.when %q", m.Publish.When)
	}
}
