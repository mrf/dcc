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

	// Format: #123 Title 2d [CI dot] [review icon]
	prNum := fmt.Sprintf("#%d", pr.Number)

	ageStr := fmt.Sprintf("%dd", pr.AgeDays)
	ageStyled := lipgloss.NewStyle().Foreground(AgeColor(pr.AgeDays)).Render(ageStr)

	var statusIndicator string
	statusVisualWidth := 0

	if showStatus {
		statusIndicator, statusVisualWidth = renderStatusIndicators(pr)
	}

	fixedWidth := len(prNum) + 1 + len(ageStr) + statusVisualWidth + 4
	titleWidth := max(maxWidth-fixedWidth, 10)
	title := Truncate(pr.Title, titleWidth)

	return fmt.Sprintf("%s%s %s %s%s", prefix, prNum, title, ageStyled, statusIndicator)
}

// renderStatusIndicators returns the styled CI and review indicators with their visual width.
func renderStatusIndicators(pr data.PullRequest) (string, int) {
	var indicator string
	visualWidth := 0

	// CI status dot
	if ciStyle, ok := ciStatusStyle(pr.CIStatus); ok {
		indicator += ciStyle.Render(" ●")
		visualWidth += 2
	}

	// Review status icon
	if reviewStyle, symbol, ok := reviewStatusStyle(pr.ReviewDecision); ok {
		indicator += reviewStyle.Render(" " + symbol)
		visualWidth += 2
	}

	return indicator, visualWidth
}

func ciStatusStyle(status string) (lipgloss.Style, bool) {
	switch status {
	case "success":
		return StatusGreen, true
	case "failure":
		return StatusRed, true
	case "pending":
		return StatusYellow, true
	default:
		return lipgloss.Style{}, false
	}
}

func reviewStatusStyle(decision string) (lipgloss.Style, string, bool) {
	switch decision {
	case "APPROVED":
		return StatusGreen, "✓", true
	case "CHANGES_REQUESTED":
		return StatusRed, "✗", true
	case "REVIEW_REQUIRED":
		return StatusYellow, "○", true
	default:
		return lipgloss.Style{}, "", false
	}
}
