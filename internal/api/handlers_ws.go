package api

import (
	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/logging"
	ws "github.com/javinizer/javinizer-go/internal/websocket"
)

// handleWebSocket handles WebSocket connections for real-time progress updates
// @Router /ws/progress [get]
// @Summary WebSocket progress updates
// @Description WebSocket endpoint for real-time progress updates during batch operations. Connect to receive streaming updates for batch scrape jobs, file organization, and downloads. Message format: JSON with job_id, type (progress/complete/error/cancelled), file, progress (0.0-1.0), message, and bytes_processed fields.
// @Success 101
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func handleWebSocket(wsHub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logging.Errorf("Failed to upgrade to websocket: %v", err)
			return
		}

		client := ws.NewClient(conn)
		wsHub.Register(client)

		// Start pumps
		go client.WritePump()
		go client.ReadPump(wsHub)
	}
}
