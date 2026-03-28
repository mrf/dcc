package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrf/dcc/internal/data"
)

const notificationTimeout = 10 * time.Second

// clearNotificationCmd returns a command that sends ClearNotificationMsg after a delay
func clearNotificationCmd() tea.Cmd {
	return tea.Tick(notificationTimeout, func(time.Time) tea.Msg {
		return ClearNotificationMsg{}
	})
}

// detectPrChanges compares previous and current PR data, returning notification messages
func detectPrChanges(prev, curr data.PrsPanel) []string {
	var notes []string

	// Detect new PRs needing review
	prevReviewNums := make(map[int]bool, len(prev.NeedsReview))
	for _, pr := range prev.NeedsReview {
		prevReviewNums[pr.Number] = true
	}
	for _, pr := range curr.NeedsReview {
		if !prevReviewNums[pr.Number] {
			notes = append(notes, fmt.Sprintf("New review request: %s #%d", pr.Repo, pr.Number))
		}
	}

	// Detect approval/changes on your PRs
	prevYourPrs := make(map[int]string, len(prev.YourPrs))
	for _, pr := range prev.YourPrs {
		prevYourPrs[pr.Number] = pr.ReviewDecision
	}
	for _, pr := range curr.YourPrs {
		prevDecision, existed := prevYourPrs[pr.Number]
		if !existed {
			continue
		}
		if pr.ReviewDecision != prevDecision {
			switch pr.ReviewDecision {
			case "APPROVED":
				notes = append(notes, fmt.Sprintf("PR approved: %s #%d", pr.Repo, pr.Number))
			case "CHANGES_REQUESTED":
				notes = append(notes, fmt.Sprintf("Changes requested: %s #%d", pr.Repo, pr.Number))
			}
		}
	}

	return notes
}

// detectMeetingSoon returns a notification if a meeting starts within 5 minutes
func detectMeetingSoon(prev, curr data.MeetingsPanel) string {
	if curr.Status == data.StatusFree && curr.NextMeeting != nil && curr.MinutesUntil <= 5 {
		// Only notify if we crossed the 5-minute threshold
		if prev.NextMeeting == nil || prev.MinutesUntil > 5 {
			return fmt.Sprintf("Meeting in %dm: %s", curr.MinutesUntil, curr.NextMeeting.Title)
		}
	}
	return ""
}
