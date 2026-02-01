package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Frame provides a consistent app frame with top bar, content, and footer.
type Frame struct {
	width  int
	height int
}

// NewFrame creates a new frame with the given dimensions.
func NewFrame(width, height int) Frame {
	return Frame{width: width, height: height}
}

// Render renders the frame with the given components.
// Footer is pinned to the bottom of the screen.
func (f Frame) Render(topBar, content, footer string) string {
	if f.width == 0 || f.height == 0 {
		return content
	}

	// Calculate heights
	topBarHeight := lipgloss.Height(topBar)
	footerHeight := lipgloss.Height(footer)
	contentHeight := f.height - topBarHeight - footerHeight - 2 // -2 for newlines

	if contentHeight < 1 {
		contentHeight = 1
	}

	// Place content in available space (top-aligned)
	placedContent := lipgloss.Place(
		f.width,
		contentHeight,
		lipgloss.Left,
		lipgloss.Top,
		content,
	)

	var b strings.Builder
	b.WriteString(topBar)
	b.WriteString("\n")
	b.WriteString(placedContent)
	b.WriteString("\n")
	b.WriteString(footer)

	return b.String()
}

// TopBarStyle returns a style for the app top bar.
func TopBarStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 2).
		Bold(true)
}

// BreadcrumbStyle returns a style for breadcrumb navigation.
func BreadcrumbStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(MutedColor)
}

// FooterStyle returns a style for the footer.
func FooterStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 2).
		Foreground(MutedColor).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "250", Dark: "235"})
}

// RenderTopBar renders a top bar with title and breadcrumb.
func RenderTopBar(width int, title string, breadcrumb string) string {
	if width == 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	breadcrumbStyle := BreadcrumbStyle()

	titleText := titleStyle.Render(title)
	breadcrumbText := ""
	if breadcrumb != "" {
		breadcrumbText = " " + breadcrumbStyle.Render("›") + " " + breadcrumbStyle.Render(breadcrumb)
	}

	bar := TopBarStyle(width).Render(titleText + breadcrumbText)
	return bar
}

// RenderFooter renders a footer with help text and status.
func RenderFooter(width int, helpText string, status string) string {
	if width == 0 {
		return ""
	}

	leftSide := helpText
	rightSide := status

	// Calculate spacing
	contentWidth := width - 4 // Account for padding
	leftWidth := lipgloss.Width(leftSide)
	rightWidth := lipgloss.Width(rightSide)
	spacing := contentWidth - leftWidth - rightWidth

	if spacing < 0 {
		spacing = 0
	}

	content := leftSide + strings.Repeat(" ", spacing) + rightSide

	return FooterStyle(width).Render(content)
}

// LoadingSpinner returns a simple loading indicator.
// Note: For animated spinners, use bubbles/spinner component with Tick messages.
// This static version is used for simple loading states where animation isn't needed.
func LoadingSpinner() string {
	return "⠋"
}

// StatusIndicator renders a status message with icon.
func StatusIndicator(icon, message string, color lipgloss.AdaptiveColor) string {
	iconStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(color)

	return iconStyle.Render(icon) + " " + textStyle.Render(message)
}

// EmptyState renders an empty state message.
func EmptyState(title, message, action string) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(MutedColor).
		MarginBottom(1)

	messageStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		MarginBottom(2)

	actionStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	var b strings.Builder
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(messageStyle.Render(message))
	b.WriteString("\n")
	if action != "" {
		b.WriteString(actionStyle.Render(action))
	}

	return lipgloss.NewStyle().
		Padding(2, 0).
		Render(b.String())
}

// InfoPanel renders a side information panel.
func InfoPanel(title string, items []string) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "250"})

	var b strings.Builder
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	for _, item := range items {
		b.WriteString(itemStyle.Render(item))
		b.WriteString("\n")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "250", Dark: "235"}).
		Padding(0, 2).
		Render(b.String())
}

// ConfirmationMessage renders a temporary confirmation message.
func ConfirmationMessage(message string) string {
	return lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true).
		Render("✓ " + message)
}

// FormatHelpKeys formats a list of help key bindings.
func FormatHelpKeys(bindings map[string]string) string {
	var parts []string
	// Order: most important first, comprehensive list for consistent ordering
	order := []string{"enter", "space", "a", "d", "y", "n", "↑↓", "j/k", "esc", "q"}

	for _, key := range order {
		if desc, ok := bindings[key]; ok {
			keyStyle := lipgloss.NewStyle().Foreground(MutedColor)
			descStyle := lipgloss.NewStyle().Foreground(MutedColor)
			parts = append(parts, keyStyle.Render(key)+" "+descStyle.Render(desc))
		}
	}

	// Collect remaining keys not in order
	var remaining []string
	for key := range bindings {
		found := false
		for _, k := range order {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			remaining = append(remaining, key)
		}
	}

	// Sort remaining keys for deterministic output
	sort.Strings(remaining)

	// Add sorted remaining keys
	for _, key := range remaining {
		desc := bindings[key]
		keyStyle := lipgloss.NewStyle().Foreground(MutedColor)
		descStyle := lipgloss.NewStyle().Foreground(MutedColor)
		parts = append(parts, keyStyle.Render(key)+" "+descStyle.Render(desc))
	}

	return strings.Join(parts, "  •  ")
}

// CenterVertical centers content vertically within available height.
func CenterVertical(content string, height int) string {
	lines := strings.Split(content, "\n")
	contentHeight := len(lines)

	if contentHeight >= height {
		return content
	}

	topPadding := (height - contentHeight) / 2
	var b strings.Builder

	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}
	b.WriteString(content)

	return b.String()
}

// TruncateWithEllipsis truncates a string to maxLen display width with ellipsis if needed.
// Uses lipgloss.Width for proper Unicode/emoji handling.
func TruncateWithEllipsis(s string, maxLen int) string {
	if lipgloss.Width(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		// Just truncate by runes when too short for ellipsis
		runes := []rune(s)
		if len(runes) <= maxLen {
			return s
		}
		return string(runes[:maxLen])
	}
	// Truncate to fit ellipsis within maxLen
	targetWidth := maxLen - 3 // Reserve space for "..."
	result := ""
	currentWidth := 0
	for _, r := range s {
		runeWidth := lipgloss.Width(string(r))
		if currentWidth+runeWidth > targetWidth {
			break
		}
		result += string(r)
		currentWidth += runeWidth
	}
	return result + "..."
}

// PadRight pads a string to width with spaces on the right.
func PadRight(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

// FormatCount formats a count ratio (e.g., "3 of 12").
func FormatCount(current, total int) string {
	return fmt.Sprintf("%d of %d", current, total)
}

// RenderStatusScreen renders a centered status screen with icon, message, and optional help keys.
// This is a common pattern for loading, error, and success screens.
func RenderStatusScreen(width, height int, title, breadcrumb, icon, message string, color lipgloss.AdaptiveColor, helpText, status string) string {
	topBar := RenderTopBar(width, title, breadcrumb)

	content := CenterVertical(
		StatusIndicator(icon, message, color),
		height-4,
	)

	footer := RenderFooter(width, helpText, status)

	frame := NewFrame(width, height)
	return frame.Render(topBar, content, footer)
}
