package data

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/mrf/dcc/internal/config"
)

// Meeting represents a calendar meeting
type Meeting struct {
	Title    string
	Start    time.Time
	End      time.Time
	Calendar string
}

// MeetingStatus represents the current meeting status
type MeetingStatus int

const (
	StatusFree MeetingStatus = iota
	StatusInMeeting
	StatusClear
)

// MeetingsPanel holds meeting data for display
type MeetingsPanel struct {
	NextMeeting   *Meeting
	Upcoming      []Meeting
	Status        MeetingStatus
	MinutesUntil  int64
	EndsIn        int64
	IsLoading     bool
	Unsupported   bool
}

// FetchMeetings retrieves meetings from Calendar.app (macOS only)
func FetchMeetings(cfg config.MeetingsConfig) MeetingsPanel {
	if !cfg.Enabled {
		return MeetingsPanel{Status: StatusClear}
	}

	// Only works on macOS
	if runtime.GOOS != "darwin" {
		return MeetingsPanel{Status: StatusClear, Unsupported: true}
	}

	script := buildAppleScript(cfg)
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return MeetingsPanel{Status: StatusClear}
	}

	meetings := parseMeetings(string(output), cfg.IgnorePatterns)
	return buildMeetingsPanel(meetings)
}

func buildAppleScript(cfg config.MeetingsConfig) string {
	excludeList := formatAppleScriptList(cfg.CalendarsExclude)

	return fmt.Sprintf(`
tell application "Calendar"
    set now to current date
    set endTime to now + (%d * hours)
    set output to ""
    repeat with cal in calendars
        if name of cal is not in %s then
            try
                set evts to (every event of cal whose start date >= now and start date <= endTime)
                repeat with evt in evts
                    set output to output & (summary of evt) & "|||" & (start date of evt) & "|||" & (end date of evt) & "|||" & (name of cal) & "\n"
                end repeat
            end try
        end if
    end repeat
    return output
end tell
`, cfg.HoursAhead, excludeList)
}

func formatAppleScriptList(items []string) string {
	if len(items) == 0 {
		return "{}"
	}
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf(`"%s"`, item)
	}
	return "{" + strings.Join(quoted, ", ") + "}"
}

func parseMeetings(output string, ignorePatterns []string) []Meeting {
	var meetings []Meeting
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|||")
		if len(parts) != 4 {
			continue
		}

		title := strings.TrimSpace(parts[0])

		// Check if title matches ignore patterns
		skip := false
		for _, pattern := range ignorePatterns {
			if strings.Contains(strings.ToLower(title), strings.ToLower(pattern)) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		start := parseAppleScriptDate(strings.TrimSpace(parts[1]))
		end := parseAppleScriptDate(strings.TrimSpace(parts[2]))
		calendar := strings.TrimSpace(parts[3])

		if start.IsZero() || end.IsZero() {
			continue
		}

		meetings = append(meetings, Meeting{
			Title:    title,
			Start:    start,
			End:      end,
			Calendar: calendar,
		})
	}

	// Sort by start time
	sort.Slice(meetings, func(i, j int) bool {
		return meetings[i].Start.Before(meetings[j].Start)
	})

	return meetings
}

func parseAppleScriptDate(dateStr string) time.Time {
	// AppleScript returns dates in various formats depending on system locale
	// Common formats: "Friday, January 26, 2024 at 10:00:00 AM"
	//                 "January 26, 2024 at 10:00:00 AM"
	//                 "26 Jan 2024, 10:00 AM"
	formats := []string{
		"Monday, January 2, 2006 at 3:04:05 PM",
		"January 2, 2006 at 3:04:05 PM",
		"Monday, January 2, 2006 at 15:04:05",
		"January 2, 2006 at 15:04:05",
		"2 Jan 2006, 3:04 PM",
		"2 Jan 2006, 15:04",
		"Monday, 2 January 2006 at 3:04:05 PM",
		"2 January 2006 at 3:04:05 PM",
		"Monday, 2 January 2006 at 15:04:05",
		"2 January 2006 at 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, time.Local); err == nil {
			return t
		}
	}

	return time.Time{}
}

func buildMeetingsPanel(meetings []Meeting) MeetingsPanel {
	now := time.Now()
	panel := MeetingsPanel{Status: StatusClear}

	if len(meetings) == 0 {
		return panel
	}

	// Check if currently in a meeting
	for i, m := range meetings {
		if now.After(m.Start) && now.Before(m.End) {
			panel.Status = StatusInMeeting
			panel.NextMeeting = &meetings[i]
			panel.EndsIn = int64(m.End.Sub(now).Minutes())
			if i+1 < len(meetings) {
				panel.Upcoming = meetings[i+1:]
			}
			return panel
		}
	}

	// Find next meeting
	for i, m := range meetings {
		if m.Start.After(now) {
			panel.Status = StatusFree
			panel.NextMeeting = &meetings[i]
			panel.MinutesUntil = int64(m.Start.Sub(now).Minutes())
			if i+1 < len(meetings) {
				panel.Upcoming = meetings[i+1:]
			}
			return panel
		}
	}

	return panel
}
