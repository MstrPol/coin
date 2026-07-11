package branching

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const defaultVersion = "0.0.1"

var semverOnly = regexp.MustCompile(`^v?(\d+\.\d+\.\d+)$`)

// ResolveVersion computes COIN_VERSION from matched rule template and git tags.
func ResolveVersion(m *Model, g GitContext) (string, error) {
	if override := strings.TrimSpace(os.Getenv("COIN_VERSION")); override != "" {
		return override, nil
	}
	if err := m.validateConfigured(); err != nil {
		return "", err
	}
	match, err := m.Match(g.Branch)
	if err != nil {
		return "", err
	}
	if tag := strings.TrimSpace(g.TagName); tag != "" {
		return coinVersionFromTag(tag), nil
	}
	tags, err := g.Tags("")
	if err != nil {
		return "", fmt.Errorf("list git tags: %w", err)
	}
	if v := latestFromTags(tags, match); v != "" {
		return v, nil
	}
	return defaultVersion, nil
}

func coinVersionFromTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ""
	}
	return strings.TrimPrefix(tag, "v")
}

func latestFromTags(tags []string, match MatchResult) string {
	template := match.Rule.Template
	if !strings.Contains(template, "{n}") {
		best := ""
		for _, raw := range tags {
			v := coinVersionFromTag(raw)
			if semverOnly.MatchString(raw) || semverOnly.MatchString("v"+v) {
				if best == "" || compareSemverBase(v, best) > 0 {
					best = v
				}
			}
		}
		return best
	}
	base := defaultVersion
	prefix := seriesPrefix(template, match.Captures, match.Rule, base)
	stripPrefix := strings.TrimPrefix(prefix, "v")
	maxN := 0
	best := ""
	for _, raw := range tags {
		v := coinVersionFromTag(raw)
		if !strings.HasPrefix(v, stripPrefix) {
			continue
		}
		suffix := strings.TrimPrefix(v, stripPrefix)
		n, err := strconv.Atoi(suffix)
		if err != nil || n <= maxN {
			continue
		}
		maxN = n
		best = v
	}
	return best
}

func renderTemplate(template string, caps map[string]string, rule BranchRule, base string, n int) string {
	out := template
	out = strings.ReplaceAll(out, "{base}", base)
	out = strings.ReplaceAll(out, "{branch}", rule.Name)
	if jira, ok := caps["jira"]; ok {
		out = strings.ReplaceAll(out, "{jira}", jira)
	}
	out = strings.ReplaceAll(out, "{n}", strconv.Itoa(n))
	return out
}

func seriesPrefix(template string, caps map[string]string, rule BranchRule, base string) string {
	return renderTemplate(strings.ReplaceAll(template, "{n}", ""), caps, rule, base, 0)
}

func compareSemverBase(a, b string) int {
	ap := strings.Split(strings.TrimPrefix(a, "v"), ".")
	bp := strings.Split(strings.TrimPrefix(b, "v"), ".")
	for len(ap) < 3 {
		ap = append(ap, "0")
	}
	for len(bp) < 3 {
		bp = append(bp, "0")
	}
	for i := 0; i < 3; i++ {
		av, _ := strconv.Atoi(ap[i])
		bv, _ := strconv.Atoi(bp[i])
		if av != bv {
			return av - bv
		}
	}
	return 0
}

// Bump computes the next git tag for the matched branch rule.
func Bump(m *Model, g GitContext, bumpType string) (string, error) {
	if err := m.validateConfigured(); err != nil {
		return "", err
	}
	match, err := m.Match(g.Branch)
	if err != nil {
		return "", err
	}
	current, err := ResolveVersion(m, g)
	if err != nil {
		return "", err
	}
	base := baseSemver(current)
	switch bumpType {
	case "major", "minor", "patch":
		base = bumpSemver(base, bumpType)
	case "rc", "snapshot", "none":
	default:
		return "", fmt.Errorf("unknown bump type %q", bumpType)
	}
	if !strings.Contains(match.Rule.Template, "{n}") {
		return renderTemplate(match.Rule.Template, match.Captures, match.Rule, base, 0), nil
	}
	prefix := seriesPrefix(match.Rule.Template, match.Captures, match.Rule, base)
	n := nextSeriesNumber(g, prefix)
	return renderTemplate(match.Rule.Template, match.Captures, match.Rule, base, n), nil
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

func nextSeriesNumber(g GitContext, prefix string) int {
	tags, err := g.Tags("")
	if err != nil {
		return 1
	}
	maxN := 0
	stripPrefix := strings.TrimPrefix(prefix, "v")
	for _, raw := range tags {
		v := coinVersionFromTag(raw)
		if !strings.HasPrefix(v, stripPrefix) {
			continue
		}
		suffix := strings.TrimPrefix(v, stripPrefix)
		n, err := strconv.Atoi(suffix)
		if err == nil && n > maxN {
			maxN = n
		}
	}
	return maxN + 1
}

// sortStrings is used by tests.
func sortStrings(ss []string) { sort.Strings(ss) }
