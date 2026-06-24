package branching

import (
	"fmt"
	"strings"

	"coin.local/coin-executor/internal/manifest"
)

// Model is manifest.branching rules materialized at resolve time.
type Model struct {
	Name        string
	Version     string
	TrunkBranch string
	BranchTypes []string
	TagPrefix   string
	Qualifiers  Qualifiers
	Publish     PublishPolicy
}

type Qualifiers struct {
	SnapshotEnabled           bool
	RCEnabled                 bool
	RCReleaseBranchesOnly     bool
}

type PublishPolicy struct {
	When   string // tag | branch | always | never
	Branch string
}

func FromManifest(m *manifest.Manifest) *Model {
	if m == nil || m.Branching == nil {
		return nil
	}
	b := m.Branching
	prefix := "v"
	if b.Versioning.TagPrefix != "" {
		prefix = b.Versioning.TagPrefix
	}
	return &Model{
		Name:        strings.TrimSpace(b.Name),
		Version:     strings.TrimSpace(b.Version),
		TrunkBranch: strings.TrimSpace(b.Trunk.Branch),
		BranchTypes: append([]string(nil), b.BranchTypes...),
		TagPrefix:   prefix,
		Qualifiers: Qualifiers{
			SnapshotEnabled:       b.Versioning.Qualifiers.Snapshot.Enabled,
			RCEnabled:             b.Versioning.Qualifiers.RC.Enabled,
			RCReleaseBranchesOnly: b.Versioning.Qualifiers.RC.ReleaseBranchesOnly,
		},
		Publish: PublishPolicy{
			When:   strings.TrimSpace(b.Publish.When),
			Branch: strings.TrimSpace(b.Publish.Branch),
		},
	}
}

func (m *Model) validateConfigured() error {
	if m == nil {
		return fmt.Errorf("branching model is nil")
	}
	if m.TrunkBranch == "" {
		return fmt.Errorf("branching.trunk.branch is required")
	}
	if len(m.BranchTypes) == 0 {
		return fmt.Errorf("branching.branchTypes must not be empty")
	}
	switch m.Publish.When {
	case "tag", "branch", "always", "never":
	default:
		return fmt.Errorf("branching.publish.when must be tag, branch, always or never")
	}
	if m.Publish.When == "branch" && m.Publish.Branch == "" {
		return fmt.Errorf("branching.publish.branch is required when publish.when is branch")
	}
	return nil
}

func (m *Model) isSemverTagModel() bool {
	return m.Name == "semver-tag"
}
