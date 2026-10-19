package stats

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func configPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "typetrainer")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "config.json")
}

func statsPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "typetrainer")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "stats.json")
}

// LoadConfig loads configuration from disk
func LoadConfig() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return DefaultConfig()
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	return cfg
}

// SaveConfig saves configuration to disk
func (c *Config) Save() {
	data, _ := json.MarshalIndent(c, "", "  ")
	os.WriteFile(configPath(), data, 0644)
}

// Load loads stats from disk
func Load() *Stats {
	data, err := os.ReadFile(statsPath())
	if err != nil {
		return New()
	}

	var s Stats
	if err := json.Unmarshal(data, &s); err != nil {
		return New()
	}
	return &s
}

// Save saves stats to disk
func (s *Stats) Save() {
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(statsPath(), data, 0644)
}
