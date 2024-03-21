package ui

import "github.com/charmbracelet/lipgloss"

var notificationStyle = lipgloss.NewStyle().
	Align(lipgloss.Left).
	Padding(0, 1).
	Height(1)

var InfoNotificationStyle = notificationStyle.Copy().
	Background(lipgloss.Color("#0000ff"))

var ErrorNotificationStyle = notificationStyle.Copy().
	Background(lipgloss.Color("#ff0000"))
