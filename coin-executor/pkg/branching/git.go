package branching

import (
	"os"
	"os/exec"
	"strings"
)

// GitContext captures branch/tag state for branching policy.
type GitContext struct {
	Branch  string
	TagName string
	WorkDir string
	tags    []string // optional synthetic tags for preview
}

// GitFromEnv loads branch/tag from CI environment and git.
func GitFromEnv(workDir string) (GitContext, error) {
	g := GitContext{WorkDir: workDir}
	// PR builds: source branch when multibranch sets CHANGE_BRANCH.
	g.Branch = firstNonEmpty(
		os.Getenv("CHANGE_BRANCH"),
		os.Getenv("GIT_BRANCH"),
		os.Getenv("BRANCH_NAME"),
	)
	g.TagName = firstNonEmpty(os.Getenv("TAG_NAME"), os.Getenv("GIT_TAG_NAME"))

	if g.Branch == "" {
		out, err := gitOutput(workDir, "rev-parse", "--abbrev-ref", "HEAD")
		if err == nil && out != "HEAD" {
			g.Branch = out
		}
	}
	if g.TagName == "" {
		out, err := gitOutput(workDir, "describe", "--tags", "--exact-match", "HEAD")
		if err == nil {
			g.TagName = out
		}
	}
	return g, nil
}

// WithSyntheticTags returns a copy using in-memory tags (preview API).
func (g GitContext) WithSyntheticTags(tags []string) GitContext {
	g.tags = append([]string(nil), tags...)
	return g
}

func (g GitContext) Tags(prefix string) ([]string, error) {
	if g.tags != nil {
		if prefix == "" {
			return append([]string(nil), g.tags...), nil
		}
		var out []string
		for _, t := range g.tags {
			if strings.HasPrefix(t, prefix) {
				out = append(out, t)
			}
		}
		return out, nil
	}
	pattern := prefix + "*"
	if prefix == "" {
		pattern = "*"
	}
	out, err := gitOutput(g.WorkDir, "tag", "-l", pattern)
	if err != nil {
		return nil, nil
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(out), "\n"), nil
}

func gitOutput(workDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
