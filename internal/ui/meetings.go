package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrf/dcc/internal/data"
)

// RenderMeetingsPanel renders the meetings panel
func RenderMeetingsPanel(panel data.MeetingsPanel, width, height int, selected, loading bool) string {
	// Determine border color based on status
	borderColor := determineMeetingsBorderColor(panel)

	style := GetPanelStyle(selected, loading, borderColor).
		Width(width).
		Height(height)

	title := TitleStyle.Render("MEETINGS")

	var content strings.Builder
	content.WriteString(title + "\n\n")

	if loading || panel.IsLoading {
		content.WriteString(ItalicStyle.Render("Checking calendar..."))
		return style.Render(content.String())
	}

	if panel.Unsupported {
		content.WriteString(DimStyle.Render("Calendar integration not available\non this platform (macOS only)"))
		return style.Render(content.String())
	}

	// Render status line
	statusLine := renderMeetingStatus(panel)
	content.WriteString(statusLine + "\n")

	// Render next meeting
	if panel.NextMeeting != nil {
		meetingTitle := Truncate(panel.NextMeeting.Title, 25)
		content.WriteString(meetingTitle + "\n")
	}

	// Render upcoming meetings (up to 3)
	if len(panel.Upcoming) > 0 {
		content.WriteString("\n" + DimStyle.Render("Then:") + "\n")
		for i, m := range panel.Upcoming {
			if i >= 3 {
				break
			}
			timeStr := m.Start.Format("3:04 PM")
			meetingTitle := Truncate(m.Title, width-10)
			content.WriteString(fmt.Sprintf("%s  %s\n", DimStyle.Render(timeStr), meetingTitle))
		}
	}

	return style.Render(content.String())
}

func determineMeetingsBorderColor(panel data.MeetingsPanel) lipgloss.Color {
	if panel.IsLoading || panel.Unsupported {
		return ColorDarkGray
	}

	switch panel.Status {
	case data.StatusInMeeting:
		return ColorBlue
	case data.StatusFree:
		return MeetingStatusColor(panel.MinutesUntil, false)
	default:
		return ColorGreen
	}
}

func renderMeetingStatus(panel data.MeetingsPanel) string {
	var statusColor lipgloss.Color
	var statusText string

	switch panel.Status {
	case data.StatusInMeeting:
		statusColor = ColorBlue
		statusText = fmt.Sprintf("In meeting (%dm left)", panel.EndsIn)
	case data.StatusFree:
		statusColor = MeetingStatusColor(panel.MinutesUntil, false)
		if panel.MinutesUntil >= 60 {
			hours := panel.MinutesUntil / 60
			mins := panel.MinutesUntil % 60
			if mins > 0 {
				statusText = fmt.Sprintf("Free for %dh %dm", hours, mins)
			} else {
				statusText = fmt.Sprintf("Free for %dh", hours)
			}
		} else {
			statusText = fmt.Sprintf("Free for %dm", panel.MinutesUntil)
		}
	default:
		statusColor = ColorGreen
		statusText = "All clear!"
	}

	bullet := lipgloss.NewStyle().Foreground(statusColor).Render("●")
	return bullet + " " + statusText
}
