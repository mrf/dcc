package app

import (
	"time"

	"github.com/mrf/dcc/internal/data"
)

// DataFetchedMsg is sent when data fetching completes
type DataFetchedMsg struct {
	Meetings data.MeetingsPanel
	Prs      data.PrsPanel
	Ports    data.PortsPanel
	Git      data.GitPanel
}

// TickMsg is sent periodically for auto-refresh
type TickMsg time.Time

// ClearNotificationMsg clears the notification bar after a timeout
type ClearNotificationMsg struct{}
