package store

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidateContentRefOnWriteForAgentSkipsValidation(t *testing.T) {
	bad := json.RawMessage(`{"schemaVersion":2,"packageUrl":"not-a-url"}`)
	if err := validateContentRefOnWriteForType("agent", bad); err != nil {
		t.Fatalf("agent must skip content_ref validation: %v", err)
	}
}

func TestValidateContentRefOnWriteForGPContentStillValidates(t *testing.T) {
	bad := json.RawMessage(`{
		"schemaVersion": 2,
		"package": {"sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	}`)
	if err := validateContentRefOnWriteForType("gp-content", bad); err == nil {
		t.Fatal("expected validation error for gp-content content_ref")
	}
}

func TestDerivedExecutorPin(t *testing.T) {
	pin, ok := DerivedExecutorPin("coin-agent", "1.2.0")
	if !ok {
		t.Fatal("expected derived pin")
	}
	if pin.Type != "executor" || pin.Name != "coin-executor" || pin.Version != "1.2.0" {
		t.Fatalf("unexpected pin: %#v", pin)
	}
	pin, ok = DerivedExecutorPin("coin-agent-arm", "1.0.0")
	if !ok {
		t.Fatal("expected derived pin for alternate profile")
	}
	if pin.Name != "coin-executor" || pin.Version != "1.0.0" {
		t.Fatalf("unexpected pin: %#v", pin)
	}
}

func TestValidateAgentMetadataForPromote(t *testing.T) {
	validDigest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	validMeta := json.RawMessage(`{"image":"nexus:8082/coin-docker/coin-agent:1.2.0","digest":"` + validDigest + `"}`)

	if err := validateAgentMetadataForPromote("1.2.0", validMeta); err != nil {
		t.Fatalf("valid metadata: %v", err)
	}

	cases := []struct {
		name    string
		version string
		meta    string
		field   string
	}{
		{"missing image", "1.2.0", `{"digest":"` + validDigest + `"}`, "metadata.image"},
		{"missing digest", "1.2.0", `{"image":"nexus:8082/coin-docker/coin-agent:1.2.0"}`, "metadata.digest"},
		{"bad digest", "1.2.0", `{"image":"nexus:8082/coin-docker/coin-agent:1.2.0","digest":"sha256:deadbeef"}`, "metadata.digest"},
		{"tag mismatch", "1.2.0", `{"image":"nexus:8082/coin-docker/coin-agent:9.9.9","digest":"` + validDigest + `"}`, "metadata.image"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAgentMetadataForPromote(tc.version, json.RawMessage(tc.meta))
			if err == nil {
				t.Fatal("expected error")
			}
			var fieldErr AgentMetadataFieldError
			if !errors.As(err, &fieldErr) || fieldErr.Field != tc.field {
				t.Fatalf("expected field %q, got %v", tc.field, err)
			}
		})
	}
}

func TestNormalizeAgentMetadataStripsGoarch(t *testing.T) {
	in := json.RawMessage(`{"image":"img:1.0.0","digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","goarch":"amd64"}`)
	out := normalizeAgentMetadata(in)
	var meta map[string]any
	if err := json.Unmarshal(out, &meta); err != nil {
		t.Fatal(err)
	}
	if _, ok := meta["goarch"]; ok {
		t.Fatal("goarch must be stripped")
	}
	if meta["image"] != "img:1.0.0" {
		t.Fatalf("image preserved: %#v", meta)
	}
}