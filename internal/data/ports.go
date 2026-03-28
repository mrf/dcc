package data

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/mrf/dcc/internal/config"
)

// PortInfo represents a listening port
type PortInfo struct {
	Port    uint16
	Process string
	PID     uint32
}

// PortsPanel holds port data for display
type PortsPanel struct {
	Ports     []PortInfo
	IsLoading bool
}

var systemProcesses = map[string]bool{
	"launchd":           true,
	"systemd":           true,
	"sharingd":          true,
	"SystemUIServer":    true,
	"loginwindow":       true,
	"UserEventAgent":    true,
	"mDNSResponder":     true,
	"configd":           true,
	"airportd":          true,
	"CommCenter":        true,
	"identityservicesd": true,
	"cloudd":            true,
}

// FetchPorts retrieves listening ports using lsof
func FetchPorts(cfg config.PortsConfig) PortsPanel {
	if !cfg.Enabled {
		return PortsPanel{}
	}

	cmd := exec.Command("lsof", "-i", "-P", "-n")
	output, err := cmd.Output()
	if err != nil {
		return PortsPanel{}
	}

	ports := parseLsofOutput(string(output), cfg)

	// Sort by port number
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	// Deduplicate by port
	seen := make(map[uint16]bool)
	unique := make([]PortInfo, 0, len(ports))
	for _, p := range ports {
		if !seen[p.Port] {
			seen[p.Port] = true
			unique = append(unique, p)
		}
	}

	return PortsPanel{Ports: unique}
}

func parseLsofOutput(output string, cfg config.PortsConfig) []PortInfo {
	var ports []PortInfo
	lines := strings.Split(output, "\n")

	// Build hidden processes map
	hiddenProcesses := make(map[string]bool)
	for _, p := range cfg.HiddenProcesses {
		hiddenProcesses[p] = true
	}

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Only process LISTEN lines
		if !strings.Contains(line, "LISTEN") {
			continue
		}

		port := parseLsofLine(line, cfg, hiddenProcesses)
		if port != nil {
			ports = append(ports, *port)
		}
	}

	return ports
}

func parseLsofLine(line string, cfg config.PortsConfig, hiddenProcesses map[string]bool) *PortInfo {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return nil
	}

	process := fields[0]
	pidStr := fields[1]
	address := fields[8]

	// Filter hidden processes
	if hiddenProcesses[process] {
		return nil
	}

	// Filter system processes
	if cfg.HideSystem && systemProcesses[process] {
		return nil
	}

	// Parse PID
	pid, err := strconv.ParseUint(pidStr, 10, 32)
	if err != nil {
		return nil
	}

	// Parse port from address (format: *:3000, 127.0.0.1:3000, [::1]:3000)
	port := extractPort(address)
	if port == 0 {
		return nil
	}

	// Filter ephemeral ports (49152-65535)
	if cfg.HideEphemeral && port >= 49152 {
		return nil
	}

	return &PortInfo{
		Port:    port,
		Process: process,
		PID:     uint32(pid),
	}
}

func extractPort(address string) uint16 {
	// Find the last colon (handles IPv6)
	lastColon := strings.LastIndex(address, ":")
	if lastColon == -1 {
		return 0
	}

	portStr := address[lastColon+1:]

	// Remove any suffix like "(LISTEN)"
	if idx := strings.Index(portStr, "("); idx != -1 {
		portStr = portStr[:idx]
	}

	// Handle format like "3000->3000"
	if idx := strings.Index(portStr, "-"); idx != -1 {
		portStr = portStr[:idx]
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0
	}

	return uint16(port)
}

// OpenPort opens localhost at the given port index in browser
func OpenPort(panel PortsPanel, idx int) error {
	if idx < 0 || idx >= len(panel.Ports) || idx >= 10 {
		return nil
	}
	url := fmt.Sprintf("http://localhost:%d", panel.Ports[idx].Port)
	return exec.Command("open", url).Run()
}
