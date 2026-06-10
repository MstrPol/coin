package compatibility

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

type Requirement struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	Min           string `json:"min"`
	MaxExclusive  string `json:"maxExclusive"`
}

type Rule struct {
	SourceType        string
	SourceName        string
	VersionPrefix     string
	Requirements      map[string]Requirement
}

type CompositionSlot struct {
	Key  string
	Type string
	Name string
}

func Validate(slots []CompositionSlot, composition map[string]string, rules []Rule) error {
	if len(composition) == 0 {
		return fmt.Errorf("composition is required")
	}

	byKey := make(map[string]CompositionSlot, len(slots))
	for _, slot := range slots {
		byKey[slot.Key] = slot
	}

	for _, slot := range slots {
		ver, ok := composition[slot.Key]
		if !ok || ver == "" {
			return fmt.Errorf("composition.%s is required", slot.Key)
		}
	}

	for _, rule := range rules {
		slot, ok := byKey["pipeline"]
		if !ok {
			continue
		}
		if rule.SourceType != slot.Type || rule.SourceName != slot.Name {
			continue
		}
		pipelineVer := composition["pipeline"]
		if !versionHasPrefix(pipelineVer, rule.VersionPrefix) {
			continue
		}
		for key, req := range rule.Requirements {
			ver, ok := composition[key]
			if !ok {
				return fmt.Errorf("compatibility: composition.%s required by pipeline %s", key, pipelineVer)
			}
			slot := byKey[key]
			if slot.Type != req.Type || slot.Name != req.Name {
				return fmt.Errorf("compatibility: slot %s mismatch for rule", key)
			}
			if err := checkRange(ver, req.Min, req.MaxExclusive); err != nil {
				return fmt.Errorf("compatibility: %s@%s: %w", key, ver, err)
			}
		}
	}
	return nil
}

func versionHasPrefix(version, prefix string) bool {
	if prefix == "" {
		return true
	}
	if strings.HasSuffix(prefix, ".") {
		return strings.HasPrefix(version, prefix)
	}
	return version == prefix || strings.HasPrefix(version, prefix+".")
}

func checkRange(version, min, maxExclusive string) error {
	v := norm(version)
	if min != "" && semver.Compare(v, norm(min)) < 0 {
		return fmt.Errorf("version %s below minimum %s", version, min)
	}
	if maxExclusive != "" && semver.Compare(v, norm(maxExclusive)) >= 0 {
		return fmt.Errorf("version %s must be below %s", version, maxExclusive)
	}
	return nil
}

func norm(v string) string {
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}

func ParseRequirements(raw json.RawMessage) (map[string]Requirement, error) {
	var reqs map[string]Requirement
	if err := json.Unmarshal(raw, &reqs); err != nil {
		return nil, err
	}
	return reqs, nil
}
