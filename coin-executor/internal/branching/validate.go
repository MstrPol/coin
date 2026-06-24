package branching

import (
	"fmt"
	"regexp"
	"strings"
)

var branchPattern = regexp.MustCompile(`^([a-z]+)/([A-Z][A-Z0-9]*-[0-9]+)(?:-.*)?$`)

// ValidateBranch checks the current branch against model branch types and naming rules.
func ValidateBranch(m *Model, branch string) error {
	if err := m.validateConfigured(); err != nil {
		return err
	}
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return fmt.Errorf("git branch is empty")
	}
	if branch == m.TrunkBranch {
		return nil
	}
	match := branchPattern.FindStringSubmatch(branch)
	if match == nil {
		return fmt.Errorf("branch %q must be %s or <type>/<JIRA-ID>[-slug]", branch, m.TrunkBranch)
	}
	typ := match[1]
	if !containsString(m.BranchTypes, typ) {
		return fmt.Errorf("branch type %q is not allowed (allowed: %s)", typ, strings.Join(m.BranchTypes, ", "))
	}
	return nil
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
