package tui

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/spf13/viper"
)

// Keep the primary Cerebras brand color available within the package.
var (
    primaryColor = lipgloss.Color("rgb(241, 90, 41)") // Cerebras Orange
)

// Palette defines the color set used across the TUI.
type Palette struct {
    Primary lipgloss.Color
    Bg      lipgloss.Color
    Surface lipgloss.Color
    Muted   lipgloss.Color
    Text    lipgloss.Color
    Subtle  lipgloss.Color
    Success lipgloss.Color
    Warning lipgloss.Color
    Error   lipgloss.Color
}

// Styles is a collection of reusable lipgloss styles for the TUI.
type Styles struct {
    Palette      Palette
    Header       lipgloss.Style
    TabActive    lipgloss.Style
    TabInactive  lipgloss.Style
    StatusBar    lipgloss.Style
    Content      lipgloss.Style
    SectionTitle lipgloss.Style
    TableHeader  lipgloss.Style
    TableCell    lipgloss.Style
    Error        lipgloss.Style
    Info         lipgloss.Style
    ListItem     lipgloss.Style
    ListSelected lipgloss.Style
    ListCursor   lipgloss.Style
    ProgressEmpty lipgloss.Style
    Key          lipgloss.Style
    Hint         lipgloss.Style
}

// GetStyles builds and returns a Styles struct based on the configured theme.
// Supported themes: "dark" (default) and "light". Configure via viper key: theme.
func GetStyles() Styles {
    theme := viper.GetString("theme")
    if theme == "" {
        theme = "dark"
    }

    var pal Palette
    if theme == "light" {
        pal = Palette{
            Primary: primaryColor,
            Bg:      lipgloss.Color("#f7f7f7"),
            Surface: lipgloss.Color("#ffffff"),
            Muted:   lipgloss.Color("#d0d0d0"),
            Text:    lipgloss.Color("#111111"),
            Subtle:  lipgloss.Color("#555555"),
            Success: lipgloss.Color("#2ecc71"),
            Warning: lipgloss.Color("#f39c12"),
            Error:   lipgloss.Color("#e74c3c"),
        }
    } else { // dark
        pal = Palette{
            Primary: primaryColor,
            Bg:      lipgloss.Color("#0b0b0b"),
            Surface: lipgloss.Color("#101010"),
            Muted:   lipgloss.Color("#3a3a3a"),
            Text:    lipgloss.Color("#eaeaea"),
            Subtle:  lipgloss.Color("#a0a0a0"),
            Success: lipgloss.Color("#27ae60"),
            Warning: lipgloss.Color("#f39c12"),
            Error:   lipgloss.Color("#ff4d4d"),
        }
    }

    return Styles{
        Palette: pal,
        Header: lipgloss.NewStyle().
            Foreground(pal.Text).
            Background(pal.Surface).
            Bold(true).
            Padding(0, 1).
            Height(1),
        TabInactive: lipgloss.NewStyle().
            Foreground(pal.Subtle).
            Background(pal.Bg).
            Padding(0, 2).
            MarginRight(1),
        TabActive: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#ffffff")).
            Background(pal.Primary).
            Bold(true).
            Padding(0, 2).
            MarginRight(1),
        StatusBar: lipgloss.NewStyle().
            Foreground(pal.Text).
            Background(pal.Surface).
            Padding(0, 1),
        Content: lipgloss.NewStyle().
            Background(pal.Bg).
            Padding(1, 2),
        SectionTitle: lipgloss.NewStyle().
            Bold(true).
            Foreground(pal.Primary),
        TableHeader: lipgloss.NewStyle().
            Bold(true).
            Foreground(pal.Text).
            PaddingRight(2),
        TableCell: lipgloss.NewStyle().
            Foreground(pal.Subtle).
            PaddingRight(2),
        Error: lipgloss.NewStyle().Foreground(pal.Error),
        Info:  lipgloss.NewStyle().Foreground(pal.Subtle),
        ListItem:     lipgloss.NewStyle().Foreground(pal.Text),
        ListSelected: lipgloss.NewStyle().Foreground(pal.Text).Background(pal.Primary).Bold(true),
        ListCursor:   lipgloss.NewStyle().Foreground(pal.Primary).Bold(true),
        ProgressEmpty: lipgloss.NewStyle().Foreground(pal.Muted),
        Key:  lipgloss.NewStyle().Foreground(pal.Text).Bold(true),
        Hint: lipgloss.NewStyle().Foreground(pal.Subtle),
    }
}
