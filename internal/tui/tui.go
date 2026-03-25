package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type View int

const (
	ViewMain View = iota
	ViewMachines
	ViewSites
	ViewWebhooks
	ViewScripts
)

type Model struct {
	currentView View
	machines    MachinesModel
	sites       SitesModel
	webhooks    WebhooksModel
	scripts     ScriptsModel
	width       int
	height      int
	err         error
}

func NewModel() Model {
	return Model{
		currentView: ViewMain,
		machines:    NewMachinesModel(),
		sites:       NewSitesModel(),
		webhooks:    NewWebhooksModel(),
		scripts:     NewScriptsModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.machines.Init(),
		m.sites.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView == ViewMain {
				return m, tea.Quit
			}
			m.currentView = ViewMain
			return m, nil
		case "1":
			m.currentView = ViewMachines
		case "2":
			m.currentView = ViewSites
		case "3":
			m.currentView = ViewWebhooks
		case "4":
			m.currentView = ViewScripts
		case "esc":
			m.currentView = ViewMain
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.machines.width = msg.Width
		m.machines.height = msg.Height
		m.sites.width = msg.Width
		m.sites.height = msg.Height
		m.webhooks.width = msg.Width
		m.webhooks.height = msg.Height
		m.scripts.width = msg.Width
		m.scripts.height = msg.Height
	}

	switch m.currentView {
	case ViewMachines:
		newM, cmd := m.machines.Update(msg)
		m.machines = newM.(MachinesModel)
		cmds = append(cmds, cmd)
	case ViewSites:
		newS, cmd := m.sites.Update(msg)
		m.sites = newS.(SitesModel)
		cmds = append(cmds, cmd)
	case ViewWebhooks:
		newW, cmd := m.webhooks.Update(msg)
		m.webhooks = newW.(WebhooksModel)
		cmds = append(cmds, cmd)
	case ViewScripts:
		newSc, cmd := m.scripts.Update(msg)
		m.scripts = newSc.(ScriptsModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.currentView {
	case ViewMachines:
		return m.machines.View()
	case ViewSites:
		return m.sites.View()
	case ViewWebhooks:
		return m.webhooks.View()
	case ViewScripts:
		return m.scripts.View()
	default:
		return m.renderMainMenu()
	}
}

func (m Model) renderMainMenu() string {
	menu := []string{
		StyleTitle.Render("godash - Indie Hacker Dashboard"),
		"",
		StyleKey.Render("1") + " " + StyleNormal.Render("Machines"),
		StyleKey.Render("2") + " " + StyleNormal.Render("Sites"),
		StyleKey.Render("3") + " " + StyleNormal.Render("Webhooks"),
		StyleKey.Render("4") + " " + StyleNormal.Render("Scripts"),
		"",
		StyleMuted.Render("Press number to select, q to quit"),
	}
	return lipgloss.JoinVertical(lipgloss.Left, menu...)
}
