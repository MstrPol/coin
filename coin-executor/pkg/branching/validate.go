package branching

// ValidateBranch checks the current branch against ordered branch rules.
func ValidateBranch(m *Model, branch string) error {
	_, err := m.Match(branch)
	return err
}
