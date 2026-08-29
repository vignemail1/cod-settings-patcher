package main

import (
	"fmt"
	"path/filepath"
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

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

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
	return installationsLoadedMsg{root: root, installations: installations, err: err}
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
			m.err = fmt.Errorf("aucun dossier players ou playersBeta contenant des fichiers .txt/.txt0/.txt1 n'a été trouvé dans %q", m.root)
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
		case "y":
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
		return "\n  Détection des configurations Call of Duty…\n"
	case screenGameSelection:
		return m.viewGameSelection()
	case screenPreview:
		return m.viewPreview()
	case screenConfirm:
		return m.viewConfirmation()
	case screenApplying:
		return "\n  Création des backups et application atomique des modifications…\n"
	case screenResult:
		return m.viewResult()
	case screenError:
		return "\n  " + errStyle.Render("Erreur : "+m.err.Error()) + "\n\n  Entrée pour quitter.\n"
	default:
		return ""
	}
}

func (m model) viewGameSelection() string {
	var builder strings.Builder
	builder.WriteString("\n  " + titleStyle.Render("Call of Duty Settings Patcher") + "\n\n")
	builder.WriteString("  Racine détectée : " + dimStyle.Render(m.root) + "\n\n")
	builder.WriteString("  Sélectionnez une installation :\n\n")

	for i, game := range m.installations {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		_, _ = fmt.Fprintf(&builder, "  %s%s %s\n", cursor, game.Name, dimStyle.Render("["+game.Variant+"]"))
		builder.WriteString("      " + dimStyle.Render(game.PlayersDir) + "\n")
		_, _ = fmt.Fprintf(&builder, "      %d fichier(s) de configuration détecté(s)\n", len(game.Files))
	}

	builder.WriteString("\n  " + dimStyle.Render("↑/↓ ou j/k : naviguer • Entrée : analyser • q : quitter") + "\n")
	return builder.String()
}

func (m model) viewPreview() string {
	var builder strings.Builder
	builder.WriteString("\n  " + titleStyle.Render("Aperçu des modifications") + "\n\n")
	builder.WriteString("  Jeu : " + m.plan.Game.Name + " " + dimStyle.Render("["+m.plan.Game.Variant+"]") + "\n")
	builder.WriteString("  Dossier : " + dimStyle.Render(m.plan.Game.PlayersDir) + "\n\n")

	if !m.plan.HasChanges() {
		builder.WriteString("  " + okStyle.Render("Aucune modification nécessaire : tous les settings trouvés sont déjà conformes.") + "\n")
		builder.WriteString("\n  " + dimStyle.Render("b/Échap : retour • Entrée : terminer • q : quitter") + "\n")
		return builder.String()
	}

	_, _ = fmt.Fprintf(&builder, "  %s\n\n", warnStyle.Render(fmt.Sprintf("%d fichier(s), %d setting(s) seront modifiés.", m.plan.ChangedFileCount(), m.plan.ChangedSettingCount())))
	for _, file := range m.plan.Files {
		builder.WriteString("  Fichier : " + filepath.Base(file.Path) + "\n")
		builder.WriteString("    " + dimStyle.Render(file.Path) + "\n")
		for _, change := range file.Changes {
			_, _ = fmt.Fprintf(&builder, "    L%-4d %-32s %q → %q\n", change.Line, change.Key, change.OldValue, change.NewValue)
		}
		builder.WriteByte('\n')
	}

	builder.WriteString("  " + dimStyle.Render("Aucun fichier ne sera encore modifié. Entrée/y : poursuivre • b/Échap : retour • q : quitter") + "\n")
	return builder.String()
}

func (m model) viewConfirmation() string {
	return fmt.Sprintf("\n  %s\n\n  Jeu : %s\n  Dossier : %s\n  Fichiers : %d\n  Settings : %d\n\n  Une sauvegarde datée sera créée pour chaque fichier avant écriture.\n  Confirmer l'application ? %s\n",
		titleStyle.Render("Confirmation requise"),
		m.plan.Game.Name,
		dimStyle.Render(m.plan.Game.PlayersDir),
		m.plan.ChangedFileCount(),
		m.plan.ChangedSettingCount(),
		warnStyle.Render("y = appliquer • n/Échap = annuler"),
	)
}

func (m model) viewResult() string {
	var builder strings.Builder
	builder.WriteString("\n  " + okStyle.Render("Modifications appliquées avec succès.") + "\n\n")
	if len(m.backups) == 0 {
		builder.WriteString("  Aucune écriture n'était nécessaire.\n")
	} else {
		builder.WriteString("  Backups créés :\n")
		for _, backup := range m.backups {
			builder.WriteString("    " + dimStyle.Render(backup) + "\n")
		}
	}
	builder.WriteString("\n  " + dimStyle.Render("Entrée pour quitter.") + "\n")
	return builder.String()
}
