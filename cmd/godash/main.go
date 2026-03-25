package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"godash/internal/config"
	"godash/internal/tui"
)

//go:embed sample_config.yaml
var sampleConfig string

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
			fmt.Printf("Created config at %s\n", cfgPath)
			os.Exit(0)
		case "config":
			if len(os.Args) > 2 && os.Args[2] == "edit" {
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "vim"
				}
				cmd := exec.Command(editor, cfgPath)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
				os.Exit(0)
			}
			fmt.Println(cfgPath)
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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(sampleConfig), 0644)
}

func printHelp() {
	fmt.Println(`godash - Indie Hacker Dashboard

A terminal dashboard for developers managing servers, sites, 
webhooks, and scripts. Think of it as a k9s for indie hackers.

Usage:
  godash                  Start the dashboard
  godash init             Create config file
  godash config           Show config file path
  godash config edit      Edit config in $EDITOR

Flags:
  -c, --config string     Use specific config file

Config File:
  ~/.config/godash/config.yaml

  Run 'godash init' to create a sample config with examples
  for machines, sites, webhooks, and scripts.

Key Bindings:
  Tab/Shift+Tab    Switch panels
  j/k or arrows     Navigate items
  Enter            Drill into item details
  Esc/q            Go back / quit
  r                Refresh / Run
  s                SSH into machine
  t                Trigger webhook

Examples:
  godash                    # Start dashboard
  godash init               # Create ~/.config/godash/config.yaml
  godash config edit        # Edit config in your editor
  godash -c ./custom.yaml   # Use custom config`)
}
