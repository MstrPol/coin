package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrInvalidAgentMetadata = errors.New("invalid agent metadata")

type AgentMetadataFieldError struct {
	Field   string
	Message string
}

func (e AgentMetadataFieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

var agentDigestRE = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

func validateAgentMetadataForPromote(version string, metadata []byte) error {
	if version == "" {
		return fmt.Errorf("version is required")
	}
	if len(metadata) == 0 {
		return AgentMetadataFieldError{Field: "metadata.image", Message: "required"}
	}
	var meta map[string]any
	if err := json.Unmarshal(metadata, &meta); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAgentMetadata, err)
	}
	image, _ := meta["image"].(string)
	image = strings.TrimSpace(image)
	if image == "" {
		return AgentMetadataFieldError{Field: "metadata.image", Message: "required"}
	}
	expectedTag := ":" + version
	if !strings.HasSuffix(image, expectedTag) {
		return AgentMetadataFieldError{
			Field:   "metadata.image",
			Message: fmt.Sprintf("tag must equal version %q", version),
		}
	}
	digest, _ := meta["digest"].(string)
	digest = strings.TrimSpace(digest)
	if digest == "" {
		return AgentMetadataFieldError{Field: "metadata.digest", Message: "required"}
	}
	if !agentDigestRE.MatchString(digest) {
		return AgentMetadataFieldError{
			Field:   "metadata.digest",
			Message: "must match sha256: followed by 64 lowercase hex digits",
		}
	}
	return nil
}

func normalizeAgentMetadata(metadata json.RawMessage) json.RawMessage {
	if len(metadata) == 0 {
		return json.RawMessage(`{}`)
	}
	var meta map[string]any
	if err := json.Unmarshal(metadata, &meta); err != nil {
		return metadata
	}
	delete(meta, "goarch")
	delete(meta, "architecture")
	out, err := json.Marshal(meta)
	if err != nil {
		return metadata
	}
	return out
}
