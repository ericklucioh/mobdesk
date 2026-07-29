package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	colorBlack  = "#000000"
	colorPanel  = "#050505"
	colorPanel2 = "#0b0b0b"
	colorLine   = "#8e63c7"
	colorPurple = "#b58cff"
	colorLilac  = "#c3a9ff"
	colorGreen  = "#56d6a0"
	colorYellow = "#f2c14e"
	colorRed    = "#ef6f91"
	colorMuted  = "#8e83a8"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorLilac)).
			Background(lipgloss.Color(colorBlack)).
			Padding(1, 1).
			BorderBottom(true).
			BorderForeground(lipgloss.Color(colorLine))

	headerBrandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorLilac)).Bold(true)
	headerLinkStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorLilac)).Underline(true)
	headerCloseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorLilac)).Bold(true)
	tagStyle         = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorLilac)).
				Bold(true)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorLilac)).
			Bold(true).
			MarginTop(0).
			MarginBottom(0)

	bodyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f5f0ff"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))

	homeStatusLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#c0b8d2"))
	homeActiveStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(colorGreen)).Bold(true)
	homeInactiveStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorRed)).Bold(true)

	cardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f5f0ff")).
			Background(lipgloss.Color(colorBlack)).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorLine)).
			Padding(1, 1)

	cardSelectedStyle = cardStyle.
				BorderForeground(lipgloss.Color(colorLilac)).
				Bold(true)

	primaryCardStyle = cardStyle.
				BorderForeground(lipgloss.Color(colorPurple))

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorLilac)).
			Background(lipgloss.Color(colorBlack)).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorLine)).
			Padding(0, 1)

	primaryButtonStyle = buttonStyle.
				Foreground(lipgloss.Color(colorBlack)).
				Background(lipgloss.Color(colorLilac)).
				BorderForeground(lipgloss.Color(colorLilac)).
				Bold(true)

	stepStyle          = cardStyle.Padding(0, 1)
	stepDoneStyle      = stepStyle.Foreground(lipgloss.Color(colorGreen))
	stepActiveStyle    = stepStyle.Foreground(lipgloss.Color(colorLilac)).Bold(true)
	operationWaitStyle = bodyStyle.Bold(true).Foreground(lipgloss.Color(colorYellow))

	modalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f5f0ff")).
			Background(lipgloss.Color(colorPanel)).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(colorLilac)).
			Padding(1, 2).
			Align(lipgloss.Center)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			BorderTop(true).
			BorderForeground(lipgloss.Color(colorLine)).
			Padding(0, 1)
)

func statusColor(state string) lipgloss.Style {
	switch state {
	case "ok", "healthy", "installed", "running", "ativo", "ativado", "current", "updated":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorGreen))
	case "warning", "degraded", "installing", "starting", "stopped", "parada", "desativado", "available":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorYellow))
	case "error", "failed":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorRed))
	default:
		return mutedStyle
	}
}

func terminalWidth(width int) int {
	if width <= 0 {
		return 80
	}
	return width
}

func contentWidth(width int) int {
	width = terminalWidth(width)
	if width < 20 {
		return 20
	}
	if width > 76 {
		return 76
	}
	return width
}

func contentColumns(width int) int {
	if contentWidth(width) >= 64 {
		return 2
	}
	return 1
}

func wrapText(text string, width int) string {
	return lipgloss.NewStyle().Width(contentWidth(width)).Render(text)
}

func joinCards(cards []string, width int) string {
	if len(cards) == 0 {
		return ""
	}
	if contentColumns(width) == 1 {
		return strings.Join(cards, "\n\n")
	}
	rows := make([]string, 0, (len(cards)+1)/2)
	cardWidth := (contentWidth(width) - 2) / 2
	for index := 0; index < len(cards); index += 2 {
		left := lipgloss.NewStyle().Width(cardWidth).Render(cards[index])
		if index+1 >= len(cards) {
			rows = append(rows, left)
			continue
		}
		right := lipgloss.NewStyle().Width(cardWidth).Render(cards[index+1])
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right))
	}
	return strings.Join(rows, "\n\n")
}
