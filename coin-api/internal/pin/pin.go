package pin

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

type Kind int

const (
	KindExact Kind = iota
	KindTilde
	KindCaret
	KindLatest
)

type Pin struct {
	Raw  string
	Kind Kind
	Base string // semver without operator prefix
}

func Parse(raw string) (Pin, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Pin{}, fmt.Errorf("empty pin")
	}
	switch {
	case raw == "*":
		return Pin{Raw: raw, Kind: KindLatest}, nil
	case strings.HasPrefix(raw, "="):
		base := strings.TrimSpace(strings.TrimPrefix(raw, "="))
		if base == "" {
			return Pin{}, fmt.Errorf("invalid exact pin %q", raw)
		}
		return Pin{Raw: raw, Kind: KindExact, Base: base}, nil
	case strings.HasPrefix(raw, "~"):
		base := strings.TrimSpace(strings.TrimPrefix(raw, "~"))
		if base == "" {
			return Pin{}, fmt.Errorf("invalid tilde pin %q", raw)
		}
		return Pin{Raw: raw, Kind: KindTilde, Base: base}, nil
	case strings.HasPrefix(raw, "^"):
		base := strings.TrimSpace(strings.TrimPrefix(raw, "^"))
		if base == "" {
			return Pin{}, fmt.Errorf("invalid caret pin %q", raw)
		}
		return Pin{Raw: raw, Kind: KindCaret, Base: base}, nil
	default:
		return Pin{Raw: raw, Kind: KindExact, Base: raw}, nil
	}
}

// Satisfies reports whether resolvedVersion matches the config pin.
func (p Pin) Satisfies(resolvedVersion string) bool {
	resolved := normSemver(resolvedVersion)
	base := normSemver(p.Base)
	switch p.Kind {
	case KindExact:
		return semver.Compare(resolved, base) == 0
	case KindTilde:
		if semver.Compare(resolved, base) < 0 {
			return false
		}
		return majorMinor(resolved) == majorMinor(base)
	case KindCaret:
		if semver.Compare(resolved, base) < 0 {
			return false
		}
		return majorOf(resolved) == majorOf(base)
	case KindLatest:
		return semver.IsValid(resolved)
	default:
		return false
	}
}

// SelectBest picks the highest semver from candidates that satisfies pin (for resolve engine).
func (p Pin) SelectBest(candidates []string, catalogLatest string) (string, error) {
	switch p.Kind {
	case KindExact:
		for _, v := range candidates {
			if semver.Compare(normSemver(v), normSemver(p.Base)) == 0 {
				return v, nil
			}
		}
		return "", fmt.Errorf("no release for exact pin %q", p.Raw)
	case KindLatest:
		if catalogLatest != "" {
			return catalogLatest, nil
		}
		return maxSemver(candidates)
	case KindTilde, KindCaret:
		var best string
		for _, v := range candidates {
			if !p.Satisfies(v) {
				continue
			}
			if best == "" || semver.Compare(normSemver(v), normSemver(best)) > 0 {
				best = v
			}
		}
		if best == "" {
			return "", fmt.Errorf("no published release satisfies pin %q", p.Raw)
		}
		return best, nil
	default:
		return "", fmt.Errorf("unsupported pin kind")
	}
}

// TildePointer returns mutable pointer key for a resolved version (e.g. 1.0.4 → ~1.0.0).
func TildePointer(resolvedVersion string) string {
	maj, min := majorMinorParts(resolvedVersion)
	return fmt.Sprintf("~%d.%d.0", maj, min)
}

// CaretPointer returns mutable pointer key for a resolved version (e.g. 1.0.4 → ^1.0.0).
func CaretPointer(resolvedVersion string) string {
	maj, _ := majorMinorParts(resolvedVersion)
	return fmt.Sprintf("^%d.0.0", maj)
}

func maxSemver(versions []string) (string, error) {
	var best string
	for _, v := range versions {
		nv := normSemver(v)
		if !semver.IsValid(nv) {
			continue
		}
		if best == "" || semver.Compare(nv, normSemver(best)) > 0 {
			best = v
		}
	}
	if best == "" {
		return "", fmt.Errorf("no semver candidates")
	}
	return best, nil
}

func normSemver(v string) string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}

func majorMinor(v string) string {
	maj, min := majorMinorParts(v)
	return fmt.Sprintf("v%d.%d", maj, min)
}

func majorOf(v string) string {
	maj, _ := majorMinorParts(v)
	return fmt.Sprintf("v%d", maj)
}

func majorMinorParts(v string) (major, minor int) {
	v = strings.TrimPrefix(normSemver(v), "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) >= 1 {
		fmt.Sscanf(parts[0], "%d", &major)
	}
	if len(parts) >= 2 {
		fmt.Sscanf(parts[1], "%d", &minor)
	}
	return major, minor
}

func IsSnapshotVersion(v string) bool {
	return strings.Contains(v, "-snapshot.")
}
