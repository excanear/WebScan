package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// TUIConfig holds simple persisted UI preferences.
type TUIConfig struct {
	Style string `json:"style,omitempty"`
	Sort  string `json:"sort,omitempty"`
}

const configFileName = ".webscan_config.json"

// configPath returns the path to the config file in the current working directory.
func configPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, configFileName), nil
}

// LoadConfig loads the persisted TUI config if present; otherwise returns zero-value.
func LoadConfig() (TUIConfig, error) {
	var cfg TUIConfig
	p, err := configPath()
	if err != nil {
		return cfg, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// SaveConfig writes the given TUI config to disk.
func SaveConfig(cfg TUIConfig) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0644)
}
