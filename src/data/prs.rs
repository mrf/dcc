use anyhow::Result;
use chrono::{DateTime, Utc};
use serde::Deserialize;
use std::process::{Command, Stdio};
use std::thread;

use super::config::PrsConfig;

#[derive(Debug, Clone, Default)]
pub struct PrsPanel {
    pub needs_review: Vec<PullRequest>,
    pub your_prs: Vec<PullRequest>,
}

#[derive(Debug, Clone, PartialEq)]
#[allow(dead_code)]
pub struct PullRequest {
    pub number: u32,
    pub title: String,
    pub repo: String,
    pub url: String,
    pub age_days: i64,
    pub review_decision: Option<String>,
}

// For gh search prs output
#[derive(Debug, Deserialize)]
pub(crate) struct GhSearchPr {
    pub number: u32,
    pub title: String,
    pub url: String,
    #[serde(rename = "createdAt")]
    pub created_at: String,
    pub repository: GhRepository,
}

#[derive(Debug, Deserialize)]
pub(crate) struct GhRepository {
    #[serde(rename = "nameWithOwner")]
    pub name_with_owner: String,
}

impl From<GhSearchPr> for PullRequest {
    fn from(pr: GhSearchPr) -> Self {
        let age_days = calculate_age_days(&pr.created_at);

        PullRequest {
            number: pr.number,
            title: pr.title,
            repo: pr.repository.name_with_owner,
            url: pr.url,
            age_days,
            review_decision: None,
        }
    }
}

pub(crate) fn calculate_age_days(created_at: &str) -> i64 {
    if let Ok(created) = DateTime::parse_from_rfc3339(created_at) {
        let now = Utc::now();
        (now - created.with_timezone(&Utc)).num_days()
    } else {
        0
    }
}

pub(crate) fn parse_pr_json(json: &[u8]) -> Vec<PullRequest> {
    let prs: Vec<GhSearchPr> = serde_json::from_slice(json).unwrap_or_default();
    prs.into_iter().map(PullRequest::from).collect()
}

pub fn fetch_prs(config: &PrsConfig) -> Result<PrsPanel> {
    if !config.enabled {
        return Ok(PrsPanel::default());
    }

    // Run both gh search commands in parallel for speed
    let review_handle = thread::spawn(fetch_review_requested);
    let authored_handle = thread::spawn(fetch_authored_prs);

    let needs_review = review_handle.join().unwrap_or_else(|_| Ok(Vec::new()))?;
    let your_prs = authored_handle.join().unwrap_or_else(|_| Ok(Vec::new()))?;

    Ok(PrsPanel {
        needs_review,
        your_prs,
    })
}

fn fetch_review_requested() -> Result<Vec<PullRequest>> {
    let output = Command::new("gh")
        .args([
            "search",
            "prs",
            "--review-requested=@me",
            "--state=open",
            "--json",
            "number,title,url,createdAt,repository",
            "--limit",
            "10",
        ])
        .stderr(Stdio::null())
        .output()?;

    if !output.status.success() {
        return Ok(Vec::new());
    }

    let mut result = parse_pr_json(&output.stdout);

    // Sort by age (oldest first - they need attention!)
    result.sort_by(|a, b| b.age_days.cmp(&a.age_days));

    Ok(result)
}

fn fetch_authored_prs() -> Result<Vec<PullRequest>> {
    let output = Command::new("gh")
        .args([
            "search",
            "prs",
            "--author=@me",
            "--state=open",
            "--json",
            "number,title,url,createdAt,repository",
            "--limit",
            "10",
        ])
        .stderr(Stdio::null())
        .output()?;

    if !output.status.success() {
        return Ok(Vec::new());
    }

    let mut result = parse_pr_json(&output.stdout);

    // Sort by age (newest first for your PRs)
    result.sort_by(|a, b| a.age_days.cmp(&b.age_days));

    Ok(result)
}

#[cfg(test)]
mod tests {
    use super::*;

    const SAMPLE_PR_JSON: &str = r#"[
        {
            "createdAt": "2025-12-10T20:20:26Z",
            "number": 1960,
            "repository": {"name": "SafeSenseTerraform", "nameWithOwner": "HaydenAI-Org/SafeSenseTerraform"},
            "title": "[CLOUDINFRA-1684] create crowdstrike state",
            "url": "https://github.com/HaydenAI-Org/SafeSenseTerraform/pull/1960"
        },
        {
            "createdAt": "2025-12-11T23:29:52Z",
            "number": 1966,
            "repository": {"name": "SafeSenseTerraform", "nameWithOwner": "HaydenAI-Org/SafeSenseTerraform"},
            "title": "Bootstrap Capability ArgoCD",
            "url": "https://github.com/HaydenAI-Org/SafeSenseTerraform/pull/1966"
        }
    ]"#;

    #[test]
    fn test_parse_pr_json_extracts_fields_correctly() {
        let prs = parse_pr_json(SAMPLE_PR_JSON.as_bytes());

        assert_eq!(prs.len(), 2);
        assert_eq!(prs[0].number, 1960);
        assert_eq!(prs[0].title, "[CLOUDINFRA-1684] create crowdstrike state");
        assert_eq!(prs[0].repo, "HaydenAI-Org/SafeSenseTerraform");
        assert_eq!(
            prs[0].url,
            "https://github.com/HaydenAI-Org/SafeSenseTerraform/pull/1960"
        );
    }

    #[test]
    fn test_parse_pr_json_handles_empty_array() {
        let prs = parse_pr_json(b"[]");
        assert!(prs.is_empty());
    }

    #[test]
    fn test_parse_pr_json_handles_invalid_json() {
        let prs = parse_pr_json(b"not valid json");
        assert!(prs.is_empty());
    }

    #[test]
    fn test_calculate_age_days_valid_date() {
        // Use a fixed recent date for testing
        let age = calculate_age_days("2025-12-01T00:00:00Z");
        // Should be positive (in the past)
        assert!(age >= 0);
    }

    #[test]
    fn test_calculate_age_days_invalid_date() {
        let age = calculate_age_days("not a date");
        assert_eq!(age, 0);
    }

    #[test]
    fn test_pull_request_from_gh_search_pr() {
        let gh_pr = GhSearchPr {
            number: 123,
            title: "Test PR".to_string(),
            url: "https://github.com/owner/repo/pull/123".to_string(),
            created_at: "2025-12-10T00:00:00Z".to_string(),
            repository: GhRepository {
                name_with_owner: "owner/repo".to_string(),
            },
        };

        let pr: PullRequest = gh_pr.into();

        assert_eq!(pr.number, 123);
        assert_eq!(pr.title, "Test PR");
        assert_eq!(pr.repo, "owner/repo");
        assert_eq!(pr.url, "https://github.com/owner/repo/pull/123");
        assert!(pr.review_decision.is_none());
    }
}
