package versioning

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// jiraIDRe extracts the Jira ticket ID (first component) from a branch name.
//
//	release/PROJ-404        → PROJ-404
//	feature/PROJ-101-login  → PROJ-101  (slug after ID is ignored)
var jiraIDRe = regexp.MustCompile(`^[^/]+/([A-Z]+-\d+)`)

// BranchJiraID returns the Jira ticket identifier from a branch name.
// For trunk branches (main, master, detached) returns the branch name itself.
func BranchJiraID(branch string) string {
	if m := jiraIDRe.FindStringSubmatch(branch); m != nil {
		return m[1]
	}
	return branch
}

// rcTagRe matches: v1.5.0-PROJ-404-rc-2
var rcTagRe = regexp.MustCompile(`^v(\d+\.\d+\.\d+)-(.+)-rc-(\d+)$`)

// snapshotTagRe matches: v0.0.0-PROJ-101-snapshot-3
var snapshotTagRe = regexp.MustCompile(`^v(\d+\.\d+\.\d+)-(.+)-snapshot-(\d+)$`)

// CurrentVersion returns the version string for display.
//   - HEAD is tagged → that tag's version (e.g. "1.5.0-PROJ-404-rc-2")
//   - HEAD not tagged, tags exist → latest tag's version
//   - No tags at all → "0.0.1" (starting point for a new project)
func CurrentVersion() (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "0.0.1", nil
	}

	head, err := repo.Head()
	if err != nil {
		return "0.0.1", nil
	}

	// Prefer tag on HEAD
	if _, v := headTag(repo, head.Hash()); v != "" {
		return v, nil
	}

	// Fall back to latest tag in the repo
	if tag := latestTagAny(repo); tag != "" {
		return strings.TrimPrefix(tag, "v"), nil
	}

	return "0.0.1", nil
}

// NextVersionTag computes and returns the next tag to create.
//
// Logic:
//   - Find the latest existing series for (jiraID, versionType).
//     If found: continue that series (keep same base, increment N).
//   - If no existing series: compute new base = (latest_base + bump), N = 1.
//
// versionType must be "rc" or "snapshot".
// bumpLevel must be "major", "minor", or "patch".
func NextVersionTag(jiraID, bumpLevel, versionType string) (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		// No git repo — compute from scratch
		repo = nil
	}

	base, n := latestSeries(repo, jiraID, versionType)
	if base != "" {
		// Continue existing series: same base, next N
		return fmt.Sprintf("v%s-%s-%s-%d", base, jiraID, versionType, n+1), nil
	}

	// New series: bump from latest base
	currentBase := latestBase(repo)
	newBase, err := bump(currentBase, bumpLevel)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("v%s-%s-%s-1", newBase, jiraID, versionType), nil
}

// ─── internal helpers ────────────────────────────────────────────────────────

// latestSeries returns (base, maxN) for the latest series matching (jiraID, versionType).
// Returns ("", 0) if no such series exists.
func latestSeries(repo *git.Repository, jiraID, versionType string) (base string, maxN int) {
	if repo == nil {
		return "", 0
	}
	tags, _ := repo.Tags()

	var bestBase string
	var bestParsed [3]int
	var bestN int

	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		var tagBase, tagJiraID string
		var tagN int

		switch versionType {
		case "rc":
			m := rcTagRe.FindStringSubmatch(name)
			if m == nil {
				return nil
			}
			tagBase, tagJiraID, tagN = m[1], m[2], atoi(m[3])
		case "snapshot":
			m := snapshotTagRe.FindStringSubmatch(name)
			if m == nil {
				return nil
			}
			tagBase, tagJiraID, tagN = m[1], m[2], atoi(m[3])
		}

		if tagJiraID != jiraID {
			return nil
		}

		parsed := parseSemver(tagBase)
		if bestBase == "" || compareSemver(parsed, bestParsed) > 0 ||
			(compareSemver(parsed, bestParsed) == 0 && tagN > bestN) {
			bestBase = tagBase
			bestParsed = parsed
			bestN = tagN
		}
		return nil
	})

	return bestBase, bestN
}

// latestBase returns the major.minor.patch from the most recent RC or snapshot tag.
// Returns "0.0.0" if no tags found.
func latestBase(repo *git.Repository) string {
	if repo == nil {
		return "0.0.0"
	}
	tags, _ := repo.Tags()

	var best string
	var bestParsed [3]int

	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		var base string

		if m := rcTagRe.FindStringSubmatch(name); m != nil {
			base = m[1]
		} else if m := snapshotTagRe.FindStringSubmatch(name); m != nil {
			base = m[1]
		} else {
			return nil
		}

		parsed := parseSemver(base)
		if best == "" || compareSemver(parsed, bestParsed) > 0 {
			best = base
			bestParsed = parsed
		}
		return nil
	})

	if best == "" {
		return "0.0.0"
	}
	return best
}

// latestTagAny returns the name of the latest RC or snapshot tag in the repo.
func latestTagAny(repo *git.Repository) string {
	if repo == nil {
		return ""
	}
	tags, _ := repo.Tags()

	var best string
	var bestBase [3]int
	var bestN int
	var bestType int // 1=rc, 0=snapshot (rc wins same base/N)

	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		var base string
		var n, typ int

		if m := rcTagRe.FindStringSubmatch(name); m != nil {
			base, n, typ = m[1], atoi(m[3]), 1
		} else if m := snapshotTagRe.FindStringSubmatch(name); m != nil {
			base, n, typ = m[1], atoi(m[3]), 0
		} else {
			return nil
		}

		parsed := parseSemver(base)
		better := false
		switch {
		case best == "":
			better = true
		case compareSemver(parsed, bestBase) > 0:
			better = true
		case compareSemver(parsed, bestBase) == 0 && n > bestN:
			better = true
		case compareSemver(parsed, bestBase) == 0 && n == bestN && typ > bestType:
			better = true
		}
		if better {
			best = name
			bestBase = parsed
			bestN = n
			bestType = typ
		}
		return nil
	})

	return best
}

// headTag returns the tag name and version string if any RC/snapshot tag points to hash.
func headTag(repo *git.Repository, hash plumbing.Hash) (tag, version string) {
	tags, _ := repo.Tags()
	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		obj, err := repo.TagObject(ref.Hash())
		var commitHash plumbing.Hash
		if err == nil {
			commitHash = obj.Target
		} else {
			commitHash = ref.Hash()
		}
		if commitHash != hash {
			return nil
		}
		name := ref.Name().Short()
		if rcTagRe.MatchString(name) || snapshotTagRe.MatchString(name) {
			tag = name
			version = strings.TrimPrefix(name, "v")
		}
		return nil
	})
	return tag, version
}

// ─── semver helpers ──────────────────────────────────────────────────────────

var semverRe = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func bump(base, level string) (string, error) {
	if !semverRe.MatchString(base) {
		return "", fmt.Errorf("не удалось распарсить semver %q", base)
	}
	parts := strings.SplitN(base, ".", 3)
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch level {
	case "major":
		major++
		minor, patch = 0, 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("неизвестный уровень bump: %q (ожидается major, minor или patch)", level)
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

func parseSemver(s string) [3]int {
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return [3]int{}
	}
	var out [3]int
	for i, p := range parts {
		out[i], _ = strconv.Atoi(p)
	}
	return out
}

func compareSemver(a, b [3]int) int {
	for i := range a {
		if a[i] > b[i] {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
	}
	return 0
}

func safeName(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^0-9a-z.-]+`)
	s = re.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func dockerTag(version string) string {
	v := strings.ReplaceAll(version, "+", "-")
	re := regexp.MustCompile(`[^0-9A-Za-z_.-]+`)
	return re.ReplaceAllString(v, "-")
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// CurrentBranch returns the current git branch name.
// Checks BRANCH_NAME env var first (for CI), then git rev-parse.
func CurrentBranch() string {
	if env := os.Getenv("BRANCH_NAME"); env != "" {
		return env
	}
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "unknown"
	}
	head, err := repo.Head()
	if err != nil {
		return "unknown"
	}
	if head.Name().IsBranch() {
		return head.Name().Short()
	}
	return "detached"
}

// Kept for backward compatibility with architecture docs / release notes use-case.
func LatestReleaseTag(excludeBase string) (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("не git-репозиторий: %w", err)
	}

	tags, err := repo.Tags()
	if err != nil {
		return "", err
	}

	var best string
	var bestKey [4]int
	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		m := rcTagRe.FindStringSubmatch(name)
		if m == nil {
			return nil
		}
		if excludeBase != "" && m[1] == excludeBase {
			return nil
		}
		sem := parseSemver(m[1])
		n := atoi(m[3])
		key := [4]int{sem[0], sem[1], sem[2], n}
		if best == "" || compareKey(key, bestKey) > 0 {
			best = name
			bestKey = key
		}
		return nil
	})
	return best, nil
}

func compareKey(a, b [4]int) int {
	for i := range a {
		if a[i] > b[i] {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
	}
	return 0
}

// Compute is kept for internal use by `coin run` stages that need a full Result.
// Use CurrentVersion() for user-facing display.
func Compute(tagPrefix string) (*Result, error) {
	v, err := CurrentVersion()
	if err != nil {
		return &Result{Version: "0.0.1", ImageTag: "0.0.1", Source: "local"}, nil
	}
	return &Result{Version: v, ImageTag: dockerTag(v), Source: "computed"}, nil
}

// Result holds a computed version for a build (used by coin run stages).
type Result struct {
	Version  string
	ImageTag string
	Source   string
}
