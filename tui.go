package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen uint8

const (
	screenLoading screen = iota
	screenGameSelection
	screenPreview
	screenConfirm
	screenApplying
	screenResult
	screenError
)

type model struct {
	screen screen

	root          string
	installations []GameInstallation
	cursor        int

	selectedGame GameInstallation
	plan         ChangePlan
	backups      []string
	err          error

	width  int
	height int
}

type installationsLoadedMsg struct {
	root          string
	installations []GameInstallation
	err           error
}

type planBuiltMsg struct {
	plan ChangePlan
	err  error
}

type appliedMsg struct {
	backups []string
	err     error
}

func initialModel() model {
	return model{screen: screenLoading}
}

func (m model) Init() tea.Cmd {
	return loadInstallationsCmd
}

func loadInstallationsCmd() tea.Msg {
	root, err := findCODRoot()
	if err != nil {
		return installationsLoadedMsg{err: err}
	}

	installations, err := discoverInstallations(root)
	return installationsLoadedMsg{
		root:          root,
		installations: installations,
		err:           err,
	}
}

func buildPlanCmd(game GameInstallation) tea.Cmd {
	return func() tea.Msg {
		plan, err := buildPlan(game)
		return planBuiltMsg{plan: plan, err: err}
	}
}

func applyPlanCmd(plan ChangePlan) tea.Cmd {
	return func() tea.Msg {
		backups, err := applyPlan(plan)
		return appliedMsg{backups: backups, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case installationsLoadedMsg:
		if msg.err != nil {
			m.screen = screenError
			m.err = msg.err
			return m, nil
		}

		m.root = msg.root
		m.installations = msg.installations
		if len(m.installations) == 0 {
			m.screen = screenError
			m.err = fmt.Errorf(
				"aucun dossier players ou playersBeta contenant des fichiers .txt/.txt0/.txt1 n'a été trouvé dans %q",
				m.root,
			)
			return m, nil
		}

		m.screen = screenGameSelection
		return m, nil

	case planBuiltMsg:
		if msg.err != nil {
			m.screen = screenError
			m.err = msg.err
			return m, nil
		}

		m.plan = msg.plan
		m.screen = screenPreview
		return m, nil

	case appliedMsg:
		if msg.err != nil {
			m.screen = screenError
			m.err = msg.err
			return m, nil
		}

		m.backups = msg.backups
		m.screen = screenResult
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.screen {
	case screenGameSelection:
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.installations)-1 {
				m.cursor++
			}

		case "enter":
			m.selectedGame = m.installations[m.cursor]
			return m, buildPlanCmd(m.selectedGame)
		}

	case screenPreview:
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "b":
			m.screen = screenGameSelection

		case "enter", "y":
			if !m.plan.HasChanges() {
				m.screen = screenResult
				return m, nil
			}

			m.screen = screenConfirm
		}

	case screenConfirm:
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "n", "esc", "b":
			m.screen = screenPreview

		case "y", "enter":
			m.screen = screenApplying
			return m, applyPlanCmd(m.plan)
		}

	case screenResult, screenError:
		switch key {
		case "enter", "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case screenLoading:
		return m.viewLoading()

	case screenGameSelection:
		return m.viewGameSelection()

	case screenPreview:
		return m.viewPreview()

	case screenConfirm:
		return m.viewConfirmation()

	case screenApplying:
		return m.viewApplying()

	case screenResult:
		return m.viewResult()

	case screenError:
		return m.viewError()

	default:
		return ""
	}
}

func (m model) viewLoading() string {
	content := renderStatus(
		"●",
		"Détection des configurations Call of Duty…",
		selectedStyle,
	)

	return "\n" + renderPanel(activePanelStyle, "Initialisation", content) + "\n"
}

func (m model) viewGameSelection() string {
	var items []string
	pathWidth := max(24, m.width-12)

	for index, game := range m.installations {
		prefix := "  "
		nameStyle := valueStyle

		if index == m.cursor {
			prefix = "▶ "
			nameStyle = selectedStyle
		}

		item := strings.Builder{}
		item.WriteString(prefix)
		item.WriteString(nameStyle.Render(game.Name))
		item.WriteString(" ")
		item.WriteString(dimStyle.Render("[" + game.Variant + "]"))
		item.WriteString("\n    ")
		item.WriteString(dimStyle.Render(truncateMiddle(game.PlayersDir, pathWidth)))
		item.WriteString("\n    ")
		item.WriteString(
			dimStyle.Render(
				strconv.Itoa(len(game.Files)) + " fichier(s) de configuration détecté(s)",
			),
		)

		if index == m.cursor {
			items = append(items, activePanelStyle.Render(item.String()))
			continue
		}

		items = append(items, panelStyle.Render(item.String()))
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		labelStyle.Render("Racine détectée"),
		valueStyle.Render(truncateMiddle(m.root, max(24, m.width-8))),
		"",
		strings.Join(items, "\n"),
	)

	return "\n" +
		renderHeader(m.width, screenGameSelection) +
		"\n\n" +
		renderPanel(activePanelStyle, "Sélectionnez une installation", content) +
		"\n" +
		renderFooter("↑/↓ ou j/k : naviguer • Entrée : analyser • q : quitter") +
		"\n"
}

func (m model) viewPreview() string {
	pathWidth := max(24, m.width-16)

	summary := lipgloss.JoinVertical(
		lipgloss.Left,
		labelStyle.Render("Jeu"),
		valueStyle.Render(m.plan.Game.Name+" ["+m.plan.Game.Variant+"]"),
		"",
		labelStyle.Render("Dossier"),
		valueStyle.Render(truncateMiddle(m.plan.Game.PlayersDir, pathWidth)),
	)

	var body string
	if !m.plan.HasChanges() {
		body = lipgloss.JoinVertical(
			lipgloss.Left,
			renderStatus(
				"✓",
				"Aucune modification nécessaire : tous les settings trouvés sont déjà conformes.",
				okStyle,
			),
			"",
			dimStyle.Render("Aucun fichier ne sera modifié."),
		)

		return "\n" +
			renderHeader(m.width, screenPreview) +
			"\n\n" +
			renderPanel(panelStyle, "Configuration analysée", summary) +
			"\n\n" +
			renderPanel(successPanelStyle, "Résultat", body) +
			"\n" +
			renderFooter("b/Échap : retour • Entrée : terminer • q : quitter") +
			"\n"
	}

	filePanels := make([]string, 0, len(m.plan.Files))
	for _, file := range m.plan.Files {
		rows := make([]string, 0, len(file.Changes)+1)
		rows = append(rows, dimStyle.Render(truncateMiddle(file.Path, pathWidth)))

		for _, change := range file.Changes {
			line := dimStyle.Copy().Width(6).Render("L" + strconv.Itoa(change.Line))
			key := keyStyle.Copy().Width(28).Render(change.Key)
			oldValue := dimStyle.Render(fmt.Sprintf("%q", change.OldValue))
			newValue := okStyle.Render(fmt.Sprintf("%q", change.NewValue))

			rows = append(
				rows,
				lipgloss.JoinHorizontal(
					lipgloss.Top,
					line,
					key,
					oldValue,
					dimStyle.Render(" → "),
					newValue,
				),
			)
		}

		filePanels = append(
			filePanels,
			renderPanel(
				panelStyle,
				"Fichier : "+filepath.Base(file.Path),
				lipgloss.JoinVertical(lipgloss.Left, rows...),
			),
		)
	}

	changeSummary := warnStyle.Render(
		fmt.Sprintf(
			"%d fichier(s), %d setting(s) seront modifiés.",
			m.plan.ChangedFileCount(),
			m.plan.ChangedSettingCount(),
		),
	)

	return "\n" +
		renderHeader(m.width, screenPreview) +
		"\n\n" +
		renderPanel(panelStyle, "Configuration analysée", summary) +
		"\n\n" +
		renderPanel(warningPanelStyle, "Modifications prévues", changeSummary) +
		"\n\n" +
		lipgloss.JoinVertical(lipgloss.Left, filePanels...) +
		"\n" +
		renderFooter("Aucun fichier n'est encore modifié • Entrée/y : poursuivre • b/Échap : retour • q : quitter") +
		"\n"
}

func (m model) viewConfirmation() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		labelStyle.Render("Jeu"),
		valueStyle.Render(m.plan.Game.Name+" ["+m.plan.Game.Variant+"]"),
		"",
		labelStyle.Render("Dossier"),
		valueStyle.Render(truncateMiddle(m.plan.Game.PlayersDir, max(24, m.width-8))),
		"",
		labelStyle.Render("Fichiers à modifier"),
		valueStyle.Render(strconv.Itoa(m.plan.ChangedFileCount())),
		"",
		labelStyle.Render("Settings à modifier"),
		valueStyle.Render(strconv.Itoa(m.plan.ChangedSettingCount())),
		"",
		renderStatus(
			"!",
			"Une sauvegarde datée sera créée pour chaque fichier avant écriture.",
			warnStyle,
		),
	)

	actions := lipgloss.JoinHorizontal(
		lipgloss.Top,
		okStyle.Render("[y / Entrée] Appliquer"),
		"  ",
		dimStyle.Render("[n / Échap] Annuler"),
	)

	return "\n" +
		renderHeader(m.width, screenConfirm) +
		"\n\n" +
		renderPanel(warningPanelStyle, "Confirmation requise", content) +
		"\n\n  " +
		actions +
		"\n" +
		renderFooter("Aucune écriture n'a encore été effectuée.") +
		"\n"
}

func (m model) viewApplying() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		renderStatus("●", "Création des backups…", selectedStyle),
		renderStatus("●", "Application atomique des modifications…", selectedStyle),
		"",
		dimStyle.Render("Veuillez patienter, ne fermez pas cette fenêtre."),
	)

	return "\n" +
		renderHeader(m.width, screenApplying) +
		"\n\n" +
		renderPanel(activePanelStyle, "Application en cours", content) +
		"\n"
}

func (m model) viewResult() string {
	var content strings.Builder

	content.WriteString(
		renderStatus(
			"✓",
			"Modifications appliquées avec succès.",
			okStyle,
		),
	)
	content.WriteString("\n\n")

	if len(m.backups) == 0 {
		content.WriteString(dimStyle.Render("Aucune écriture n'était nécessaire."))
	} else {
		content.WriteString(labelStyle.Render("Backups créés"))
		content.WriteString("\n")

		for _, backup := range m.backups {
			content.WriteString("  ")
			content.WriteString(dimStyle.Render(truncateMiddle(backup, max(24, m.width-10))))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(
		dimStyle.Render(
			"Vous pouvez démarrer le jeu et vérifier le comportement en partie.",
		),
	)

	return "\n" +
		renderHeader(m.width, screenResult) +
		"\n\n" +
		renderPanel(successPanelStyle, "Terminé", content.String()) +
		"\n" +
		renderFooter("Entrée, Échap ou q : quitter") +
		"\n"
}

func (m model) viewError() string {
	message := "Une erreur inconnue est survenue."
	if m.err != nil {
		message = m.err.Error()
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		renderStatus("✗", message, errStyle),
		"",
		dimStyle.Render("Aucun changement supplémentaire ne sera appliqué."),
	)

	return "\n" +
		renderPanel(errorPanelStyle, "Erreur", content) +
		"\n" +
		renderFooter("Entrée, Échap ou q : quitter") +
		"\n"
}
