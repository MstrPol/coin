package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Validate проверяет связность golden-paths, starters и agents.
func Validate() error {
	root, err := Root()
	if err != nil {
		return err
	}

	gpCatalogPath := filepath.Join(root, goldenPathsSubdir, "catalog.yaml")
	agentsCatalogPath := filepath.Join(root, agentsSubdir, "catalog.yaml")
	startersDir := filepath.Join(root, startersSubdir)

	gpCatalog, err := loadYAMLMap(gpCatalogPath)
	if err != nil {
		return fmt.Errorf("golden-paths catalog: %w", err)
	}
	agentsCatalog, err := loadAgentsCatalog(agentsCatalogPath)
	if err != nil {
		return err
	}

	var errs []string

	for name := range gpCatalog {
		gpDir := filepath.Join(root, goldenPathsSubdir, name)
		entries, err := os.ReadDir(gpDir)
		if err != nil {
			errs = append(errs, fmt.Sprintf("GP %q: %v", name, err))
			continue
		}
		foundProfile := false
		for _, e := range entries {
			if !e.IsDir() || !strings.HasPrefix(e.Name(), "v") {
				continue
			}
			foundProfile = true
			ver := e.Name()
			profilePath := filepath.Join(gpDir, ver, "profile.yaml")
			errs = append(errs, validateProfile(root, name, ver, profilePath, gpCatalog, agentsCatalog)...)
		}
		if !foundProfile {
			errs = append(errs, fmt.Sprintf("GP %q: нет каталогов vN с profile.yaml", name))
		}
	}

	starterNames, err := listStarters(startersDir)
	if err != nil {
		return err
	}
	for _, name := range starterNames {
		if _, ok := gpCatalog[name]; !ok {
			errs = append(errs, fmt.Sprintf("starter %q: нет записи в golden-paths/catalog.yaml", name))
		}
		cfgPath := filepath.Join(startersDir, name, ".coin", "config.yaml")
		cfg, err := loadYAMLMap(cfgPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("starter %q: %v", name, err))
			continue
		}
		coinSec, _ := cfg["coin"].(map[string]any)
		tpl, _ := coinSec["template"].(string)
		if tpl != name {
			errs = append(errs, fmt.Sprintf("starter %q: coin.template=%q", name, tpl))
		}
	}

	stacks, _ := agentsCatalog["stacks"].(map[string]any)
	for stack, rawRuntimes := range stacks {
		runtimes, ok := rawRuntimes.(map[string]any)
		if !ok {
			continue
		}
		for runtime, rawEntry := range runtimes {
			entry, ok := rawEntry.(map[string]any)
			if !ok {
				continue
			}
			df, _ := entry["dockerfile"].(string)
			if df == "" {
				errs = append(errs, fmt.Sprintf("agents: stacks.%s.%s — dockerfile пуст", stack, runtime))
				continue
			}
			if _, err := os.Stat(filepath.Join(root, agentsSubdir, df)); err != nil {
				errs = append(errs, fmt.Sprintf("agents: stacks.%s.%s — файл %q не найден", stack, runtime, df))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("platform validate failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

func validateProfile(root, name, version, profilePath string, gpCatalog, agentsCatalog map[string]any) []string {
	var errs []string
	label := fmt.Sprintf("GP %q/%s", name, version)

	profile, err := loadYAMLMap(profilePath)
	if err != nil {
		return []string{fmt.Sprintf("%s: profile: %v", label, err)}
	}

	agent, _ := profile["agent"].(map[string]any)
	stack, _ := agent["stack"].(string)
	if stack == "" {
		return []string{fmt.Sprintf("%s: agent.stack пуст", label)}
	}

	catalogEntry, _ := gpCatalog[name].(map[string]any)
	catalogStack, _ := catalogEntry["stack"].(string)
	if catalogStack != "" && catalogStack != stack {
		errs = append(errs, fmt.Sprintf("%s: catalog.stack=%q != profile.agent.stack=%q", label, catalogStack, stack))
	}

	runtime, err := runtimeFromProfile(stack, agent)
	if err != nil {
		return append(errs, fmt.Sprintf("%s: %v", label, err))
	}
	if !agentStackExists(agentsCatalog, stack, runtime) {
		errs = append(errs, fmt.Sprintf("%s: agents/catalog.yaml — нет stacks.%s.%q", label, stack, runtime))
	}

	rev, ok := intFromYAML(agent["rev"])
	if !ok || rev < 0 {
		errs = append(errs, fmt.Sprintf("%s: agent.rev обязателен (int >= 0)", label))
	} else if catalogRev, ok := catalogRev(agentsCatalog, stack, runtime); ok && rev > catalogRev {
		errs = append(errs, fmt.Sprintf("%s: agent.rev=%d выше catalog rev=%d (образ ещё не собран?)", label, rev, catalogRev))
	}

	coinCli, _ := profile["coinCli"].(map[string]any)
	cliVersion, _ := coinCli["version"].(string)
	if strings.TrimSpace(cliVersion) == "" {
		errs = append(errs, fmt.Sprintf("%s: coinCli.version обязателен", label))
	}

	return errs
}

func intFromYAML(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func catalogRev(catalog map[string]any, stack, runtime string) (int, bool) {
	stacks, _ := catalog["stacks"].(map[string]any)
	runtimes, ok := stacks[stack].(map[string]any)
	if !ok {
		return 0, false
	}
	entry, ok := runtimes[runtime].(map[string]any)
	if !ok {
		return 0, false
	}
	return intFromYAML(entry["rev"])
}

func loadAgentsCatalog(path string) (map[string]any, error) {
	data, err := loadYAMLMap(path)
	if err != nil {
		return nil, fmt.Errorf("agents catalog: %w", err)
	}
	if _, ok := data["stacks"]; !ok {
		return nil, fmt.Errorf("agents catalog: stacks: отсутствует")
	}
	return data, nil
}

func loadYAMLMap(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func runtimeFromProfile(stack string, agent map[string]any) (string, error) {
	runtimes, _ := agent["runtime"].(map[string]any)
	if runtimes == nil {
		return "", fmt.Errorf("agent.runtime пуст")
	}
	key := runtimeKey(stack)
	v, ok := runtimes[key]
	if !ok {
		return "", fmt.Errorf("agent.runtime.%s не задан", key)
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("agent.runtime.%s — пустое значение", key)
	}
	return s, nil
}

func runtimeKey(stack string) string {
	switch stack {
	case "python-uv", "python-pip":
		return "python"
	case "java-maven", "java-gradle":
		return "java"
	default:
		return stack
	}
}

func agentStackExists(catalog map[string]any, stack, runtime string) bool {
	stacks, _ := catalog["stacks"].(map[string]any)
	runtimes, ok := stacks[stack].(map[string]any)
	if !ok {
		return false
	}
	_, ok = runtimes[runtime]
	return ok
}

func listStarters(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("starters/: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, e.Name(), ".coin", "config.yaml")); err == nil {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
