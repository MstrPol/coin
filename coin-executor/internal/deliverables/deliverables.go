package deliverables

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultName = "app"

var P0Types = []string{"image", "liquibase-image", "artifact"}

type ArtifactSource struct {
	Path string `yaml:"path"`
	As   string `yaml:"as,omitempty"`
}

type Spec struct {
	Type      string           `yaml:"type"`
	Format    string           `yaml:"format,omitempty"`
	Context   string           `yaml:"context,omitempty"`
	Path      string           `yaml:"path,omitempty"`
	Source    string           `yaml:"source,omitempty"`
	DependsOn string           `yaml:"dependsOn,omitempty"`
	Sources   []ArtifactSource `yaml:"sources,omitempty"`
}

func Normalize(raw map[string]Spec) map[string]Spec {
	if len(raw) == 0 {
		return map[string]Spec{
			DefaultName: {Type: "image", Context: "."},
		}
	}
	out := make(map[string]Spec, len(raw))
	for name, spec := range raw {
		out[name] = spec
	}
	return out
}

func Validate(items map[string]Spec, allowedTypes []string) error {
	if len(items) == 0 {
		return fmt.Errorf("deliverables: at least one deliverable is required")
	}
	allowed := make(map[string]struct{}, len(allowedTypes))
	for _, t := range allowedTypes {
		allowed[t] = struct{}{}
	}
	if len(allowed) == 0 {
		for _, t := range P0Types {
			allowed[t] = struct{}{}
		}
	}

	names := make(map[string]struct{}, len(items))
	for name, spec := range items {
		if name == "" {
			return fmt.Errorf("deliverables: name is required")
		}
		if _, dup := names[name]; dup {
			return fmt.Errorf("deliverables: duplicate name %q", name)
		}
		names[name] = struct{}{}

		typ := strings.TrimSpace(spec.Type)
		if typ == "" {
			return fmt.Errorf("deliverables.%s: type is required", name)
		}
		if _, ok := allowed[typ]; !ok {
			return fmt.Errorf("deliverables.%s: type %q not supported by gp-content", name, typ)
		}

		switch typ {
		case "image":
			if err := validatePathOrContext(name, spec); err != nil {
				return err
			}
		case "liquibase-image":
			if strings.TrimSpace(spec.Source) == "" && strings.TrimSpace(spec.Path) == "" {
				return fmt.Errorf("deliverables.%s: source or path is required for liquibase-image", name)
			}
		case "artifact":
			if spec.Format != "" && spec.Format != "zip" {
				return fmt.Errorf("deliverables.%s: only format zip is supported in P0", name)
			}
			if len(spec.Sources) == 0 && strings.TrimSpace(spec.Source) == "" {
				return fmt.Errorf("deliverables.%s: source or sources is required for artifact", name)
			}
			for i, src := range spec.Sources {
				if strings.TrimSpace(src.Path) == "" {
					return fmt.Errorf("deliverables.%s.sources[%d]: path is required", name, i)
				}
				if strings.Contains(src.Path, "*") || strings.Contains(src.Path, "?") {
					return fmt.Errorf("deliverables.%s.sources[%d]: glob patterns are not allowed", name, i)
				}
			}
		}

		if dep := strings.TrimSpace(spec.DependsOn); dep != "" {
			if _, ok := items[dep]; !ok {
				return fmt.Errorf("deliverables.%s: dependsOn %q not found", name, dep)
			}
		}
	}
	return nil
}

func validatePathOrContext(name string, spec Spec) error {
	ctx := strings.TrimSpace(spec.Context)
	p := strings.TrimSpace(spec.Path)
	if ctx == "" && p == "" {
		return nil
	}
	if ctx != "" && strings.Contains(ctx, "*") {
		return fmt.Errorf("deliverables.%s: glob in context is not allowed", name)
	}
	if p != "" && strings.Contains(p, "*") {
		return fmt.Errorf("deliverables.%s: glob in path is not allowed", name)
	}
	return nil
}

func WriteState(workspace string, items map[string]Spec) error {
	path := filepath.Join(workspace, ".coin", "deliverables.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("{\n")
	first := true
	for name, spec := range items {
		if !first {
			b.WriteString(",\n")
		}
		first = false
		b.WriteString(fmt.Sprintf("  %q: {\"type\":%q", name, spec.Type))
		if spec.Context != "" {
			b.WriteString(fmt.Sprintf(",\"context\":%q", spec.Context))
		}
		if spec.Path != "" {
			b.WriteString(fmt.Sprintf(",\"path\":%q", spec.Path))
		}
		if spec.Source != "" {
			b.WriteString(fmt.Sprintf(",\"source\":%q", spec.Source))
		}
		if spec.DependsOn != "" {
			b.WriteString(fmt.Sprintf(",\"dependsOn\":%q", spec.DependsOn))
		}
		b.WriteString("}")
	}
	b.WriteString("\n}\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
