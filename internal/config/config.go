package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	HTTPPort int    `json:"http_port"`
}

func Load(configPath string) (Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		exe, _ := os.Executable()
		configPath = filepath.Join(filepath.Dir(exe), "config.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
