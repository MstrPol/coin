package outputs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Entry struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
	Ref    string `json:"ref,omitempty"`
	Digest string `json:"digest,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

const stateFile = ".coin/outputs.json"

func Path(workspace string) string {
	return filepath.Join(workspace, stateFile)
}

func Load(workspace string) ([]Entry, error) {
	raw, err := os.ReadFile(Path(workspace))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var items []Entry
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse outputs: %w", err)
	}
	return items, nil
}

func Merge(workspace string, entries ...Entry) error {
	current, err := Load(workspace)
	if err != nil {
		return err
	}
	byName := make(map[string]Entry, len(current)+len(entries))
	for _, e := range current {
		byName[e.Name] = e
	}
	for _, e := range entries {
		if e.Name == "" {
			continue
		}
		byName[e.Name] = e
	}
	out := make([]Entry, 0, len(byName))
	for _, e := range byName {
		out = append(out, e)
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(Path(workspace)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(Path(workspace), raw, 0o644)
}
