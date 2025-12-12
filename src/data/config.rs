use anyhow::Result;
use serde::Deserialize;
use std::path::PathBuf;

#[derive(Debug, Clone, Deserialize, Default)]
pub struct Config {
    #[serde(default)]
    pub general: GeneralConfig,
    #[serde(default)]
    pub meetings: MeetingsConfig,
    #[serde(default)]
    pub prs: PrsConfig,
    #[serde(default)]
    pub ports: PortsConfig,
    #[serde(default)]
    pub git: GitConfig,
}

#[derive(Debug, Clone, Deserialize)]
pub struct GeneralConfig {
    #[serde(default = "default_refresh_interval")]
    pub refresh_interval_seconds: u64,
    #[serde(default = "default_projects_dir")]
    pub projects_dir: String,
}

fn default_refresh_interval() -> u64 {
    30
}

fn default_projects_dir() -> String {
    dirs::home_dir()
        .map(|h| h.join("Projects").display().to_string())
        .unwrap_or_else(|| "~/Projects".to_string())
}

impl Default for GeneralConfig {
    fn default() -> Self {
        Self {
            refresh_interval_seconds: default_refresh_interval(),
            projects_dir: default_projects_dir(),
        }
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct MeetingsConfig {
    #[serde(default = "default_true")]
    pub enabled: bool,
    #[serde(default = "default_hours_ahead")]
    pub hours_ahead: u32,
    #[serde(default)]
    pub calendars_exclude: Vec<String>,
    #[serde(default)]
    pub ignore_patterns: Vec<String>,
}

fn default_true() -> bool {
    true
}

fn default_hours_ahead() -> u32 {
    8
}

impl Default for MeetingsConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            hours_ahead: 8,
            calendars_exclude: vec![
                "Birthdays".to_string(),
                "US Holidays".to_string(),
                "Siri Suggestions".to_string(),
            ],
            ignore_patterns: vec![
                "Focus Time".to_string(),
                "Lunch".to_string(),
                "OOO".to_string(),
            ],
        }
    }
}

#[derive(Debug, Clone, Deserialize)]
#[allow(dead_code)]
pub struct PrsConfig {
    #[serde(default = "default_true")]
    pub enabled: bool,
    #[serde(default)]
    pub repos: Vec<String>,
}

impl Default for PrsConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            repos: Vec::new(),
        }
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct PortsConfig {
    #[serde(default = "default_true")]
    pub enabled: bool,
    #[serde(default = "default_true")]
    pub hide_system: bool,
    #[serde(default = "default_true")]
    pub hide_ephemeral: bool,
    #[serde(default)]
    pub hidden_processes: Vec<String>,
}

impl Default for PortsConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            hide_system: true,
            hide_ephemeral: true,
            hidden_processes: vec![
                "rapportd".to_string(),
                "ControlCenter".to_string(),
                "mDNSResponder".to_string(),
            ],
        }
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct GitConfig {
    #[serde(default = "default_true")]
    pub enabled: bool,
    #[serde(default = "default_scan_depth")]
    pub scan_depth: usize,
    #[serde(default)]
    pub ignore_dirs: Vec<String>,
}

fn default_scan_depth() -> usize {
    2
}

impl Default for GitConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            scan_depth: 2,
            ignore_dirs: vec![
                "node_modules".to_string(),
                ".git".to_string(),
                "target".to_string(),
                "vendor".to_string(),
            ],
        }
    }
}

impl Config {
    pub fn load() -> Result<Self> {
        let config_path = Self::config_path();

        if config_path.exists() {
            let content = std::fs::read_to_string(&config_path)?;
            let config: Config = toml::from_str(&content)?;
            Ok(config)
        } else {
            Ok(Config::default())
        }
    }

    pub fn config_path() -> PathBuf {
        dirs::config_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join("dcc")
            .join("config.toml")
    }

    pub fn projects_path(&self) -> PathBuf {
        let path = shellexpand(&self.general.projects_dir);
        PathBuf::from(path)
    }
}

fn shellexpand(path: &str) -> String {
    if let Some(stripped) = path.strip_prefix("~/") {
        if let Some(home) = dirs::home_dir() {
            return home.join(stripped).display().to_string();
        }
    }
    path.to_string()
}
