package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrf/dcc/internal/data"
)

// RenderStashesPanel renders the stashes panel
func RenderStashesPanel(panel data.GitPanel, width, height int, selected, loading bool) string {
	style := GetPanelStyle(selected, loading, ColorMagenta).
		Width(width).
		Height(height)

	title := TitleStyle.Render("STASHES")

	var content strings.Builder
	content.WriteString(title + "\n\n")

	if loading || panel.IsLoading {
		content.WriteString(ItalicStyle.Render("Scanning stashes..."))
		return style.Render(content.String())
	}

	if len(panel.Stashes) == 0 {
		content.WriteString(DimStyle.Render("No stashes"))
		return style.Render(content.String())
	}

	// Group stashes by repo
	stashesByRepo := make(map[string][]data.StashInfo)
	var repoOrder []string
	for _, stash := range panel.Stashes {
		if _, exists := stashesByRepo[stash.Repo]; !exists {
			repoOrder = append(repoOrder, stash.Repo)
		}
		stashesByRepo[stash.Repo] = append(stashesByRepo[stash.Repo], stash)
	}

	maxRepos := 4
	reposShown := 0
	for _, repo := range repoOrder {
		if reposShown >= maxRepos {
			remaining := len(repoOrder) - maxRepos
			content.WriteString(DimStyle.Render(fmt.Sprintf("+%d more repos...", remaining)) + "\n")
			break
		}

		stashes := stashesByRepo[repo]
		content.WriteString(renderStashGroup(repo, stashes, width-4) + "\n")
		reposShown++
	}

	return style.Render(content.String())
}

func renderStashGroup(repo string, stashes []data.StashInfo, maxWidth int) string {
	// Format: repo_name (count) first_message age
	count := len(stashes)
	firstStash := stashes[0]

	// Age indicator
	ageColor := StashAgeColor(firstStash.AgeDays)
	var ageStr string
	if firstStash.AgeDays > 30 {
		ageStr = "(ancient)"
	} else if firstStash.AgeDays > 0 {
		ageStr = fmt.Sprintf("(%dd)", firstStash.AgeDays)
	} else {
		ageStr = "(today)"
	}
	ageStyled := lipgloss.NewStyle().Foreground(ageColor).Render(ageStr)

	// Repo name with count
	repoWithCount := fmt.Sprintf("%s (%d)", repo, count)
	repoStyled := BoldStyle.Render(repoWithCount)

	// Message (truncated)
	messageWidth := maxWidth - len(repoWithCount) - len(ageStr) - 6
	if messageWidth < 10 {
		messageWidth = 10
	}
	message := Truncate(firstStash.Message, messageWidth)

	return fmt.Sprintf("  %s %s %s", repoStyled, message, ageStyled)
}
