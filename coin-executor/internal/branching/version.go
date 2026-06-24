package branching

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	qualifierTagPattern = regexp.MustCompile(`^v?(\d+\.\d+\.\d+)-([A-Za-z0-9-]+)-(snapshot|rc)-(\d+)$`)
	semverTagPattern    = regexp.MustCompile(`^v?(\d+\.\d+\.\d+)$`)
)

const defaultVersion = "0.0.1"

// ResolveVersion computes COIN_VERSION from branching rules and git tags.
// COIN_VERSION env overrides when set (Jenkins pre-compute path).
func ResolveVersion(m *Model, g GitContext) (string, error) {
	if override := strings.TrimSpace(os.Getenv("COIN_VERSION")); override != "" {
		return override, nil
	}
	if err := m.validateConfigured(); err != nil {
		return "", err
	}
	if err := ValidateBranch(m, g.Branch); err != nil {
		return "", err
	}

	tags, err := g.Tags(m.TagPrefix)
	if err != nil {
		return "", fmt.Errorf("list git tags: %w", err)
	}
	if m.isSemverTagModel() {
		return resolveSemverTag(tags, m.TagPrefix), nil
	}
	return resolveTrunkBased(tags, m, g)
}

func resolveSemverTag(tags []string, prefix string) string {
	best := ""
	for _, raw := range tags {
		v := stripTagPrefix(raw, prefix)
		if semverTagPattern.MatchString(v) {
			if best == "" || compareSemverBase(v, best) > 0 {
				best = v
			}
		}
	}
	if best == "" {
		return defaultVersion
	}
	return strings.TrimPrefix(best, "v")
}

func resolveTrunkBased(tags []string, m *Model, g GitContext) (string, error) {
	jiraID := jiraIDFromBranch(g.Branch, m.TrunkBranch)
	if tag := strings.TrimSpace(g.TagName); tag != "" {
		if v := coinVersionFromTag(tag, m.TagPrefix); v != "" {
			return v, nil
		}
	}

	var matches []string
	for _, raw := range tags {
		v := coinVersionFromTag(raw, m.TagPrefix)
		if v == "" {
			continue
		}
		if strings.Contains(v, "-"+jiraID+"-") {
			matches = append(matches, v)
		}
	}
	if len(matches) == 0 {
		return defaultVersion, nil
	}
	sort.Strings(matches)
	return matches[len(matches)-1], nil
}

func coinVersionFromTag(tag, prefix string) string {
	v := stripTagPrefix(tag, prefix)
	if qualifierTagPattern.MatchString(v) || semverTagPattern.MatchString(v) {
		return strings.TrimPrefix(v, "v")
	}
	return ""
}

func stripTagPrefix(tag, prefix string) string {
	tag = strings.TrimSpace(tag)
	if prefix != "" && strings.HasPrefix(tag, prefix) {
		return tag[len(prefix):]
	}
	return tag
}

func jiraIDFromBranch(branch, trunk string) string {
	if branch == trunk {
		return "main"
	}
	match := branchPattern.FindStringSubmatch(branch)
	if match == nil {
		return "main"
	}
	return match[2]
}

func compareSemverBase(a, b string) int {
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")
	ap := strings.SplitN(a, ".", 3)
	bp := strings.SplitN(b, ".", 3)
	for i := 0; i < 3; i++ {
		av, bv := 0, 0
		if i < len(ap) {
			fmt.Sscanf(ap[i], "%d", &av)
		}
		if i < len(bp) {
			fmt.Sscanf(bp[i], "%d", &bv)
		}
		if av != bv {
			return av - bv
		}
	}
	return 0
}

// Bump computes the next git tag name for coin version bump (future CLI).
func Bump(m *Model, g GitContext, bumpType string) (string, error) {
	if err := m.validateConfigured(); err != nil {
		return "", err
	}
	if err := ValidateBranch(m, g.Branch); err != nil {
		return "", err
	}
	if m.isSemverTagModel() {
		return "", fmt.Errorf("semver-tag bump is not implemented in executor v1")
	}
	if bumpType == "rc" {
		if m.Qualifiers.RCReleaseBranchesOnly && !strings.HasPrefix(g.Branch, "release/") {
			return "", fmt.Errorf("rc bump is only allowed on release/* branches")
		}
	}
	current, err := ResolveVersion(m, g)
	if err != nil {
		return "", err
	}
	jiraID := jiraIDFromBranch(g.Branch, m.TrunkBranch)
	qual := "snapshot"
	if bumpType == "rc" {
		qual = "rc"
	}
	base := baseSemver(current)
	switch bumpType {
	case "major", "minor", "patch":
		base = bumpSemver(base, bumpType)
	case "rc", "snapshot", "none":
	default:
		return "", fmt.Errorf("unknown bump type %q", bumpType)
	}
	seriesPrefix := fmt.Sprintf("%s-%s-%s-", base, jiraID, qual)
	n := nextSeriesNumber(g, m.TagPrefix, seriesPrefix)
	return fmt.Sprintf("%s%s-%s-%s-%d", m.TagPrefix, base, jiraID, qual, n), nil
}

func baseSemver(version string) string {
	if i := strings.Index(version, "-"); i >= 0 {
		version = version[:i]
	}
	parts := strings.Split(version, ".")
	for len(parts) < 3 {
		parts = append(parts, "0")
	}
	return strings.Join(parts[:3], ".")
}

func bumpSemver(base, bumpType string) string {
	parts := strings.Split(base, ".")
	for len(parts) < 3 {
		parts = append(parts, "0")
	}
	var major, minor, patch int
	fmt.Sscanf(parts[0], "%d", &major)
	fmt.Sscanf(parts[1], "%d", &minor)
	fmt.Sscanf(parts[2], "%d", &patch)
	switch bumpType {
	case "major":
		return fmt.Sprintf("%d.0.0", major+1)
	case "minor":
		return fmt.Sprintf("%d.%d.0", major, minor+1)
	case "patch":
		return fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
	default:
		return base
	}
}

func nextSeriesNumber(g GitContext, prefix, seriesPrefix string) int {
	tags, err := g.Tags(prefix)
	if err != nil {
		return 1
	}
	maxN := 0
	for _, raw := range tags {
		v := stripTagPrefix(raw, prefix)
		if !strings.HasPrefix(v, seriesPrefix) {
			continue
		}
		suffix := strings.TrimPrefix(v, seriesPrefix)
		var n int
		if _, err := fmt.Sscanf(suffix, "%d", &n); err == nil && n > maxN {
			maxN = n
		}
	}
	return maxN + 1
}
