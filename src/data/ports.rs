use anyhow::Result;
use std::process::Command;

use super::config::PortsConfig;

#[derive(Debug, Clone, Default)]
pub struct PortsPanel {
    pub ports: Vec<PortInfo>,
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct PortInfo {
    pub port: u16,
    pub process: String,
    pub pid: u32,
}

pub fn fetch_ports(config: &PortsConfig) -> Result<PortsPanel> {
    if !config.enabled {
        return Ok(PortsPanel::default());
    }

    let output = Command::new("lsof").args(["-i", "-P", "-n"]).output()?;

    let stdout = String::from_utf8_lossy(&output.stdout);
    let mut ports: Vec<PortInfo> = stdout
        .lines()
        .filter(|line| line.contains("LISTEN"))
        .filter_map(parse_lsof_line)
        .filter(|p| !is_hidden_process(p, config))
        .filter(|p| !is_system_port(p, config))
        .filter(|p| !is_ephemeral_port(p, config))
        .collect();

    // Remove duplicates (same port can appear multiple times)
    ports.sort_by_key(|p| p.port);
    ports.dedup_by_key(|p| p.port);

    Ok(PortsPanel { ports })
}

fn parse_lsof_line(line: &str) -> Option<PortInfo> {
    let parts: Vec<&str> = line.split_whitespace().collect();
    if parts.len() < 9 {
        return None;
    }

    let process = parts[0].to_string();
    let pid = parts[1].parse().ok()?;

    // The name field contains the address:port
    // Could be in various formats: *:3000, 127.0.0.1:3000, [::1]:3000
    let name = parts.get(8)?;

    let port = extract_port(name)?;

    Some(PortInfo { port, process, pid })
}

fn extract_port(name: &str) -> Option<u16> {
    // Handle formats like:
    // *:3000 (LISTEN)
    // 127.0.0.1:3000
    // [::1]:3000

    let clean = name.trim_end_matches("(LISTEN)").trim();

    // Find the last colon and extract port after it
    if let Some(pos) = clean.rfind(':') {
        let port_str = &clean[pos + 1..];
        return port_str.parse().ok();
    }

    None
}

fn is_hidden_process(port: &PortInfo, config: &PortsConfig) -> bool {
    config
        .hidden_processes
        .iter()
        .any(|p| port.process.to_lowercase() == p.to_lowercase())
}

fn is_system_port(port: &PortInfo, config: &PortsConfig) -> bool {
    if !config.hide_system {
        return false;
    }

    // Common system ports to hide
    let system_processes = [
        "launchd",
        "systemd",
        "rapportd",
        "sharingd",
        "airplayui",
        "controlce",
        "identitys",
    ];

    system_processes
        .iter()
        .any(|p| port.process.to_lowercase().contains(p))
}

fn is_ephemeral_port(port: &PortInfo, config: &PortsConfig) -> bool {
    if !config.hide_ephemeral {
        return false;
    }

    // Ephemeral ports are typically 49152-65535 or 32768-65535 depending on OS
    port.port >= 49152
}
