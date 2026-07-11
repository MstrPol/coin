package branching

import (
	"fmt"
	"regexp"
	"strings"

	"coin.local/coin-executor/internal/manifest"
)

// BranchRule is one ordered branch policy entry (first match wins).
type BranchRule struct {
	Name     string
	Pattern  string
	Template string
	Publish  bool
	re       *regexp.Regexp
}

// Model is manifest.branching rules materialized at resolve time (schema v2).
type Model struct {
	Name     string
	Version  string
	Branches []BranchRule
}

// MatchResult is a matched branch rule with named captures.
type MatchResult struct {
	Rule     BranchRule
	Captures map[string]string
}

// PatternHint documents the RE2 named-group style used in v2 models.
const PatternHint = "Go RE2 named groups: (?P<jira>...); first matching rule wins"

// FromManifest builds a v2 branching model from manifest.branching.
func FromManifest(m *manifest.Manifest) *Model {
	if m == nil || m.Branching == nil || len(m.Branching.Branches) == 0 {
		return nil
	}
	b := m.Branching
	out := &Model{
		Name:    strings.TrimSpace(b.Name),
		Version: strings.TrimSpace(b.Version),
	}
	for _, br := range b.Branches {
		out.Branches = append(out.Branches, BranchRule{
			Name:     strings.TrimSpace(br.Name),
			Pattern:  strings.TrimSpace(br.Pattern),
			Template: strings.TrimSpace(br.Versioning.Template),
			Publish:  br.Publish,
		})
	}
	_ = out.compile()
	return out
}

func (m *Model) compile() error {
	if m == nil {
		return fmt.Errorf("branching model is nil")
	}
	for i := range m.Branches {
		pat := m.Branches[i].Pattern
		if pat == "" {
			return fmt.Errorf("branches[%d].pattern is required", i)
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return fmt.Errorf("branches[%d].pattern: %w", i, err)
		}
		m.Branches[i].re = re
		if strings.TrimSpace(m.Branches[i].Template) == "" {
			return fmt.Errorf("branches[%d].versioning.template is required", i)
		}
	}
	return nil
}

func (m *Model) validateConfigured() error {
	if m == nil {
		return fmt.Errorf("branching model is nil")
	}
	if len(m.Branches) == 0 {
		return fmt.Errorf("branching.branches must not be empty")
	}
	return m.compile()
}

// Match returns the first matching branch rule.
func (m *Model) Match(branch string) (MatchResult, error) {
	if err := m.validateConfigured(); err != nil {
		return MatchResult{}, err
	}
	branch = strings.TrimSpace(branch)
	for _, rule := range m.Branches {
		if rule.re == nil {
			continue
		}
		subs := rule.re.FindStringSubmatch(branch)
		if subs == nil {
			continue
		}
		caps := map[string]string{}
		for i, name := range rule.re.SubexpNames() {
			if i > 0 && name != "" {
				caps[name] = subs[i]
			}
		}
		return MatchResult{Rule: rule, Captures: caps}, nil
	}
	return MatchResult{}, fmt.Errorf("branch %q does not match any branching rule", branch)
}
