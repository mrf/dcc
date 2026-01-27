package data

import "testing"

func TestParseGitStatusModifiedFiles(t *testing.T) {
	output := " M src/main.rs\n M src/lib.rs\n"
	changes := parseGitStatus(output)

	if changes.Modified != 2 {
		t.Errorf("expected 2 modified, got %d", changes.Modified)
	}
	if changes.Staged != 0 {
		t.Errorf("expected 0 staged, got %d", changes.Staged)
	}
	if changes.Untracked != 0 {
		t.Errorf("expected 0 untracked, got %d", changes.Untracked)
	}
}

func TestParseGitStatusStagedFiles(t *testing.T) {
	output := "M  src/main.rs\nA  src/new.rs\n"
	changes := parseGitStatus(output)

	if changes.Staged != 2 {
		t.Errorf("expected 2 staged, got %d", changes.Staged)
	}
	if changes.Modified != 0 {
		t.Errorf("expected 0 modified, got %d", changes.Modified)
	}
	if changes.Untracked != 0 {
		t.Errorf("expected 0 untracked, got %d", changes.Untracked)
	}
}

func TestParseGitStatusUntrackedFiles(t *testing.T) {
	output := "?? newfile.txt\n?? another.txt\n"
	changes := parseGitStatus(output)

	if changes.Untracked != 2 {
		t.Errorf("expected 2 untracked, got %d", changes.Untracked)
	}
	if changes.Modified != 0 {
		t.Errorf("expected 0 modified, got %d", changes.Modified)
	}
	if changes.Staged != 0 {
		t.Errorf("expected 0 staged, got %d", changes.Staged)
	}
}

func TestParseGitStatusMixed(t *testing.T) {
	output := " M modified.rs\nM  staged.rs\n?? untracked.txt\n"
	changes := parseGitStatus(output)

	if changes.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", changes.Modified)
	}
	if changes.Staged != 1 {
		t.Errorf("expected 1 staged, got %d", changes.Staged)
	}
	if changes.Untracked != 1 {
		t.Errorf("expected 1 untracked, got %d", changes.Untracked)
	}
}

func TestParseGitStatusEmpty(t *testing.T) {
	changes := parseGitStatus("")

	if changes.Modified != 0 || changes.Staged != 0 || changes.Untracked != 0 {
		t.Errorf("expected no changes for empty output")
	}
}

func TestParseStashList(t *testing.T) {
	output := "stash@{0}|||WIP on main: abc123 some commit|||3 days ago\nstash@{1}|||feature work|||2 weeks ago\n"
	stashes := parseGitStashes(output, "test-repo")

	if len(stashes) != 2 {
		t.Fatalf("expected 2 stashes, got %d", len(stashes))
	}

	if stashes[0].Repo != "test-repo" {
		t.Errorf("expected repo 'test-repo', got '%s'", stashes[0].Repo)
	}
	if stashes[0].Index != 0 {
		t.Errorf("expected index 0, got %d", stashes[0].Index)
	}
	if stashes[0].Message != "WIP on main: abc123 some commit" {
		t.Errorf("expected message 'WIP on main: abc123 some commit', got '%s'", stashes[0].Message)
	}
	if stashes[0].AgeDays != 3 {
		t.Errorf("expected age 3 days, got %d", stashes[0].AgeDays)
	}

	if stashes[1].Index != 1 {
		t.Errorf("expected index 1, got %d", stashes[1].Index)
	}
	if stashes[1].Message != "feature work" {
		t.Errorf("expected message 'feature work', got '%s'", stashes[1].Message)
	}
	if stashes[1].AgeDays != 14 { // 2 weeks
		t.Errorf("expected age 14 days, got %d", stashes[1].AgeDays)
	}
}

func TestParseStashListEmpty(t *testing.T) {
	stashes := parseGitStashes("", "test-repo")

	if len(stashes) != 0 {
		t.Errorf("expected 0 stashes, got %d", len(stashes))
	}
}

func TestParseRelativeTimeDays(t *testing.T) {
	if parseRelativeTime("3 days ago") != 3 {
		t.Errorf("expected 3 days")
	}
	if parseRelativeTime("1 day ago") != 1 {
		t.Errorf("expected 1 day")
	}
}

func TestParseRelativeTimeWeeks(t *testing.T) {
	if parseRelativeTime("2 weeks ago") != 14 {
		t.Errorf("expected 14 days for 2 weeks")
	}
	if parseRelativeTime("1 week ago") != 7 {
		t.Errorf("expected 7 days for 1 week")
	}
}

func TestParseRelativeTimeMonths(t *testing.T) {
	if parseRelativeTime("1 month ago") != 30 {
		t.Errorf("expected 30 days for 1 month")
	}
	if parseRelativeTime("3 months ago") != 90 {
		t.Errorf("expected 90 days for 3 months")
	}
}

func TestParseRelativeTimeHoursMinutesSecondsAreZero(t *testing.T) {
	// These should all return 0 since they're less than a day
	if parseRelativeTime("5 hours ago") != 0 {
		t.Errorf("expected 0 for hours")
	}
	if parseRelativeTime("30 minutes ago") != 0 {
		t.Errorf("expected 0 for minutes")
	}
	if parseRelativeTime("10 seconds ago") != 0 {
		t.Errorf("expected 0 for seconds")
	}
}

func TestParseRelativeTimeInvalid(t *testing.T) {
	if parseRelativeTime("invalid") != 0 {
		t.Errorf("expected 0 for invalid")
	}
	if parseRelativeTime("") != 0 {
		t.Errorf("expected 0 for empty")
	}
}

func TestParseStashIndex(t *testing.T) {
	if parseStashIndex("stash@{0}") != 0 {
		t.Errorf("expected 0")
	}
	if parseStashIndex("stash@{5}") != 5 {
		t.Errorf("expected 5")
	}
	if parseStashIndex("stash@{123}") != 123 {
		t.Errorf("expected 123")
	}
	if parseStashIndex("invalid") != 0 {
		t.Errorf("expected 0 for invalid")
	}
}
