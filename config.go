package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultAPIURL = "https://prilog.ai/api"
	configDir     = ".prilog"
	configFile    = "config.json"
)

func configuredAPIURL(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if env := os.Getenv("PRILOG_API_URL"); env != "" {
		return env
	}
	global, err := loadGlobalConfig()
	if err == nil && global.APIURL != "" {
		return global.APIURL
	}
	return defaultAPIURL
}

func globalConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "prilog", configFile), nil
}

func loadGlobalConfig() (globalConfig, error) {
	path, err := globalConfigPath()
	if err != nil {
		return globalConfig{}, err
	}

	var cfg globalConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func saveGlobalConfig(cfg globalConfig) error {
	path, err := globalConfigPath()
	if err != nil {
		return err
	}
	if cfg.APIURL == "" {
		cfg.APIURL = defaultAPIURL
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadRepoConfig(root string) (repoConfig, error) {
	var cfg repoConfig
	data, err := os.ReadFile(filepath.Join(root, configDir, configFile))
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func saveRepoConfig(root string, cfg repoConfig) error {
	dir := filepath.Join(root, configDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, configFile), data, 0644)
}

func trimConfiguredURL(value string) string {
	return strings.TrimRight(value, "/")
}
