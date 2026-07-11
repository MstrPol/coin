package store

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AgentImageParsed holds repository name and tag from a container image reference.
type AgentImageParsed struct {
	Repository string
	Tag        string
}

func ParseAgentImageRef(image string) (AgentImageParsed, error) {
	image = strings.TrimSpace(image)
	if image == "" {
		return AgentImageParsed{}, AgentMetadataFieldError{Field: "metadata.image", Message: "required"}
	}
	if at := strings.Index(image, "@sha256:"); at >= 0 {
		image = image[:at]
	}
	repoTag := image
	if slash := strings.LastIndex(image, "/"); slash >= 0 {
		repoTag = image[slash+1:]
	}
	colon := strings.LastIndex(repoTag, ":")
	if colon < 0 || colon == len(repoTag)-1 {
		return AgentImageParsed{}, AgentMetadataFieldError{
			Field:   "metadata.image",
			Message: "must include image tag after repository name",
		}
	}
	repo := repoTag[:colon]
	tag := repoTag[colon+1:]
	if repo == "" || tag == "" {
		return AgentImageParsed{}, AgentMetadataFieldError{Field: "metadata.image", Message: "invalid image reference"}
	}
	if tag == "latest" {
		return AgentImageParsed{}, AgentMetadataFieldError{Field: "metadata.image", Message: "tag latest is not allowed"}
	}
	if strings.HasPrefix(tag, "sha256:") {
		return AgentImageParsed{}, AgentMetadataFieldError{
			Field:   "metadata.image",
			Message: "digest-only reference requires a version tag",
		}
	}
	return AgentImageParsed{Repository: repo, Tag: tag}, nil
}

func resolveAgentDraftVersion(profileName, requestedVersion string, metadata json.RawMessage) (string, json.RawMessage, error) {
	if len(metadata) == 0 {
		return "", nil, AgentMetadataFieldError{Field: "metadata.image", Message: "required"}
	}
	var meta map[string]any
	if err := json.Unmarshal(metadata, &meta); err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrInvalidAgentMetadata, err)
	}
	image, _ := meta["image"].(string)
	parsed, err := ParseAgentImageRef(image)
	if err != nil {
		return "", nil, err
	}
	if parsed.Repository != profileName {
		return "", nil, AgentMetadataFieldError{
			Field:   "metadata.image",
			Message: fmt.Sprintf("repository %q must match profile %q", parsed.Repository, profileName),
		}
	}
	if requestedVersion != "" && requestedVersion != parsed.Tag {
		return "", nil, AgentMetadataFieldError{
			Field:   "version",
			Message: fmt.Sprintf("must match image tag %q", parsed.Tag),
		}
	}
	normalized := normalizeAgentMetadata(metadata)
	if err := validateAgentMetadataDigestIfPresent(normalized); err != nil {
		return "", nil, err
	}
	return parsed.Tag, normalized, nil
}

func validateAgentMetadataForPatch(version string, metadata json.RawMessage) error {
	if version == "" {
		return fmt.Errorf("version is required")
	}
	parsed, err := parseAgentImageFromMetadata(metadata)
	if err != nil {
		return err
	}
	if parsed.Tag != version {
		return AgentMetadataFieldError{
			Field:   "metadata.image",
			Message: fmt.Sprintf("tag must equal version %q", version),
		}
	}
	return validateAgentMetadataDigestIfPresent(metadata)
}

func parseAgentImageFromMetadata(metadata json.RawMessage) (AgentImageParsed, error) {
	if len(metadata) == 0 {
		return AgentImageParsed{}, AgentMetadataFieldError{Field: "metadata.image", Message: "required"}
	}
	var meta map[string]any
	if err := json.Unmarshal(metadata, &meta); err != nil {
		return AgentImageParsed{}, fmt.Errorf("%w: %v", ErrInvalidAgentMetadata, err)
	}
	image, _ := meta["image"].(string)
	return ParseAgentImageRef(image)
}

func validateAgentMetadataDigestIfPresent(metadata json.RawMessage) error {
	var meta map[string]any
	if err := json.Unmarshal(metadata, &meta); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAgentMetadata, err)
	}
	digest, _ := meta["digest"].(string)
	digest = strings.TrimSpace(digest)
	if digest == "" {
		return nil
	}
	if !agentDigestRE.MatchString(digest) {
		return AgentMetadataFieldError{
			Field:   "metadata.digest",
			Message: "must match sha256: followed by 64 lowercase hex digits",
		}
	}
	return nil
}
