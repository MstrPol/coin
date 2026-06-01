package versioning

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Result struct {
	Version string
	ImageTag string
	Source   string
}

func Compute(tagPrefix string) (*Result, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return fallback(), nil
	}

	head, err := repo.Head()
	if err != nil {
		return fallback(), nil
	}

	sha := head.Hash().String()
	if len(sha) > 8 {
		sha = sha[:8]
	}
	build := os.Getenv("BUILD_NUMBER")
	if build == "" {
		build = "0"
	}

	// Проверяем: HEAD помечен релизным тегом?
	tags, _ := repo.Tags()
	var matchedTag string
	releasePattern := regexp.MustCompile(
		`^` + regexp.QuoteMeta(tagPrefix) + `\d+\.\d+\.\d+([-.][0-9A-Za-z.-]+)?$`,
	)
	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		obj, err := repo.TagObject(ref.Hash())
		var commitHash plumbing.Hash
		if err == nil {
			commitHash = obj.Target
		} else {
			commitHash = ref.Hash()
		}
		if commitHash == head.Hash() {
			name := ref.Name().Short()
			if releasePattern.MatchString(name) {
				matchedTag = name
			}
		}
		return nil
	})

	if matchedTag != "" {
		version := strings.TrimPrefix(matchedTag, tagPrefix)
		return &Result{
			Version:  version,
			ImageTag: dockerTag(version),
			Source:   "tag:" + matchedTag,
		}, nil
	}

	// Определяем ветку
	branch := "detached"
	if head.Name().IsBranch() {
		branch = head.Name().Short()
	} else if envBranch := os.Getenv("BRANCH_NAME"); envBranch != "" {
		branch = envBranch
	}

	// release/<JIRA-ID> → release candidate
	if strings.HasPrefix(branch, "release/") {
		safeID := safeName(strings.TrimPrefix(branch, "release/"))
		version := fmt.Sprintf("0.0.0-rc.%s.%s+%s", safeID, build, sha)
		return &Result{
			Version:  version,
			ImageTag: dockerTag(version),
			Source:   "release-branch:" + branch,
		}, nil
	}

	safeBranch := safeName(branch)
	version := fmt.Sprintf("0.0.0-%s.%s+%s", safeBranch, build, sha)
	return &Result{
		Version:  version,
		ImageTag: dockerTag(version),
		Source:   fmt.Sprintf("branch:%s:%s", branch, sha),
	}, nil
}

// LatestReleaseTag возвращает последний релизный тег, достижимый из HEAD,
// с другим minor/major относительно currentMinor (используется в coinRelease).
func LatestReleaseTag(tagPrefix, excludeMinor string) (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("не git-репозиторий: %w", err)
	}

	releasePattern := regexp.MustCompile(
		`^` + regexp.QuoteMeta(tagPrefix) + `(\d+)\.(\d+)\.(\d+)$`,
	)

	tags, err := repo.Tags()
	if err != nil {
		return "", err
	}

	var best string
	var bestParsed [3]int
	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		if !releasePattern.MatchString(name) {
			return nil
		}
		// Пропускаем теги из того же minor-потока
		if excludeMinor != "" && strings.HasPrefix(name, tagPrefix+excludeMinor+".") {
			return nil
		}
		var maj, min, pat int
		fmt.Sscanf(strings.TrimPrefix(name, tagPrefix), "%d.%d.%d", &maj, &min, &pat)
		if best == "" ||
			maj > bestParsed[0] ||
			(maj == bestParsed[0] && min > bestParsed[1]) ||
			(maj == bestParsed[0] && min == bestParsed[1] && pat > bestParsed[2]) {
			best = name
			bestParsed = [3]int{maj, min, pat}
		}
		return nil
	})
	return best, nil
}

func safeName(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^0-9a-z.-]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func dockerTag(version string) string {
	v := strings.ReplaceAll(version, "+", "-")
	re := regexp.MustCompile(`[^0-9A-Za-z_.-]+`)
	return re.ReplaceAllString(v, "-")
}

func fallback() *Result {
	build := os.Getenv("BUILD_NUMBER")
	if build == "" {
		build = "0"
	}
	version := "0.0.0-local." + build
	return &Result{Version: version, ImageTag: version, Source: "local"}
}
