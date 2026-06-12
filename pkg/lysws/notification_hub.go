package lysws

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
)

const MaxUserConnections int = 5

// notificationHub manages WebSocket connections for user notifications.
type notificationHub struct {
	closed   bool
	conns    map[int64][]*websocket.Conn // user_id → active sockets
	errorLog *slog.Logger
	mu       sync.RWMutex
}

// NewNotificationHub creates a new notificationHub instance.
func NewNotificationHub(errorLog *slog.Logger) *notificationHub {

	if errorLog == nil {
		log.Fatalln("notificationHub: errorLog is required")
	}

	return &notificationHub{
		conns:    make(map[int64][]*websocket.Conn),
		errorLog: errorLog,
	}
}

func (h *notificationHub) broadcast(userID int64, msg []byte, logFailures bool) (err error) {

	// copy user conns to avoid iteration issues due to Unregister calls
	h.mu.RLock()
	conns := h.conns[userID]
	connsCopy := make([]*websocket.Conn, len(conns))
	copy(connsCopy, conns)
	h.mu.RUnlock()

	for _, conn := range connsCopy {
		if connErr := conn.WriteMessage(websocket.TextMessage, msg); connErr != nil {
			if logFailures {
				h.errorLog.Error("Failed to send notification", "user_id", userID, "error", connErr)
			}
			h.Unregister(userID, conn)
			err = errors.Join(err, connErr)
		}
	}
	return err
}

// Broadcast sends a message to all active WebSocket connections for a given user ID.
func (h *notificationHub) Broadcast(userID int64, msg []byte) {
	_ = h.broadcast(userID, msg, true)
}

// BroadcastE sends a message to all active WebSocket connections for a given user ID.
// It returns an error if any connection fails to send the message, but continues to attempt sending to all connections.
func (h *notificationHub) BroadcastE(userID int64, msg []byte) (err error) {
	return h.broadcast(userID, msg, false)
}

// Close closes all active WebSocket connections and clears the connection map.
func (h *notificationHub) Close() (err error) {

	h.mu.Lock()

	// mark hub as closed to prevent new registrations
	h.closed = true

	// snapshot all connections and clear conns map while holding the lock
	all := make([]*websocket.Conn, 0)
	for _, conns := range h.conns {
		all = append(all, conns...)
	}
	h.conns = make(map[int64][]*websocket.Conn)
	h.mu.Unlock()

	// close sockets outside the lock so slow network closes do not block remaining ops
	for _, conn := range all {
		if closeErr := conn.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}

	return err
}

// Register adds a WebSocket connection for a given user ID.
func (h *notificationHub) Register(userID int64, c *websocket.Conn) error {

	if c == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	var shouldClose bool
	var err error

	h.mu.Lock()

	switch {

	// reject if hub is closed
	case h.closed:
		shouldClose = true
		err = fmt.Errorf("notification hub is closed")

	// reject if c is already registered for userID (shouldn't happen but just in case)
	case slices.Contains(h.conns[userID], c):
		err = fmt.Errorf("connection already registered for user %d", userID)

	// reject if userID already has maximum active connections
	case len(h.conns[userID]) >= MaxUserConnections:
		shouldClose = true
		err = fmt.Errorf("maximum active connections reached for user %d", userID)

	// register new connection
	default:
		h.conns[userID] = append(h.conns[userID], c)
	}

	h.mu.Unlock()

	if shouldClose {
		_ = c.Close()
	}

	return err
}

// Unregister removes a WebSocket connection for a given user ID.
func (h *notificationHub) Unregister(userID int64, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.conns[userID]
	for i, conn := range conns {
		if conn == c {
			h.conns[userID] = append(conns[:i], conns[i+1:]...)
			if len(h.conns[userID]) == 0 {
				delete(h.conns, userID)
			}
			break
		}
	}
}

func (h *notificationHub) UserConnCount(userID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if conns, exists := h.conns[userID]; exists {
		return len(conns)
	}
	return 0
}

func (h *notificationHub) UserExists(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.conns[userID]
	return exists
}
