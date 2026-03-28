package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrf/dcc/internal/data"
)

// RenderPrsPanel renders the pull requests panel
func RenderPrsPanel(panel data.PrsPanel, width, height int, selected, loading bool, cursorIdx int) string {
	style := GetPanelStyle(selected, loading, ColorMagenta).
		Width(width).
		Height(height)

	var content strings.Builder

	if loading || panel.IsLoading {
		content.WriteString(TitleStyle.Render("PULL REQUESTS") + "\n\n")
		content.WriteString(ItalicStyle.Render("Fetching PRs..."))
		return style.Render(content.String())
	}

	// Build title with counts
	title := "PULL REQUESTS"
	var countParts []string
	if len(panel.NeedsReview) > 0 {
		countParts = append(countParts, fmt.Sprintf("%d review", len(panel.NeedsReview)))
	}
	if len(panel.YourPrs) > 0 {
		countParts = append(countParts, fmt.Sprintf("%d yours", len(panel.YourPrs)))
	}
	if len(countParts) > 0 {
		title = fmt.Sprintf("PULL REQUESTS (%s)", strings.Join(countParts, ", "))
	}
	content.WriteString(TitleStyle.Render(title) + "\n\n")

	itemIdx := 0

	// Needs Review section
	needsReviewTitle := fmt.Sprintf("Needs Review (%d)", len(panel.NeedsReview))
	content.WriteString(BoldStyle.Render(needsReviewTitle) + "\n")

	if len(panel.NeedsReview) == 0 {
		content.WriteString(DimStyle.Render("  None") + "\n")
	} else {
		for i, pr := range panel.NeedsReview {
			if i >= 4 {
				content.WriteString(DimStyle.Render(fmt.Sprintf("  +%d more...", len(panel.NeedsReview)-4)) + "\n")
				break
			}
			isCursor := selected && itemIdx == cursorIdx
			content.WriteString(renderPrLine(pr, width-4, false, isCursor) + "\n")
			itemIdx++
		}
	}

	content.WriteString("\n")

	// Your PRs section
	yourPrsTitle := fmt.Sprintf("Your PRs (%d)", len(panel.YourPrs))
	content.WriteString(BoldStyle.Render(yourPrsTitle) + "\n")

	if len(panel.YourPrs) == 0 {
		content.WriteString(DimStyle.Render("  None") + "\n")
	} else {
		for i, pr := range panel.YourPrs {
			if i >= 4 {
				content.WriteString(DimStyle.Render(fmt.Sprintf("  +%d more...", len(panel.YourPrs)-4)) + "\n")
				break
			}
			isCursor := selected && itemIdx == cursorIdx
			content.WriteString(renderPrLine(pr, width-4, true, isCursor) + "\n")
			itemIdx++
		}
	}

	return style.Render(content.String())
}

func renderPrLine(pr data.PullRequest, maxWidth int, showStatus bool, isCursor bool) string {
	prefix := ItemPrefix(isCursor)

	// Format: #123 Title 2d [status]
	prNum := fmt.Sprintf("#%d", pr.Number)

	// Age with color
	ageStr := fmt.Sprintf("%dd", pr.AgeDays)
	ageColor := AgeColor(pr.AgeDays)
	ageStyled := lipgloss.NewStyle().Foreground(ageColor).Render(ageStr)

	// Status indicator for user's PRs
	var statusIndicator string
	if showStatus {
		switch pr.ReviewDecision {
		case "APPROVED":
			statusIndicator = lipgloss.NewStyle().Foreground(ColorGreen).Render(" ✓")
		case "CHANGES_REQUESTED":
			statusIndicator = lipgloss.NewStyle().Foreground(ColorRed).Render(" ✗")
		case "REVIEW_REQUIRED":
			statusIndicator = lipgloss.NewStyle().Foreground(ColorYellow).Render(" ○")
		}
	}

	// Calculate available width for title
	fixedWidth := len(prNum) + 1 + len(ageStr) + len(statusIndicator) + 4 // padding and spaces
	titleWidth := maxWidth - fixedWidth
	if titleWidth < 10 {
		titleWidth = 10
	}
	title := Truncate(pr.Title, titleWidth)

	return fmt.Sprintf("%s%s %s %s%s", prefix, prNum, title, ageStyled, statusIndicator)
}
