package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles holds styled Lipgloss components for an active theme.
type Styles struct {
	Theme Theme

	BorderLine  lipgloss.Style
	BorderBox   lipgloss.Style
	BorderTitle lipgloss.Style

	StatusBar   lipgloss.Style
	StatusTitle lipgloss.Style
	ProgressBar lipgloss.Style
	StatusDim   lipgloss.Style

	HelpBar   lipgloss.Style
	HelpAlert lipgloss.Style
	HelpKey   lipgloss.Style

	SelectedRow lipgloss.Style
	HeaderRow   lipgloss.Style
	ActiveSong  lipgloss.Style

	VizBase lipgloss.Style
	VizMid  lipgloss.Style
	VizHigh lipgloss.Style
	VizPeak lipgloss.Style

	ModalBox   lipgloss.Style
	ModalBtn   lipgloss.Style
	ModalBtnOn lipgloss.Style
}

// MakeStyles creates the complete Lipgloss style suite for a given Theme.
func MakeStyles(t Theme) Styles {
	lineStyle := lipgloss.NewStyle().Foreground(t.BorderColor)

	return Styles{
		Theme: t,

		BorderLine:  lineStyle,
		BorderBox:   lineStyle,
		BorderTitle: lipgloss.NewStyle().Bold(true).Foreground(t.BorderColor),

		StatusBar:   lineStyle,
		StatusTitle: lipgloss.NewStyle().Bold(true).Foreground(t.BorderColor),
		ProgressBar: lipgloss.NewStyle().Foreground(t.ProgressColor),
		StatusDim:   lipgloss.NewStyle().Faint(true).Foreground(t.BorderColor),

		HelpBar:   lineStyle,
		HelpAlert: lipgloss.NewStyle().Bold(true).Foreground(t.AlertColor),
		HelpKey:   lipgloss.NewStyle().Foreground(t.AlertColor),

		SelectedRow: lipgloss.NewStyle().Bold(true).Foreground(t.SelectedFgColor).Background(t.SelectedBgColor),
		HeaderRow:   lipgloss.NewStyle().Bold(true).Underline(true),
		ActiveSong:  lipgloss.NewStyle().Bold(true).Foreground(t.ProgressColor),

		VizBase: lipgloss.NewStyle().Bold(true).Foreground(t.VizBaseColor),
		VizMid:  lipgloss.NewStyle().Bold(true).Foreground(t.VizMidColor),
		VizHigh: lipgloss.NewStyle().Bold(true).Foreground(t.VizHighColor),
		VizPeak: lipgloss.NewStyle().Bold(true).Foreground(t.VizPeakColor),

		ModalBox:   lineStyle,
		ModalBtn:   lipgloss.NewStyle().Foreground(t.BorderColor),
		ModalBtnOn: lipgloss.NewStyle().Bold(true).Reverse(true).Foreground(t.AlertColor),
	}
}
