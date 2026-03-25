package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Underline(true)

	StyleSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("4"))

	StyleNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14"))

	StyleBorder = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	StyleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	StyleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	StyleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	StyleMuted = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	StyleKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6"))

	StyleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))
)

type HeaderComponent struct {
	Title string
	Width int
}

func (h HeaderComponent) View() string {
	return StyleHeader.Render(h.Title)
}

func MakeBox(title, content string, width, height int) string {
	borderStyle := lipgloss.RoundedBorder()
	box := lipgloss.NewStyle().
		Border(borderStyle).
		BorderForeground(lipgloss.Color("8")).
		Width(width).
		Height(height)

	titleStyled := StyleTitle.Render(title)
	return box.Render(lipgloss.JoinVertical(lipgloss.Left,
		titleStyled,
		"",
		content,
	))
}
