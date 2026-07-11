package branching

import (
	"fmt"
	"strings"
)

// PreviewScenario is one synthetic evaluation input.
type PreviewScenario struct {
	ID              string
	Branch          string
	TagName         string
	Tags            []string
	RequestPublish  bool
}

// PreviewScenarioResult is evaluation output for one scenario.
type PreviewScenarioResult struct {
	ID              string   `json:"id"`
	MatchedRule     string   `json:"matchedRule,omitempty"`
	BranchValid     bool     `json:"branchValid"`
	BranchError     string   `json:"branchError,omitempty"`
	CoinVersion     string   `json:"coinVersion,omitempty"`
	VersionError    string   `json:"versionError,omitempty"`
	PublishOutcome  string   `json:"publishOutcome"`
	PublishReason   string   `json:"publishReason,omitempty"`
}

// PreviewResult is the full preview API response payload.
type PreviewResult struct {
	PatternHint string                  `json:"patternHint"`
	Results     []PreviewScenarioResult `json:"results"`
}

// ModelFromDoc builds a model from API/preview JSON (schema v2 subset).
type ModelDoc struct {
	Name     string          `json:"name"`
	Branches []BranchRuleDoc `json:"branches"`
}

type BranchRuleDoc struct {
	Name       string `json:"name"`
	Pattern    string `json:"pattern"`
	Versioning struct {
		Template string `json:"template"`
	} `json:"versioning"`
	Publish bool `json:"publish"`
}

func ModelFromDoc(doc ModelDoc) (*Model, error) {
	m := &Model{
		Name: strings.TrimSpace(doc.Name),
	}
	for _, br := range doc.Branches {
		m.Branches = append(m.Branches, BranchRule{
			Name:     strings.TrimSpace(br.Name),
			Pattern:  strings.TrimSpace(br.Pattern),
			Template: strings.TrimSpace(br.Versioning.Template),
			Publish:  br.Publish,
		})
	}
	if err := m.validateConfigured(); err != nil {
		return nil, err
	}
	return m, nil
}

// PreviewScenarios evaluates branching policy for synthetic scenarios.
func PreviewScenarios(m *Model, scenarios []PreviewScenario) (PreviewResult, error) {
	if err := m.validateConfigured(); err != nil {
		return PreviewResult{}, err
	}
	out := PreviewResult{
		PatternHint: PatternHint,
		Results:     make([]PreviewScenarioResult, 0, len(scenarios)),
	}
	for _, sc := range scenarios {
		res := PreviewScenarioResult{ID: sc.ID}
		match, err := m.Match(sc.Branch)
		if err != nil {
			res.BranchValid = false
			res.BranchError = err.Error()
		} else {
			res.BranchValid = true
			res.MatchedRule = match.Rule.Name
			g := GitContext{Branch: sc.Branch, TagName: sc.TagName}.WithSyntheticTags(sc.Tags)
			v, verr := ResolveVersion(m, g)
			if verr != nil {
				res.VersionError = verr.Error()
			} else {
				res.CoinVersion = v
			}
		}
		outcome, reason := EvaluatePublish(m, sc.Branch, sc.RequestPublish)
		res.PublishOutcome = string(outcome)
		res.PublishReason = reason
		out.Results = append(out.Results, res)
	}
	return out, nil
}

// FormatPreviewError wraps preview validation errors.
func FormatPreviewError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%v", err)
}
