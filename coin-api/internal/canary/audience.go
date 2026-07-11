package canary

import (
	"crypto/sha256"
	"encoding/binary"
)

// ProjectBucket returns deterministic 0..99 bucket for percent rollout.
func ProjectBucket(projectName string) int {
	sum := sha256.Sum256([]byte(projectName))
	n := binary.BigEndian.Uint16(sum[:2])
	return int(n % 100)
}

// UseCanaryLine decides whether pin * should resolve to canary for a project.
func UseCanaryLine(projectName, projectMode string, canaryPercent int, enabled bool) bool {
	if !enabled {
		return false
	}
	switch projectMode {
	case "canary":
		return true
	case "stable":
		return false
	default:
		if projectName == "" {
			return false
		}
		return ProjectBucket(projectName) < canaryPercent
	}
}
