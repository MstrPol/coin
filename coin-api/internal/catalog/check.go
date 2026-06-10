package catalog

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

var ErrBelowMinimum = errors.New("gp version below catalog minimum")

type Policy struct {
	GPName       string
	Latest       string
	LatestCanary string
	Minimum      string
	Deprecated   []string
}

func Check(policy Policy, version string) (warning string, err error) {
	if policy.Minimum != "" {
		if semver.Compare(norm(version), norm(policy.Minimum)) < 0 {
			return "", fmt.Errorf("%w: %s@%s minimum is %s", ErrBelowMinimum, policy.GPName, version, policy.Minimum)
		}
	}
	for _, dep := range policy.Deprecated {
		if dep == version {
			return fmt.Sprintf(`299 - "GP %s version %s is deprecated"`, policy.GPName, version), nil
		}
	}
	return "", nil
}

func norm(v string) string {
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}
