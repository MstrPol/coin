package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

var agentRevSuffix = regexp.MustCompile(`^(.+)-r(\d+)$`)

// NextAgentVersion returns the next {runtime}-r{N} for agent/{stack}.
// Only versions matching {runtime}-r{digits} are considered.
func (s *Store) NextAgentVersion(ctx context.Context, stack, runtime string) (currentRev, nextRev int, nextVersion string, err error) {
	if stack == "" || runtime == "" {
		return 0, 0, "", fmt.Errorf("stack and runtime are required")
	}

	versions, err := s.ListComponentVersions(ctx, "agent", stack)
	if errors.Is(err, ErrNotFound) {
		versions = nil
	} else if err != nil {
		return 0, 0, "", err
	}

	maxRev := 0
	for _, v := range versions {
		m := agentRevSuffix.FindStringSubmatch(v.Version)
		if m == nil || m[1] != runtime {
			continue
		}
		n, parseErr := strconv.Atoi(m[2])
		if parseErr != nil {
			continue
		}
		if n > maxRev {
			maxRev = n
		}
	}

	currentRev = maxRev
	nextRev = maxRev + 1
	nextVersion = fmt.Sprintf("%s-r%d", runtime, nextRev)
	return currentRev, nextRev, nextVersion, nil
}

const semverComponentInitialVersion = "1.0.0"

var semverThreePart = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

func validateSemverBump(bump string) error {
	switch bump {
	case "major", "minor", "patch":
		return nil
	default:
		return fmt.Errorf("bump must be major, minor, or patch")
	}
}

func (s *Store) nextSemverComponentVersion(
	ctx context.Context,
	typ, name, bump string,
) (currentVersion, nextVersion string, isFirst bool, err error) {
	if name == "" {
		return "", "", false, fmt.Errorf("component name is required")
	}
	if err := validateSemverBump(bump); err != nil {
		return "", "", false, err
	}

	versions, err := s.ListComponentVersions(ctx, typ, name)
	if errors.Is(err, ErrNotFound) {
		return "", semverComponentInitialVersion, true, nil
	}
	if err != nil {
		return "", "", false, err
	}

	latest := latestPublishedSemver(versions)
	if latest == "" {
		return "", semverComponentInitialVersion, true, nil
	}

	next, err := bumpSemver(latest, bump)
	if err != nil {
		return "", "", false, err
	}
	return latest, next, false, nil
}

// NextGPContentVersion returns semver after bump for gp-content/{name}.
func (s *Store) NextGPContentVersion(ctx context.Context, name, bump string) (currentVersion, nextVersion string, isFirst bool, err error) {
	if name == "" {
		return "", "", false, fmt.Errorf("gp-content name is required")
	}
	return s.nextSemverComponentVersion(ctx, "gp-content", name, bump)
}

// NextExecutorVersion returns semver after bump for executor/coin-executor.
// First publish → 1.0.0.
func (s *Store) NextExecutorVersion(ctx context.Context, name, bump string) (currentVersion, nextVersion string, isFirst bool, err error) {
	if name == "" {
		return "", "", false, fmt.Errorf("executor name is required")
	}
	return s.nextSemverComponentVersion(ctx, "executor", name, bump)
}

func latestPublishedSemver(versions []ComponentVersionListItem) string {
	var best string
	for _, v := range versions {
		if v.Status != "published" {
			continue
		}
		key := semverCanonicalKey(v.Version)
		if key == "" {
			continue
		}
		if best == "" || semver.Compare(key, semverCanonicalKey(best)) > 0 {
			best = strings.TrimPrefix(strings.TrimSpace(v.Version), "v")
		}
	}
	return best
}

func semverCanonicalKey(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		return ""
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if !semver.IsValid(v) {
		return ""
	}
	return semver.Canonical(v)
}

func bumpSemver(current, bump string) (string, error) {
	v := strings.TrimPrefix(strings.TrimSpace(current), "v")
	m := semverThreePart.FindStringSubmatch(v)
	if m == nil {
		return "", fmt.Errorf("invalid semver %q", current)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	switch bump {
	case "major":
		return fmt.Sprintf("%d.0.0", major+1), nil
	case "minor":
		return fmt.Sprintf("%d.%d.0", major, minor+1), nil
	case "patch":
		return fmt.Sprintf("%d.%d.%d", major, minor, patch+1), nil
	default:
		return "", fmt.Errorf("unsupported bump %q", bump)
	}
}
