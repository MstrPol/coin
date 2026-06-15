package store

import (
	"context"
	"sort"

	"golang.org/x/mod/semver"
)

type VersionCount struct {
	Version string `json:"version"`
	Count   int    `json:"count"`
}

type BlastRadius struct {
	GoldenPath      string         `json:"goldenPath"`
	Version         string         `json:"version"`
	OnThisVersion   int            `json:"onThisVersion"`
	OnOtherVersions int            `json:"onOtherVersions"`
	OnOlderVersions int            `json:"onOlderVersions"`
	TotalOnGP       int            `json:"totalOnGP"`
	ByVersion       []VersionCount `json:"byVersion"`
}

func (s *Store) BlastRadius(ctx context.Context, gpName, targetVersion string) (BlastRadius, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (br.project_id) COALESCE(br.resolved_version, br.gp_version)
		FROM build_reports br
		WHERE br.gp_name = $1
		ORDER BY br.project_id, br.reported_at DESC
	`, gpName)
	if err != nil {
		return BlastRadius{}, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return BlastRadius{}, err
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		return BlastRadius{}, err
	}

	return computeBlastRadius(gpName, targetVersion, versions), nil
}

func computeBlastRadius(gpName, targetVersion string, projectVersions []string) BlastRadius {
	counts := map[string]int{}
	for _, v := range projectVersions {
		counts[v]++
	}

	result := BlastRadius{
		GoldenPath: gpName,
		Version:    targetVersion,
		TotalOnGP:  len(projectVersions),
	}

	targetNorm := normSemver(targetVersion)
	for ver, count := range counts {
		result.ByVersion = append(result.ByVersion, VersionCount{Version: ver, Count: count})
		if ver == targetVersion {
			result.OnThisVersion += count
			continue
		}
		result.OnOtherVersions += count
		if semver.Compare(normSemver(ver), targetNorm) < 0 {
			result.OnOlderVersions += count
		}
	}

	sort.Slice(result.ByVersion, func(i, j int) bool {
		return semver.Compare(normSemver(result.ByVersion[i].Version), normSemver(result.ByVersion[j].Version)) > 0
	})

	if result.ByVersion == nil {
		result.ByVersion = []VersionCount{}
	}
	return result
}

func normSemver(v string) string {
	if v == "" {
		return "v0.0.0"
	}
	if v[0] == 'v' {
		return v
	}
	return "v" + v
}
