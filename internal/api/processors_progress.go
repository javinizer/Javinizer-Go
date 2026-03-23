package api

import (
	"github.com/javinizer/javinizer-go/internal/logging"
	ws "github.com/javinizer/javinizer-go/internal/websocket"
)

func broadcastProgress(msg *ws.ProgressMessage) {
	if err := wsHub.BroadcastProgress(msg); err != nil {
		logging.Warnf("Failed to broadcast progress update for job %s: %v", msg.JobID, err)
	}
}
