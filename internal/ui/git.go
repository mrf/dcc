package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrf/dcc/internal/data"
)

// RenderGitPanel renders the git dirty repos panel
func RenderGitPanel(panel data.GitPanel, width, height int, selected, loading bool, cursorIdx int) string {
	style := GetPanelStyle(selected, loading, ColorYellow).
		Width(width).
		Height(height)

	var content strings.Builder

	if loading || panel.IsLoading {
		content.WriteString(TitleStyle.Render("UNCOMMITTED WORK") + "\n\n")
		content.WriteString(ItalicStyle.Render("Scanning repos..."))
		return style.Render(content.String())
	}

	content.WriteString(TitleStyle.Render(TitleWithCount("UNCOMMITTED WORK", len(panel.DirtyRepos))) + "\n\n")

	if len(panel.DirtyRepos) == 0 {
		content.WriteString(DimStyle.Render("All repos clean!"))
		return style.Render(content.String())
	}

	maxRepos := 6
	for i, repo := range panel.DirtyRepos {
		if i >= maxRepos {
			content.WriteString(DimStyle.Render(fmt.Sprintf("+%d more repos...", len(panel.DirtyRepos)-maxRepos)) + "\n")
			break
		}

		isCursor := selected && i == cursorIdx
		content.WriteString(renderDirtyRepoLine(repo, width-4, isCursor) + "\n")
	}

	return style.Render(content.String())
}

func renderDirtyRepoLine(repo data.DirtyRepo, maxWidth int, isCursor bool) string {
	prefix := ItemPrefix(isCursor)
	// Color based on status priority
	statusColor := GitStatusColor(repo.Staged, repo.Modified, repo.Untracked)

	// Repo name (padded to 20 chars)
	repoName := fmt.Sprintf("%-20s", Truncate(repo.Name, 20))
	repoNameStyled := lipgloss.NewStyle().Foreground(statusColor).Render(repoName)

	// Branch name
	branchStr := ""
	if repo.Branch != "" {
		branchStr = DimStyle.Render(repo.Branch) + " "
	}

	// Status counts
	var statusParts []string
	if repo.Staged > 0 {
		statusParts = append(statusParts,
			lipgloss.NewStyle().Foreground(ColorGreen).Render(fmt.Sprintf("%d staged", repo.Staged)))
	}
	if repo.Modified > 0 {
		statusParts = append(statusParts,
			lipgloss.NewStyle().Foreground(ColorYellow).Render(fmt.Sprintf("%d modified", repo.Modified)))
	}
	if repo.Untracked > 0 {
		statusParts = append(statusParts,
			lipgloss.NewStyle().Foreground(ColorCyan).Render(fmt.Sprintf("%d untracked", repo.Untracked)))
	}

	statusStr := strings.Join(statusParts, ", ")

	return fmt.Sprintf("%s%s %s%s", prefix, repoNameStyled, branchStr, statusStr)
}
