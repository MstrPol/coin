package branching

import (
	"fmt"
	"os"
	"strings"
)

// CheckPublishAllowed enforces per-branch publish eligibility when publish is requested.
// Returns nil when publish is not requested or branch allows publish.
func CheckPublishAllowed(m *Model, g GitContext) error {
	if strings.TrimSpace(os.Getenv("COIN_PUBLISH_REQUEST")) != "true" {
		return nil
	}
	if err := m.validateConfigured(); err != nil {
		return err
	}
	match, err := m.Match(g.Branch)
	if err != nil {
		return err
	}
	if !match.Rule.Publish {
		return fmt.Errorf("publish not allowed on branch %q (rule %q has publish=false)", g.Branch, match.Rule.Name)
	}
	return nil
}

// PublishOutcome describes preview publish evaluation.
type PublishOutcome string

const (
	PublishNotRequested PublishOutcome = "not-requested"
	PublishAllowed      PublishOutcome = "allowed"
	PublishDenied       PublishOutcome = "denied"
)

// EvaluatePublish returns preview publish outcome for a scenario.
func EvaluatePublish(m *Model, branch string, requestPublish bool) (PublishOutcome, string) {
	if !requestPublish {
		return PublishNotRequested, "publish not requested"
	}
	match, err := m.Match(branch)
	if err != nil {
		return PublishDenied, err.Error()
	}
	if match.Rule.Publish {
		return PublishAllowed, fmt.Sprintf("rule %q allows publish", match.Rule.Name)
	}
	return PublishDenied, fmt.Sprintf("rule %q has publish=false", match.Rule.Name)
}
