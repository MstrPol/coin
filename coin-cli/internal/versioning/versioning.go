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

// NextRCNumber возвращает следующий номер release candidate для baseVersion.
// Например, если уже есть v1.5.0-rc.1 и v1.5.0-rc.2, вернёт 3.
func NextRCNumber(tagPrefix, baseVersion string) (int, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return 1, nil
	}

	pattern := regexp.MustCompile(
		`^` + regexp.QuoteMeta(tagPrefix+baseVersion+"-rc.") + `(\d+)$`,
	)

	tags, err := repo.Tags()
	if err != nil {
		return 1, err
	}

	max := 0
	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		m := pattern.FindStringSubmatch(ref.Name().Short())
		if m == nil {
			return nil
		}
		n, _ := strconv.Atoi(m[1])
		if n > max {
			max = n
		}
		return nil
	})
	return max + 1, nil
}

// LatestReleaseTag возвращает последний релизный тег, достижимый из HEAD,
// с другим minor/major относительно excludeMinor (используется в coinRelease).
//
// Совместим с Вариантом 2: RC-теги (vX.Y.Z-rc.N) являются финальными артефактами,
// поэтому функция ищет как чистые теги (v1.0.0), так и RC-теги (v1.0.0-rc.5).
// При сравнении двух тегов одной версии (v1.0.0 и v1.0.0-rc.5) побеждает чистый тег.
func LatestReleaseTag(tagPrefix, excludeMinor string) (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("не git-репозиторий: %w", err)
	}

	// Матчим как чистые теги (v1.0.0), так и RC (v1.0.0-rc.5)
	releasePattern := regexp.MustCompile(
		`^` + regexp.QuoteMeta(tagPrefix) + `(\d+)\.(\d+)\.(\d+)(-rc\.(\d+))?$`,
	)

	tags, err := repo.Tags()
	if err != nil {
		return "", err
	}

	var best string
	var bestParsed [4]int // maj, min, pat, rc (0 = чистый релиз → выше RC)
	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		m := releasePattern.FindStringSubmatch(name)
		if m == nil {
			return nil
		}
		// Пропускаем теги из того же minor-потока
		if excludeMinor != "" && strings.HasPrefix(name, tagPrefix+excludeMinor+".") {
			return nil
		}
		maj, _ := strconv.Atoi(m[1])
		min, _ := strconv.Atoi(m[2])
		pat, _ := strconv.Atoi(m[3])
		// rc = 0 означает чистый тег (выше любого RC), иначе номер RC
		rc := 0
		if m[5] != "" {
			n, _ := strconv.Atoi(m[5])
			rc = n
		}
		// rc=0 (чистый тег) считаем выше rc>0; для сравнения инвертируем:
		// используем "maxRC - rc" чтобы меньший rc.N не вытеснял больший
		parsed := [4]int{maj, min, pat, rc}
		if best == "" || compareSemverRC(parsed, bestParsed) > 0 {
			best = name
			bestParsed = parsed
		}
		return nil
	})
	return best, nil
}

// compareSemverRC сравнивает [maj, min, pat, rc] где rc=0 означает финальный тег
// (превосходит любой rc>0 той же версии).
func compareSemverRC(a, b [4]int) int {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			if a[i] > b[i] {
				return 1
			}
			return -1
		}
	}
	// Одинаковые maj.min.pat — сравниваем rc: 0 (финальный) > любой rc>0
	aRC, bRC := a[3], b[3]
	if aRC == 0 && bRC == 0 {
		return 0
	}
	if aRC == 0 {
		return 1 // финальный выше RC
	}
	if bRC == 0 {
		return -1
	}
	if aRC > bRC {
		return 1
	}
	if aRC < bRC {
		return -1
	}
	return 0
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
