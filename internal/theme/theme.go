package theme

import "github.com/charmbracelet/lipgloss"

// Tokyo Night color palette
var (
	Border    = lipgloss.Color("#7aa2f7")
	BorderSel = lipgloss.Color("#ff9e64")
	Title     = lipgloss.Color("#7dcfff")
	TitleSel  = lipgloss.Color("#ff9e64")
	Sub       = lipgloss.Color("#9ece6a")
	Text      = lipgloss.Color("#c0caf5")
	Dim       = lipgloss.Color("#565f89")
	BgSel     = lipgloss.Color("#2e3c64")

	DockerBlue    = lipgloss.Color("#2496ed")
	DockerBlueSel = lipgloss.Color("#50b4ff")
	DockerBgSel   = lipgloss.Color("#1e375a")

	AddGreen    = lipgloss.Color("#9ece6a")
	AddGreenSel = lipgloss.Color("#bee68c")
	AddBgSel    = lipgloss.Color("#28463c")

	Error   = lipgloss.Color("#f7768e")
	Warning = lipgloss.Color("#ff9e64")
	Success = lipgloss.Color("#9ece6a")
)

// Card styles
var (
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1).
			Width(36)

	CardSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
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
				Border(lipgloss.RoundedBorder()).
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
				Border(lipgloss.RoundedBorder()).
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

	CountStyle = lipgloss.NewStyle().
			Foreground(Dim)

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
