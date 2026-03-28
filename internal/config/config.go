package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config represents the main configuration structure
type Config struct {
	General  GeneralConfig  `toml:"general"`
	Meetings MeetingsConfig `toml:"meetings"`
	Prs      PrsConfig      `toml:"prs"`
	Ports    PortsConfig    `toml:"ports"`
	Git      GitConfig      `toml:"git"`
}

// GeneralConfig holds general application settings
type GeneralConfig struct {
	RefreshIntervalSeconds int    `toml:"refresh_interval_seconds"`
	ProjectsDir            string `toml:"projects_dir"`
}

// MeetingsConfig holds meetings panel settings
type MeetingsConfig struct {
	Enabled          bool     `toml:"enabled"`
	HoursAhead       int      `toml:"hours_ahead"`
	CalendarsExclude []string `toml:"calendars_exclude"`
	IgnorePatterns   []string `toml:"ignore_patterns"`
}

// PrsConfig holds pull requests panel settings
type PrsConfig struct {
	Enabled bool     `toml:"enabled"`
	Repos   []string `toml:"repos"`
}

// PortsConfig holds ports panel settings
type PortsConfig struct {
	Enabled         bool     `toml:"enabled"`
	HideSystem      bool     `toml:"hide_system"`
	HideEphemeral   bool     `toml:"hide_ephemeral"`
	HiddenProcesses []string `toml:"hidden_processes"`
}

// GitConfig holds git panel settings
type GitConfig struct {
	Enabled    bool     `toml:"enabled"`
	ScanDepth  int      `toml:"scan_depth"`
	IgnoreDirs []string `toml:"ignore_dirs"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	projectsDir := filepath.Join(homeDir, "Projects")

	return Config{
		General: GeneralConfig{
			RefreshIntervalSeconds: 30,
			ProjectsDir:            projectsDir,
		},
		Meetings: MeetingsConfig{
			Enabled:          true,
			HoursAhead:       8,
			CalendarsExclude: []string{"Birthdays", "US Holidays", "Siri Suggestions"},
			IgnorePatterns:   []string{"Focus Time", "Lunch", "OOO"},
		},
		Prs: PrsConfig{
			Enabled: true,
			Repos:   []string{},
		},
		Ports: PortsConfig{
			Enabled:         true,
			HideSystem:      true,
			HideEphemeral:   true,
			HiddenProcesses: []string{"rapportd", "ControlCenter", "mDNSResponder"},
		},
		Git: GitConfig{
			Enabled:    true,
			ScanDepth:  2,
			IgnoreDirs: []string{"node_modules", ".git", "target", "vendor"},
		},
	}
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[1:])
		}
	}
	return path
}

// Path returns the resolved path to the config file.
func Path() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "dcc", "config.toml")
}

// Load reads the config file at Path() and merges it with defaults.
// Returns defaults if the file doesn't exist or the home directory is unresolvable.
func Load() (Config, error) {
	cfg := DefaultConfig()

	configPath := Path()
	if configPath == "" {
		return cfg, nil // Return defaults
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return defaults if config doesn't exist
		}
		return cfg, err
	}

	// Parse TOML, merging with defaults
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return cfg, err
	}

	// Expand ~ in paths
	cfg.General.ProjectsDir = expandPath(cfg.General.ProjectsDir)

	// Ensure reasonable defaults for unset values
	if cfg.General.RefreshIntervalSeconds <= 0 {
		cfg.General.RefreshIntervalSeconds = 30
	}
	if cfg.Meetings.HoursAhead <= 0 {
		cfg.Meetings.HoursAhead = 8
	}
	if cfg.Git.ScanDepth <= 0 {
		cfg.Git.ScanDepth = 2
	}

	return cfg, nil
}
