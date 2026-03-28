package app

import (
	"testing"
	"time"

	"github.com/mrf/dcc/internal/data"
)

func TestDetectPrChanges_NewReviewRequest(t *testing.T) {
	prev := data.PrsPanel{
		NeedsReview: []data.PullRequest{
			{Number: 1, Repo: "org/repo", Title: "old PR"},
		},
	}
	curr := data.PrsPanel{
		NeedsReview: []data.PullRequest{
			{Number: 1, Repo: "org/repo", Title: "old PR"},
			{Number: 2, Repo: "org/repo", Title: "new PR"},
		},
	}

	notes := detectPrChanges(prev, curr)
	if len(notes) != 1 {
		t.Fatalf("expected 1 notification, got %d: %v", len(notes), notes)
	}
	if notes[0] != "New review request: org/repo #2" {
		t.Errorf("unexpected notification: %s", notes[0])
	}
}

func TestDetectPrChanges_PrApproved(t *testing.T) {
	prev := data.PrsPanel{
		YourPrs: []data.PullRequest{
			{Number: 10, Repo: "org/repo", ReviewDecision: "REVIEW_REQUIRED"},
		},
	}
	curr := data.PrsPanel{
		YourPrs: []data.PullRequest{
			{Number: 10, Repo: "org/repo", ReviewDecision: "APPROVED"},
		},
	}

	notes := detectPrChanges(prev, curr)
	if len(notes) != 1 {
		t.Fatalf("expected 1 notification, got %d: %v", len(notes), notes)
	}
	if notes[0] != "PR approved: org/repo #10" {
		t.Errorf("unexpected notification: %s", notes[0])
	}
}

func TestDetectPrChanges_ChangesRequested(t *testing.T) {
	prev := data.PrsPanel{
		YourPrs: []data.PullRequest{
			{Number: 10, Repo: "org/repo", ReviewDecision: "REVIEW_REQUIRED"},
		},
	}
	curr := data.PrsPanel{
		YourPrs: []data.PullRequest{
			{Number: 10, Repo: "org/repo", ReviewDecision: "CHANGES_REQUESTED"},
		},
	}

	notes := detectPrChanges(prev, curr)
	if len(notes) != 1 {
		t.Fatalf("expected 1 notification, got %d: %v", len(notes), notes)
	}
	if notes[0] != "Changes requested: org/repo #10" {
		t.Errorf("unexpected notification: %s", notes[0])
	}
}

func TestDetectPrChanges_NoChange(t *testing.T) {
	prs := data.PrsPanel{
		NeedsReview: []data.PullRequest{
			{Number: 1, Repo: "org/repo"},
		},
		YourPrs: []data.PullRequest{
			{Number: 10, Repo: "org/repo", ReviewDecision: "APPROVED"},
		},
	}

	notes := detectPrChanges(prs, prs)
	if len(notes) != 0 {
		t.Errorf("expected no notifications, got %d: %v", len(notes), notes)
	}
}

func TestDetectMeetingSoon(t *testing.T) {
	meeting := data.Meeting{Title: "Standup", Start: time.Now().Add(3 * time.Minute)}

	prev := data.MeetingsPanel{Status: data.StatusFree, MinutesUntil: 10}
	curr := data.MeetingsPanel{
		Status:       data.StatusFree,
		NextMeeting:  &meeting,
		MinutesUntil: 3,
	}

	note := detectMeetingSoon(prev, curr)
	if note == "" {
		t.Fatal("expected meeting notification")
	}
}

func TestDetectMeetingSoon_AlreadyNotified(t *testing.T) {
	meeting := data.Meeting{Title: "Standup", Start: time.Now().Add(3 * time.Minute)}

	// Both prev and curr are within 5 minutes — should not re-notify
	prev := data.MeetingsPanel{
		Status:       data.StatusFree,
		NextMeeting:  &meeting,
		MinutesUntil: 4,
	}
	curr := data.MeetingsPanel{
		Status:       data.StatusFree,
		NextMeeting:  &meeting,
		MinutesUntil: 3,
	}

	note := detectMeetingSoon(prev, curr)
	if note != "" {
		t.Errorf("expected no notification (already within threshold), got: %s", note)
	}
}
