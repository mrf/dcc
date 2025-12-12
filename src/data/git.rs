use anyhow::Result;
use std::path::{Path, PathBuf};
use std::process::Command;

use super::config::GitConfig;

#[derive(Debug, Clone, Default)]
pub struct GitPanel {
    pub dirty_repos: Vec<DirtyRepo>,
    pub stashes: Vec<StashInfo>,
}

#[derive(Debug, Clone)]
pub struct DirtyRepo {
    pub name: String,
    pub path: String,
    pub modified: usize,
    pub untracked: usize,
    pub staged: usize,
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct StashInfo {
    pub repo: String,
    pub index: usize,
    pub message: String,
    pub age_days: i64,
}

pub fn fetch_git_status(projects_dir: &Path, config: &GitConfig) -> Result<GitPanel> {
    if !config.enabled {
        return Ok(GitPanel::default());
    }

    let mut dirty_repos = Vec::new();
    let mut stashes = Vec::new();

    let repos = find_git_repos(projects_dir, config.scan_depth, &config.ignore_dirs)?;

    for repo_path in repos {
        // Check for uncommitted changes
        if let Ok(output) = Command::new("git")
            .args(["status", "--porcelain"])
            .current_dir(&repo_path)
            .output()
        {
            if output.status.success() {
                let stdout = String::from_utf8_lossy(&output.stdout);
                let changes = parse_git_status(&stdout);

                if changes.has_changes() {
                    let name = repo_path
                        .file_name()
                        .map(|n| n.to_string_lossy().to_string())
                        .unwrap_or_else(|| "unknown".to_string());

                    dirty_repos.push(DirtyRepo {
                        name,
                        path: repo_path.display().to_string(),
                        modified: changes.modified,
                        untracked: changes.untracked,
                        staged: changes.staged,
                    });
                }
            }
        }

        // Check for stashes
        if let Ok(output) = Command::new("git")
            .args(["stash", "list", "--format=%gd|||%s|||%ar"])
            .current_dir(&repo_path)
            .output()
        {
            if output.status.success() {
                let stdout = String::from_utf8_lossy(&output.stdout);
                let repo_name = repo_path
                    .file_name()
                    .map(|n| n.to_string_lossy().to_string())
                    .unwrap_or_else(|| "unknown".to_string());

                for stash in parse_stash_list(&stdout, &repo_name) {
                    stashes.push(stash);
                }
            }
        }
    }

    // Sort dirty repos by name
    dirty_repos.sort_by(|a, b| a.name.cmp(&b.name));

    Ok(GitPanel {
        dirty_repos,
        stashes,
    })
}

fn find_git_repos(dir: &Path, max_depth: usize, ignore_dirs: &[String]) -> Result<Vec<PathBuf>> {
    let mut repos = Vec::new();
    find_git_repos_recursive(dir, max_depth, 0, ignore_dirs, &mut repos)?;
    Ok(repos)
}

fn find_git_repos_recursive(
    dir: &Path,
    max_depth: usize,
    current_depth: usize,
    ignore_dirs: &[String],
    repos: &mut Vec<PathBuf>,
) -> Result<()> {
    if current_depth > max_depth {
        return Ok(());
    }

    if !dir.is_dir() {
        return Ok(());
    }

    // Check if this directory is a git repo
    let git_dir = dir.join(".git");
    if git_dir.exists() {
        repos.push(dir.to_path_buf());
        return Ok(()); // Don't recurse into git repos
    }

    // Recurse into subdirectories
    if let Ok(entries) = std::fs::read_dir(dir) {
        for entry in entries.flatten() {
            let path = entry.path();

            if !path.is_dir() {
                continue;
            }

            // Skip ignored directories
            if let Some(name) = path.file_name() {
                let name_str = name.to_string_lossy();
                if ignore_dirs.iter().any(|d| d == name_str.as_ref()) {
                    continue;
                }
                // Also skip hidden directories
                if name_str.starts_with('.') {
                    continue;
                }
            }

            find_git_repos_recursive(&path, max_depth, current_depth + 1, ignore_dirs, repos)?;
        }
    }

    Ok(())
}

#[derive(Default)]
struct GitChanges {
    modified: usize,
    untracked: usize,
    staged: usize,
}

impl GitChanges {
    fn has_changes(&self) -> bool {
        self.modified > 0 || self.untracked > 0 || self.staged > 0
    }
}

fn parse_git_status(output: &str) -> GitChanges {
    let mut changes = GitChanges::default();

    for line in output.lines() {
        if line.len() < 2 {
            continue;
        }

        let index_status = line.chars().next().unwrap_or(' ');
        let worktree_status = line.chars().nth(1).unwrap_or(' ');

        // Staged changes (index has changes)
        if index_status != ' ' && index_status != '?' {
            changes.staged += 1;
        }

        // Modified in worktree
        if worktree_status == 'M' || worktree_status == 'D' {
            changes.modified += 1;
        }

        // Untracked files
        if index_status == '?' {
            changes.untracked += 1;
        }
    }

    changes
}

fn parse_stash_list(output: &str, repo_name: &str) -> Vec<StashInfo> {
    let mut stashes = Vec::new();

    for line in output.lines() {
        if line.trim().is_empty() {
            continue;
        }

        let parts: Vec<&str> = line.split("|||").collect();
        if parts.len() < 3 {
            continue;
        }

        // Parse stash@{0} to get index
        let stash_ref = parts[0].trim();
        let index = stash_ref
            .split('{')
            .nth(1)
            .and_then(|s| s.trim_end_matches('}').parse().ok())
            .unwrap_or(0);

        let message = parts[1].trim().to_string();
        let age_str = parts[2].trim();
        let age_days = parse_relative_time(age_str);

        stashes.push(StashInfo {
            repo: repo_name.to_string(),
            index,
            message,
            age_days,
        });
    }

    stashes
}

fn parse_relative_time(time_str: &str) -> i64 {
    // Parse strings like "3 days ago", "2 weeks ago", "1 month ago"
    let parts: Vec<&str> = time_str.split_whitespace().collect();
    if parts.len() < 2 {
        return 0;
    }

    let num: i64 = parts[0].parse().unwrap_or(0);
    let unit = parts[1].to_lowercase();

    match unit.as_str() {
        "second" | "seconds" => 0,
        "minute" | "minutes" => 0,
        "hour" | "hours" => 0,
        "day" | "days" => num,
        "week" | "weeks" => num * 7,
        "month" | "months" => num * 30,
        "year" | "years" => num * 365,
        _ => 0,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_git_status_modified_files() {
        let output = " M src/main.rs\n M src/lib.rs\n";
        let changes = parse_git_status(output);

        assert_eq!(changes.modified, 2);
        assert_eq!(changes.staged, 0);
        assert_eq!(changes.untracked, 0);
    }

    #[test]
    fn test_parse_git_status_staged_files() {
        let output = "M  src/main.rs\nA  src/new.rs\n";
        let changes = parse_git_status(output);

        assert_eq!(changes.staged, 2);
        assert_eq!(changes.modified, 0);
        assert_eq!(changes.untracked, 0);
    }

    #[test]
    fn test_parse_git_status_untracked_files() {
        let output = "?? newfile.txt\n?? another.txt\n";
        let changes = parse_git_status(output);

        assert_eq!(changes.untracked, 2);
        assert_eq!(changes.modified, 0);
        assert_eq!(changes.staged, 0);
    }

    #[test]
    fn test_parse_git_status_mixed() {
        let output = " M modified.rs\nM  staged.rs\n?? untracked.txt\n";
        let changes = parse_git_status(output);

        assert_eq!(changes.modified, 1);
        assert_eq!(changes.staged, 1);
        assert_eq!(changes.untracked, 1);
    }

    #[test]
    fn test_parse_git_status_empty() {
        let changes = parse_git_status("");
        assert!(!changes.has_changes());
    }

    #[test]
    fn test_parse_stash_list() {
        let output = "stash@{0}|||WIP on main: abc123 some commit|||3 days ago\n\
                      stash@{1}|||feature work|||2 weeks ago\n";
        let stashes = parse_stash_list(output, "test-repo");

        assert_eq!(stashes.len(), 2);
        assert_eq!(stashes[0].repo, "test-repo");
        assert_eq!(stashes[0].index, 0);
        assert_eq!(stashes[0].message, "WIP on main: abc123 some commit");
        assert_eq!(stashes[0].age_days, 3);

        assert_eq!(stashes[1].index, 1);
        assert_eq!(stashes[1].message, "feature work");
        assert_eq!(stashes[1].age_days, 14); // 2 weeks
    }

    #[test]
    fn test_parse_stash_list_empty() {
        let stashes = parse_stash_list("", "test-repo");
        assert!(stashes.is_empty());
    }

    #[test]
    fn test_parse_relative_time_days() {
        assert_eq!(parse_relative_time("3 days ago"), 3);
        assert_eq!(parse_relative_time("1 day ago"), 1);
    }

    #[test]
    fn test_parse_relative_time_weeks() {
        assert_eq!(parse_relative_time("2 weeks ago"), 14);
        assert_eq!(parse_relative_time("1 week ago"), 7);
    }

    #[test]
    fn test_parse_relative_time_months() {
        assert_eq!(parse_relative_time("1 month ago"), 30);
        assert_eq!(parse_relative_time("3 months ago"), 90);
    }

    #[test]
    fn test_parse_relative_time_hours_minutes_seconds_are_zero() {
        // These should all return 0 since they're less than a day
        assert_eq!(parse_relative_time("5 hours ago"), 0);
        assert_eq!(parse_relative_time("30 minutes ago"), 0);
        assert_eq!(parse_relative_time("10 seconds ago"), 0);
    }

    #[test]
    fn test_parse_relative_time_invalid() {
        assert_eq!(parse_relative_time("invalid"), 0);
        assert_eq!(parse_relative_time(""), 0);
    }

    #[test]
    fn test_git_changes_has_changes() {
        let empty = GitChanges::default();
        assert!(!empty.has_changes());

        let with_modified = GitChanges {
            modified: 1,
            ..Default::default()
        };
        assert!(with_modified.has_changes());

        let with_untracked = GitChanges {
            untracked: 1,
            ..Default::default()
        };
        assert!(with_untracked.has_changes());

        let with_staged = GitChanges {
            staged: 1,
            ..Default::default()
        };
        assert!(with_staged.has_changes());
    }
}
