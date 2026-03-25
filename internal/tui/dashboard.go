package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type DashboardModel struct {
	focus    int
	width    int
	height   int
	machines []MachineItem
	sites    []SiteItem
	webhooks []WebhookItem
	scripts  []ScriptItem
}

func NewDashboardModel() DashboardModel {
	return DashboardModel{focus: 0}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.focus = (m.focus + 1) % 4
		case "shift+tab":
			m.focus = (m.focus - 1 + 4) % 4
		}
	}
	return m, nil
}

func (m DashboardModel) View() string {
	return ""
}

func (m *DashboardModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *DashboardModel) Focused() int {
	return m.focus
}

func (m *DashboardModel) SetMachines(items []MachineItem) {
	m.machines = items
}

func (m *DashboardModel) SetSites(items []SiteItem) {
	m.sites = items
}

func (m *DashboardModel) SetWebhooks(items []WebhookItem) {
	m.webhooks = items
}

func (m *DashboardModel) SetScripts(items []ScriptItem) {
	m.scripts = items
}
