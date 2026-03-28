package data

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mrf/dcc/internal/config"
)

// DirtyRepo represents a git repository with uncommitted changes
type DirtyRepo struct {
	Name      string
	Path      string
	Branch    string
	Modified  int
	Untracked int
	Staged    int
}

// StashInfo represents a git stash entry
type StashInfo struct {
	Repo    string
	Path    string
	Index   int
	Message string
	AgeDays int64
}

// GitPanel holds git data for display
type GitPanel struct {
	DirtyRepos []DirtyRepo
	Stashes    []StashInfo
	IsLoading  bool
}

// FetchGitStatus scans for git repos and checks their status
func FetchGitStatus(cfg config.GitConfig, projectsDir string) GitPanel {
	if !cfg.Enabled {
		return GitPanel{}
	}

	// Build ignore dirs map
	ignoreDirs := make(map[string]bool)
	for _, dir := range cfg.IgnoreDirs {
		ignoreDirs[dir] = true
	}

	if projectsDir == "" {
		return GitPanel{}
	}

	repos := findGitRepos(projectsDir, cfg.ScanDepth, ignoreDirs)

	var dirtyRepos []DirtyRepo
	var allStashes []StashInfo

	for _, repoPath := range repos {
		repoName := filepath.Base(repoPath)

		// Check git status
		changes := getGitChanges(repoPath)
		if changes.Modified > 0 || changes.Untracked > 0 || changes.Staged > 0 {
			dirtyRepos = append(dirtyRepos, DirtyRepo{
				Name:      repoName,
				Path:      repoPath,
				Branch:    getGitBranch(repoPath),
				Modified:  changes.Modified,
				Untracked: changes.Untracked,
				Staged:    changes.Staged,
			})
		}

		// Check stashes
		stashes := getGitStashes(repoPath, repoName)
		allStashes = append(allStashes, stashes...)
	}

	// Sort dirty repos alphabetically
	sort.Slice(dirtyRepos, func(i, j int) bool {
		return dirtyRepos[i].Name < dirtyRepos[j].Name
	})

	// Sort stashes by repo name, then by index
	sort.Slice(allStashes, func(i, j int) bool {
		if allStashes[i].Repo != allStashes[j].Repo {
			return allStashes[i].Repo < allStashes[j].Repo
		}
		return allStashes[i].Index < allStashes[j].Index
	})

	return GitPanel{
		DirtyRepos: dirtyRepos,
		Stashes:    allStashes,
	}
}

func findGitRepos(root string, maxDepth int, ignoreDirs map[string]bool) []string {
	var repos []string
	scanDir(root, 0, maxDepth, ignoreDirs, &repos)
	return repos
}

func scanDir(dir string, depth, maxDepth int, ignoreDirs map[string]bool, repos *[]string) {
	if depth > maxDepth {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden directories
		if strings.HasPrefix(name, ".") {
			continue
		}

		// Skip ignored directories
		if ignoreDirs[name] {
			continue
		}

		path := filepath.Join(dir, name)

		// Check if this is a git repo
		gitDir := filepath.Join(path, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			*repos = append(*repos, path)
			// Don't recurse into git repos
			continue
		}

		// Recurse
		scanDir(path, depth+1, maxDepth, ignoreDirs, repos)
	}
}

type gitChanges struct {
	Modified  int
	Untracked int
	Staged    int
}

func getGitBranch(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getGitChanges(repoPath string) gitChanges {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return gitChanges{}
	}

	return parseGitStatus(string(output))
}

func parseGitStatus(output string) gitChanges {
	var changes gitChanges
	// Don't use TrimSpace - leading spaces are significant in git status output
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		index := line[0]
		worktree := line[1]

		// Staged: index has a change (not ' ' or '?')
		if index != ' ' && index != '?' {
			changes.Staged++
		}

		// Modified: worktree has M or D
		if worktree == 'M' || worktree == 'D' {
			changes.Modified++
		}

		// Untracked: index is '?'
		if index == '?' {
			changes.Untracked++
		}
	}

	return changes
}

func getGitStashes(repoPath, repoName string) []StashInfo {
	cmd := exec.Command("git", "-C", repoPath, "stash", "list", "--format=%gd|||%s|||%ar")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	stashes := parseGitStashes(string(output), repoName)
	for i := range stashes {
		stashes[i].Path = repoPath
	}
	return stashes
}

func parseGitStashes(output, repoName string) []StashInfo {
	var stashes []StashInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|||")
		if len(parts) != 3 {
			continue
		}

		// Parse stash index from "stash@{0}"
		refStr := parts[0]
		index := parseStashIndex(refStr)

		message := parts[1]
		ageDays := parseRelativeTime(parts[2])

		stashes = append(stashes, StashInfo{
			Repo:    repoName,
			Index:   index,
			Message: message,
			AgeDays: ageDays,
		})
	}

	return stashes
}

func parseStashIndex(ref string) int {
	// Extract number from "stash@{N}"
	start := strings.Index(ref, "{")
	end := strings.Index(ref, "}")
	if start == -1 || end == -1 || end <= start {
		return 0
	}

	numStr := ref[start+1 : end]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return num
}

func parseRelativeTime(timeStr string) int64 {
	// Parse strings like "3 days ago", "1 week ago", "2 months ago"
	parts := strings.Fields(timeStr)
	if len(parts) < 2 {
		return 0
	}

	num, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0
	}

	unit := strings.ToLower(parts[1])

	switch {
	case strings.HasPrefix(unit, "second"), strings.HasPrefix(unit, "minute"), strings.HasPrefix(unit, "hour"):
		return 0 // Less than a day
	case strings.HasPrefix(unit, "day"):
		return num
	case strings.HasPrefix(unit, "week"):
		return num * 7
	case strings.HasPrefix(unit, "month"):
		return num * 30
	case strings.HasPrefix(unit, "year"):
		return num * 365
	}

	return 0
}

// OpenDirtyRepoByIndex opens the dirty repo at the given cursor index in editor
func OpenDirtyRepoByIndex(panel GitPanel, idx int) error {
	if idx < 0 || idx >= len(panel.DirtyRepos) || idx >= 6 {
		return nil
	}

	return openInEditor(panel.DirtyRepos[idx].Path)
}

// OpenStashRepoByIndex opens the stash repo at the given cursor index in editor.
// Items are indexed by repo group in display order (up to 4 groups).
func OpenStashRepoByIndex(panel GitPanel, idx int) error {
	var repoOrder []string
	seen := make(map[string]bool)
	for _, s := range panel.Stashes {
		if !seen[s.Repo] {
			seen[s.Repo] = true
			repoOrder = append(repoOrder, s.Repo)
		}
	}

	if idx < 0 || idx >= len(repoOrder) || idx >= 4 {
		return nil
	}

	targetRepo := repoOrder[idx]
	for _, s := range panel.Stashes {
		if s.Repo == targetRepo && s.Path != "" {
			return openInEditor(s.Path)
		}
	}
	return nil
}

func openInEditor(path string) error {
	if _, err := exec.LookPath("code"); err == nil {
		return exec.Command("code", path).Run()
	}
	return exec.Command("open", path).Run()
}
