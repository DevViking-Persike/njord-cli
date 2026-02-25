package theme

import "github.com/charmbracelet/lipgloss"

// Orange color palette
var (
	Border    = lipgloss.Color("#e07020")
	BorderSel = lipgloss.Color("#ff9800")
	Title     = lipgloss.Color("#ff8c00")
	TitleSel  = lipgloss.Color("#ffb74d")
	Sub       = lipgloss.Color("#ffcc80")
	SubSel    = lipgloss.Color("#ffe0b2")
	Text      = lipgloss.Color("#ffdab9")
	Dim       = lipgloss.Color("#8b6508")
	DimSel    = lipgloss.Color("#daa520")
	BgSel     = lipgloss.Color("#4e2a00")

	DockerBlue    = lipgloss.Color("#e67e22")
	DockerBlueSel = lipgloss.Color("#f0a04b")
	DockerBgSel   = lipgloss.Color("#5c3300")

	AddGreen    = lipgloss.Color("#ffc107")
	AddGreenSel = lipgloss.Color("#ffd54f")
	AddBgSel    = lipgloss.Color("#4e3800")

	SettingsGray    = lipgloss.Color("#90a4ae")
	SettingsGraySel = lipgloss.Color("#b0bec5")
	SettingsBgSel   = lipgloss.Color("#37474f")

	Error   = lipgloss.Color("#ff6b6b")
	Warning = lipgloss.Color("#ffb347")
	Success = lipgloss.Color("#ffd700")
)

// Card styles
var (
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1).
			Width(36)

	CardSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(BorderSel).
				Padding(0, 1).
				Width(36).
				Background(BgSel)

	DockerCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DockerBlue).
			Padding(0, 1).
			Width(36)

	DockerCardSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(DockerBlueSel).
				Padding(0, 1).
				Width(36).
				Background(DockerBgSel)

	AddCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AddGreen).
			Padding(0, 1).
			Width(36)

	AddCardSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(AddGreenSel).
				Padding(0, 1).
				Width(36).
				Background(AddBgSel)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Title)

	TitleSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(TitleSel)

	SubStyle = lipgloss.NewStyle().
			Foreground(Sub)

	SubSelectedStyle = lipgloss.NewStyle().
				Foreground(SubSel).
				Bold(true)

	CountStyle = lipgloss.NewStyle().
			Foreground(Dim)

	CountSelectedStyle = lipgloss.NewStyle().
				Foreground(DimSel)

	TextStyle = lipgloss.NewStyle().
			Foreground(Text)

	DimStyle = lipgloss.NewStyle().
			Foreground(Dim)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	DockerTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(DockerBlue)

	DockerTitleSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(DockerBlueSel)

	AddTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AddGreen)

	AddTitleSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(AddGreenSel)

	SettingsCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(SettingsGray).
				Padding(0, 1).
				Width(36)

	SettingsCardSelectedStyle = lipgloss.NewStyle().
					Border(lipgloss.DoubleBorder()).
					BorderForeground(SettingsGraySel).
					Padding(0, 1).
					Width(36).
					Background(SettingsBgSel)

	SettingsTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(SettingsGray)

	SettingsTitleSelectedStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(SettingsGraySel)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Title).
			MarginBottom(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Dim)

	StatusRunning = lipgloss.NewStyle().
			Foreground(Success)

	StatusStopped = lipgloss.NewStyle().
			Foreground(Error)

	StatusPartial = lipgloss.NewStyle().
			Foreground(Warning)
)
