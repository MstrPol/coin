package embed

import (
	"embed"
	"fmt"
	"path"
)

//go:embed scripts dockerfiles
var assets embed.FS

// Script возвращает содержимое стандартного скрипта для стека и стадии.
func Script(stack, stage string) (string, error) {
	data, err := assets.ReadFile(path.Join("scripts", stack, stage+".sh"))
	if err != nil {
		return "", fmt.Errorf("script %s/%s.sh not found: %w", stack, stage, err)
	}
	return string(data), nil
}

// DockerfileTemplate возвращает содержимое шаблона Dockerfile для стека.
func DockerfileTemplate(stack string) (string, error) {
	data, err := assets.ReadFile(path.Join("dockerfiles", stack, "Dockerfile"))
	if err != nil {
		return "", fmt.Errorf("dockerfile template %s/Dockerfile not found: %w", stack, err)
	}
	return string(data), nil
}
