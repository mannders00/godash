package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Machine struct {
	Name    string            `yaml:"name"`
	Host    string            `yaml:"host"`
	User    string            `yaml:"user"`
	Port    int               `yaml:"port"`
	KeyPath string            `yaml:"key_path"`
	Labels  map[string]string `yaml:"labels"`
}

type Site struct {
	Name       string            `yaml:"name"`
	URL        string            `yaml:"url"`
	Interval   int               `yaml:"interval"`
	ExpectCode int               `yaml:"expect_code"`
	Labels     map[string]string `yaml:"labels"`
}

type Webhook struct {
	Name   string            `yaml:"name"`
	URL    string            `yaml:"url"`
	Method string            `yaml:"method"`
	Secret string            `yaml:"secret"`
	Labels map[string]string `yaml:"labels"`
}

type Script struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Remote  bool              `yaml:"remote"`
	Host    string            `yaml:"host"`
	Labels  map[string]string `yaml:"labels"`
}

type Config struct {
	Machines []Machine `yaml:"machines"`
	Sites    []Site    `yaml:"sites"`
	Webhooks []Webhook `yaml:"webhooks"`
	Scripts  []Script  `yaml:"scripts"`
}

func DefaultConfigPath() string {
	if cfg := os.Getenv("GODASH_CONFIG"); cfg != "" {
		return cfg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "godash", "config.yaml")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	for i := range cfg.Machines {
		if cfg.Machines[i].Port == 0 {
			cfg.Machines[i].Port = 22
		}
	}
	for i := range cfg.Sites {
		if cfg.Sites[i].Interval == 0 {
			cfg.Sites[i].Interval = 60
		}
		if cfg.Sites[i].ExpectCode == 0 {
			cfg.Sites[i].ExpectCode = 200
		}
	}
	for i := range cfg.Webhooks {
		if cfg.Webhooks[i].Method == "" {
			cfg.Webhooks[i].Method = "POST"
		}
	}
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
