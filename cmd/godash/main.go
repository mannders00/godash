package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"godash/internal/config"
	"godash/internal/tui"
)

func main() {
	cfgPath := config.DefaultConfigPath()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--config", "-c":
			if len(os.Args) > 2 {
				cfgPath = os.Args[2]
			}
		case "init":
			if err := createSampleConfig(cfgPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created sample config at %s\n", cfgPath)
			os.Exit(0)
		case "help", "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &config.Config{}
		} else {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	}
	_ = cfg

	m := tui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createSampleConfig(path string) error {
	cfg := &config.Config{
		Machines: []config.Machine{
			{Name: "prod-server", Host: "192.168.1.100", User: "ubuntu", Port: 22},
			{Name: "staging", Host: "192.168.1.101", User: "ubuntu", Port: 22},
		},
		Sites: []config.Site{
			{Name: "myapp", URL: "https://myapp.com", Interval: 60, ExpectCode: 200},
			{Name: "api", URL: "https://api.myapp.com/health", Interval: 30, ExpectCode: 200},
		},
		Webhooks: []config.Webhook{
			{Name: "deploy", URL: "https://hooks.example.com/deploy", Method: "POST"},
		},
		Scripts: []config.Script{
			{Name: "backup", Command: "./scripts/backup.sh"},
			{Name: "deploy-prod", Command: "ssh", Args: []string{"prod", "deploy.sh"}, Remote: true},
		},
	}
	return cfg.Save(path)
}

func printHelp() {
	fmt.Println(`godash - Indie Hacker Dashboard

Usage:
  godash [flags]
  godash init

Flags:
  -c, --config string   config file path (default ~/.config/godash/config.yaml)
  -h, --help            show this help

Commands:
  init                  create sample config file

Key Bindings:
  1-4     Switch between views
  j/k     Navigate items
  enter   Select/action
  r       Refresh
  q/ctrl+c  Quit`)
}
