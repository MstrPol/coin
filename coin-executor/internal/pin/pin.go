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
	Base string
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
