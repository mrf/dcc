package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrf/dcc/internal/data"
)

// RenderPortsPanel renders the ports panel
func RenderPortsPanel(panel data.PortsPanel, width, height int, selected, loading bool, cursorIdx int) string {
	style := GetPanelStyle(selected, loading, ColorCyan).
		Width(width).
		Height(height)

	var content strings.Builder

	if loading || panel.IsLoading {
		content.WriteString(TitleStyle.Render("PORTS") + "\n\n")
		content.WriteString(ItalicStyle.Render("Scanning ports..."))
		return style.Render(content.String())
	}

	content.WriteString(TitleStyle.Render(TitleWithCount("PORTS", len(panel.Ports))) + "\n\n")

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

		isCursor := selected && i == cursorIdx
		prefix := ItemPrefix(isCursor)

		portColor := PortColor(port.Port)
		portStr := lipgloss.NewStyle().Foreground(portColor).Render(fmt.Sprintf(":%d", port.Port))

		processName := Truncate(port.Process, width-12)

		content.WriteString(fmt.Sprintf("%s%s %s\n", prefix, portStr, processName))
	}

	return style.Render(content.String())
}
