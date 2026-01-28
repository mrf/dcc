package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrf/dcc/internal/data"
)

// RenderPortsPanel renders the ports panel
func RenderPortsPanel(panel data.PortsPanel, width, height int, selected, loading bool) string {
	style := GetPanelStyle(selected, loading, ColorCyan).
		Width(width).
		Height(height)

	title := TitleStyle.Render("PORTS")

	var content strings.Builder
	content.WriteString(title + "\n\n")

	if loading || panel.IsLoading {
		content.WriteString(ItalicStyle.Render("Scanning ports..."))
		return style.Render(content.String())
	}

	if len(panel.Ports) == 0 {
		content.WriteString(DimStyle.Render("No listening ports"))
		return style.Render(content.String())
	}

	maxPorts := 10
	for i, port := range panel.Ports {
		if i >= maxPorts {
			content.WriteString(DimStyle.Render(fmt.Sprintf("+%d more...", len(panel.Ports)-maxPorts)) + "\n")
			break
		}

		portColor := PortColor(port.Port)
		portStr := lipgloss.NewStyle().Foreground(portColor).Render(fmt.Sprintf(":%d", port.Port))

		processName := Truncate(port.Process, width-12)

		content.WriteString(fmt.Sprintf("%s %s\n", portStr, processName))
	}

	return style.Render(content.String())
}
