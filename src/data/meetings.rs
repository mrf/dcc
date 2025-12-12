use anyhow::Result;
use chrono::{DateTime, Local, NaiveDateTime, TimeZone};
use std::process::Command;

use super::config::MeetingsConfig;

#[derive(Debug, Clone)]
pub struct MeetingsPanel {
    pub next_meeting: Option<Meeting>,
    pub upcoming: Vec<Meeting>,
    pub status: MeetingStatus,
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct Meeting {
    pub title: String,
    pub start: DateTime<Local>,
    pub end: DateTime<Local>,
    pub calendar: String,
}

#[derive(Debug, Clone)]
pub enum MeetingStatus {
    Free { minutes_until: i64 },
    InMeeting { ends_in: i64 },
    Clear,
}

impl Default for MeetingsPanel {
    fn default() -> Self {
        Self {
            next_meeting: None,
            upcoming: Vec::new(),
            status: MeetingStatus::Clear,
        }
    }
}

pub fn fetch_meetings(config: &MeetingsConfig) -> Result<MeetingsPanel> {
    if !config.enabled {
        return Ok(MeetingsPanel::default());
    }

    let excluded_cals: Vec<&str> = config
        .calendars_exclude
        .iter()
        .map(|s| s.as_str())
        .collect();
    let excluded_list = excluded_cals
        .iter()
        .map(|s| format!("\"{}\"", s))
        .collect::<Vec<_>>()
        .join(", ");

    let script = format!(
        r#"
        tell application "Calendar"
            set now to current date
            set endTime to now + ({} * hours)
            set output to ""
            repeat with cal in calendars
                if name of cal is not in {{{excluded_list}}} then
                    try
                        set evts to (every event of cal whose start date ≥ now and start date ≤ endTime)
                        repeat with evt in evts
                            set output to output & (summary of evt) & "|||"
                            set output to output & (start date of evt as string) & "|||"
                            set output to output & (end date of evt as string) & "|||"
                            set output to output & (name of cal) & "
"
                        end repeat
                    end try
                end if
            end repeat
            return output
        end tell
        "#,
        config.hours_ahead,
    );

    let output = Command::new("osascript").arg("-e").arg(&script).output();

    match output {
        Ok(output) => {
            let stdout = String::from_utf8_lossy(&output.stdout);
            parse_meetings(&stdout, config)
        }
        Err(_) => Ok(MeetingsPanel::default()),
    }
}

fn parse_meetings(output: &str, config: &MeetingsConfig) -> Result<MeetingsPanel> {
    let now = Local::now();
    let mut meetings: Vec<Meeting> = Vec::new();

    for line in output.lines() {
        if line.trim().is_empty() {
            continue;
        }

        let parts: Vec<&str> = line.split("|||").collect();
        if parts.len() < 4 {
            continue;
        }

        let title = parts[0].trim().to_string();

        // Skip ignored patterns
        if config
            .ignore_patterns
            .iter()
            .any(|p| title.to_lowercase().contains(&p.to_lowercase()))
        {
            continue;
        }

        let start = parse_applescript_date(parts[1].trim());
        let end = parse_applescript_date(parts[2].trim());
        let calendar = parts[3].trim().to_string();

        if let (Some(start), Some(end)) = (start, end) {
            meetings.push(Meeting {
                title,
                start,
                end,
                calendar,
            });
        }
    }

    // Sort by start time
    meetings.sort_by(|a, b| a.start.cmp(&b.start));

    // Determine status
    let status = if let Some(first) = meetings.first() {
        if first.start <= now && first.end > now {
            // Currently in a meeting
            let ends_in = (first.end - now).num_minutes();
            MeetingStatus::InMeeting { ends_in }
        } else {
            // Free until next meeting
            let minutes_until = (first.start - now).num_minutes();
            MeetingStatus::Free { minutes_until }
        }
    } else {
        MeetingStatus::Clear
    };

    let next_meeting = meetings.first().cloned();
    let upcoming = if meetings.len() > 1 {
        meetings[1..].to_vec()
    } else {
        Vec::new()
    };

    Ok(MeetingsPanel {
        next_meeting,
        upcoming,
        status,
    })
}

fn parse_applescript_date(date_str: &str) -> Option<DateTime<Local>> {
    // AppleScript dates look like: "Thursday, December 12, 2024 at 2:00:00 PM"
    // or sometimes: "December 12, 2024 2:00:00 PM"

    // Try multiple formats
    let formats = [
        "%A, %B %d, %Y at %I:%M:%S %p",
        "%B %d, %Y at %I:%M:%S %p",
        "%A, %B %d, %Y %I:%M:%S %p",
        "%B %d, %Y %I:%M:%S %p",
        "%m/%d/%Y %I:%M:%S %p",
        "%m/%d/%y %I:%M:%S %p",
    ];

    for fmt in &formats {
        if let Ok(naive) = NaiveDateTime::parse_from_str(date_str, fmt) {
            return Local.from_local_datetime(&naive).single();
        }
    }

    None
}
