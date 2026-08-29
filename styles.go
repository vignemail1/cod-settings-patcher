package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorPrimary = lipgloss.AdaptiveColor{
		Light: "#1D4ED8",
		Dark:  "#60A5FA",
	}
	colorSuccess = lipgloss.AdaptiveColor{
		Light: "#047857",
		Dark:  "#34D399",
	}
	colorWarning = lipgloss.AdaptiveColor{
		Light: "#B45309",
		Dark:  "#FBBF24",
	}
	colorDanger = lipgloss.AdaptiveColor{
		Light: "#B91C1C",
		Dark:  "#F87171",
	}
	colorText = lipgloss.AdaptiveColor{
		Light: "#1F2937",
		Dark:  "#E5E7EB",
	}
	colorMuted = lipgloss.AdaptiveColor{
		Light: "#6B7280",
		Dark:  "#9CA3AF",
	}
	colorBorder = lipgloss.AdaptiveColor{
		Light: "#D1D5DB",
		Dark:  "#4B5563",
	}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	activePanelStyle = panelStyle.
				BorderForeground(colorPrimary)

	successPanelStyle = panelStyle.
				BorderForeground(colorSuccess)

	warningPanelStyle = panelStyle.
				BorderForeground(colorWarning)

	errorPanelStyle = panelStyle.
			BorderForeground(colorDanger)

	okStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorSuccess)

	warnStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWarning)

	errStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorDanger)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorText)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)
)

func renderPanel(style lipgloss.Style, title, content string) string {
	return style.Render(
		sectionTitleStyle.Render(title) +
			"\n\n" +
			content,
	)
}

func renderStatus(symbol, message string, style lipgloss.Style) string {
	return style.Render(symbol) + " " + valueStyle.Render(message)
}

func renderHeader(width int, active screen) string {
	const appName = "COD Settings Patcher"

	title := titleStyle.Render(appName)
	tagline := subtitleStyle.Render("Safe Call of Duty configuration optimizer")

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		tagline,
	)

	if width > 0 {
		header = panelStyle.Width(maxInt(30, width-4)).Render(header)
	} else {
		header = panelStyle.Render(header)
	}

	return header + "\n\n" + renderStepper(active)
}

func renderStepper(active screen) string {
	currentIndex := stepIndex(active)
	steps := []struct {
		label  string
		screen screen
	}{
		{label: "Detect", screen: screenGameSelection},
		{label: "Review", screen: screenPreview},
		{label: "Apply", screen: screenApplying},
		{label: "Complete", screen: screenResult},
	}

	parts := make([]string, 0, len(steps))
	for index, item := range steps {
		label := lipgloss.NewStyle().Foreground(colorMuted).Render(
			"○ " + string(rune('1'+index)) + " " + item.label,
		)

		switch {
		case index == currentIndex:
			label = selectedStyle.Render(
				"● " + string(rune('1'+index)) + " " + item.label,
			)
		case index < currentIndex:
			label = okStyle.Render(
				"✓ " + string(rune('1'+index)) + " " + item.label,
			)
		}

		parts = append(parts, label)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		strings.Join(parts, dimStyle.Render(" ── ")),
	)
}

func renderFooter(help string) string {
	return footerStyle.Render(help)
}

func truncateMiddle(value string, maxWidth int) string {
	if maxWidth <= 0 || lipgloss.Width(value) <= maxWidth {
		return value
	}

	runes := []rune(value)
	if maxWidth <= 3 {
		return string(runes[:maxWidth])
	}

	leftLength := (maxWidth - 1) / 2
	rightLength := maxWidth - leftLength - 1

	return string(runes[:leftLength]) + "…" + string(runes[len(runes)-rightLength:])
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func stepIndex(current screen) int {
	switch current {
	case screenGameSelection:
		return 0
	case screenPreview:
		return 1
	case screenConfirm, screenApplying:
		return 2
	case screenResult:
		return 3
	default:
		return 0
	}
}
