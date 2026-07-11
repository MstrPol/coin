package executor

import "strings"

// inferStageTarget selects a Containerfile/Dockerfile multi-stage name when the author
// model omits explicit target. Pilot convention: run.output or build.type drives the stage.
func inferStageTarget(kind, output, buildType string) string {
	switch kind {
	case "run":
		switch strings.TrimSpace(output) {
		case "validate", "test", "artifact":
			return output
		case "image":
			return "runtime"
		}
	case "build":
		switch strings.TrimSpace(buildType) {
		case "artifact":
			return "artifact"
		case "image", "liquibase-image":
			return "runtime"
		}
	}
	return ""
}
