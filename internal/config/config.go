package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Version  int
	Project  struct{ ID, Name string }
	Context  struct{ Files []string }
	Commands map[string]string
}

func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "blackboard.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("blackboard.yaml not found")
		}
		dir = parent
	}
}

func Load(root string) (Config, error) {
	b, err := os.ReadFile(filepath.Join(root, "blackboard.yaml"))
	if err != nil {
		return Config{}, err
	}
	// Tiny YAML parser for this demo. Replace with yaml.v3 in real CLI.
	cfg := Config{Version: 1, Commands: map[string]string{}}
	lines := strings.Split(string(b), "\n")
	section := ""
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "project:" || line == "context:" || line == "commands:" {
			section = strings.TrimSuffix(line, ":")
			continue
		}
		if strings.HasPrefix(line, "version:") {
			fmt.Sscanf(line, "version: %d", &cfg.Version)
			continue
		}
		if section == "project" && strings.HasPrefix(line, "id:") {
			cfg.Project.ID = trimVal(line[3:])
		}
		if section == "project" && strings.HasPrefix(line, "name:") {
			cfg.Project.Name = trimVal(line[5:])
		}
		if section == "context" && strings.HasPrefix(line, "-") {
			cfg.Context.Files = append(cfg.Context.Files, trimVal(line[1:]))
		}
		if section == "commands" && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			cfg.Commands[strings.TrimSpace(parts[0])] = trimVal(parts[1])
		}
	}
	if cfg.Project.ID == "" {
		return Config{}, fmt.Errorf("project.id missing")
	}
	return cfg, nil
}

func trimVal(s string) string { return strings.Trim(strings.TrimSpace(s), "\"") }
