package ui

import (
	"testing"

	"github.com/mrf/dcc/internal/data"
)

func makeStash(repo, message string, ageDays int64) data.StashInfo {
	return data.StashInfo{
		Repo:    repo,
		Index:   0,
		Message: message,
		AgeDays: ageDays,
	}
}

// GroupStashesByRepo groups stashes by repository name for testing
func GroupStashesByRepo(stashes []data.StashInfo) map[string][]data.StashInfo {
	result := make(map[string][]data.StashInfo)
	for _, stash := range stashes {
		result[stash.Repo] = append(result[stash.Repo], stash)
	}
	return result
}

func TestGroupStashesByRepoGroupsCorrectly(t *testing.T) {
	stashes := []data.StashInfo{
		makeStash("repo-a", "stash 1", 1),
		makeStash("repo-b", "stash 2", 2),
		makeStash("repo-a", "stash 3", 3),
	}

	grouped := GroupStashesByRepo(stashes)

	if len(grouped) != 2 {
		t.Errorf("expected 2 repos, got %d", len(grouped))
	}

	// repo-a should have 2 stashes
	if len(grouped["repo-a"]) != 2 {
		t.Errorf("expected 2 stashes for repo-a, got %d", len(grouped["repo-a"]))
	}

	// repo-b should have 1 stash
	if len(grouped["repo-b"]) != 1 {
		t.Errorf("expected 1 stash for repo-b, got %d", len(grouped["repo-b"]))
	}
}

func TestGroupStashesEmpty(t *testing.T) {
	stashes := []data.StashInfo{}
	grouped := GroupStashesByRepo(stashes)

	if len(grouped) != 0 {
		t.Errorf("expected 0 repos, got %d", len(grouped))
	}
}

func TestGroupStashesSingleRepo(t *testing.T) {
	stashes := []data.StashInfo{
		makeStash("only-repo", "stash 1", 1),
		makeStash("only-repo", "stash 2", 2),
		makeStash("only-repo", "stash 3", 3),
	}

	grouped := GroupStashesByRepo(stashes)

	if len(grouped) != 1 {
		t.Errorf("expected 1 repo, got %d", len(grouped))
	}

	if len(grouped["only-repo"]) != 3 {
		t.Errorf("expected 3 stashes, got %d", len(grouped["only-repo"]))
	}
}
