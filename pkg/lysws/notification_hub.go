package lysws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const MaxUserConnections int = 5

// NotificationHub manages WebSocket connections for user notifications.
// It listens for database notifications and broadcasts messages to connected clients based on user ID.
type NotificationHub struct {
	closed          atomic.Bool
	conns           map[int64][]*websocket.Conn // user_id → active sockets
	db              *pgxpool.Pool
	dbListenConn    *pgxpool.Conn // single connection acquired from pool for LISTEN/UNLISTEN
	dbListenChannel string        // database channel to LISTEN on for notifications
	errorLog        *slog.Logger
	listenMu        sync.Mutex         // protects listenCancel and listenDone
	listenCancel    context.CancelFunc // used to signal ListenAndBroadcast to stop
	listenDone      chan struct{}      // closed when ListenAndBroadcast has fully stopped
	mu              sync.RWMutex       // protects conns
}

// NewNotificationHub creates a new NotificationHub instance.
// It acquires a database connection for listening to notifications and initializes the connection map.
func NewNotificationHub(ctx context.Context, db *pgxpool.Pool, dbListenChannel string, errorLog *slog.Logger) (hub *NotificationHub, err error) {

	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	if dbListenChannel == "" {
		return nil, fmt.Errorf("dbListenChannel is required")
	}
	if errorLog == nil {
		return nil, fmt.Errorf("errorLog is required")
	}

	// acquire a single connection from the pool for listening
	dbConn, err := db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("db.Acquire failed: %w", err)
	}

	return &NotificationHub{
		db:              db,
		dbListenConn:    dbConn,
		dbListenChannel: dbListenChannel,
		conns:           make(map[int64][]*websocket.Conn),
		errorLog:        errorLog,
	}, nil
}

func (h *NotificationHub) broadcast(userID int64, msg []byte, logFailures bool) (err error) {

	// copy user conns to avoid iteration issues due to Unregister calls
	h.mu.RLock()
	conns := h.conns[userID]
	connsCopy := make([]*websocket.Conn, len(conns))
	copy(connsCopy, conns)
	h.mu.RUnlock()

	for _, conn := range connsCopy {
		if connErr := conn.WriteMessage(websocket.TextMessage, msg); connErr != nil {
			if logFailures {
				h.errorLog.Error("conn.WriteMessage failed", "user_id", userID, "error", connErr)
			}
			h.Unregister(userID, conn)
			err = errors.Join(err, connErr)
		}
	}
	return err
}

// Broadcast sends a message to all active WebSocket connections for a given user ID.
func (h *NotificationHub) Broadcast(userID int64, msg []byte) {
	_ = h.broadcast(userID, msg, true)
}

// BroadcastE sends a message to all active WebSocket connections for a given user ID.
// It returns an error if any connection fails to send the message, but continues to attempt sending to all connections.
func (h *NotificationHub) BroadcastE(userID int64, msg []byte) (err error) {
	return h.broadcast(userID, msg, false)
}

// Close closes all active WebSocket connections and clears the connection map.
func (h *NotificationHub) Close() (err error) {

	// exit if already closed, and set to closed if not already
	if !h.closed.CompareAndSwap(false, true) {
		return nil
	}

	// snapshot all connections and clear conns map while holding the lock
	h.mu.Lock()
	all := make([]*websocket.Conn, 0)
	for _, conns := range h.conns {
		all = append(all, conns...)
	}
	h.conns = make(map[int64][]*websocket.Conn)
	h.mu.Unlock()

	// snapshot listen-related fields while holding lock
	h.listenMu.Lock()
	lisCancel := h.listenCancel
	lisDone := h.listenDone
	lisConn := h.dbListenConn
	h.listenMu.Unlock()

	// signal ListenAndBroadcast to stop and wait for it to finish
	if lisCancel != nil {
		lisCancel()
	}
	if lisDone != nil {
		<-lisDone
	}

	// close the listen connection if it exists
	if lisConn != nil {
		lisConn.Release()
		h.listenMu.Lock()
		if h.dbListenConn == lisConn {
			h.dbListenConn = nil
		}
		h.listenMu.Unlock()
	}

	// close sockets outside the lock so slow network closes do not block remaining ops
	for _, conn := range all {
		if closeErr := conn.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}

	return err
}

type NotificationSelectFunc func(ctx context.Context, db *pgxpool.Pool, notId int64) (userId int64, notType, message string, err error)

// ListenAndBroadcast listens for database notifications and broadcasts messages to users based on the notification payload.
func (h *NotificationHub) ListenAndBroadcast(ctx context.Context, selectFunc NotificationSelectFunc) (err error) {

	if selectFunc == nil {
		return fmt.Errorf("selectFunc is required")
	}

	h.listenMu.Lock()
	if h.closed.Load() {
		h.listenMu.Unlock()
		return fmt.Errorf("notification hub is closed")
	}

	if h.listenDone != nil {
		h.listenMu.Unlock()
		return fmt.Errorf("listener already running")
	}

	// snapshot listen conn while holding lock
	lisConn := h.dbListenConn
	if lisConn == nil {
		h.listenMu.Unlock()
		return fmt.Errorf("dbListenConn is not initialized")
	}

	// create listen context and done channel for signaling shutdown
	lisCtx, cancel := context.WithCancel(ctx)

	// ensure resources are cleaned up if we exit early with an error
	defer cancel()

	// set listen-related fields before starting listener loop
	done := make(chan struct{})
	h.listenCancel = cancel
	h.listenDone = done
	h.listenMu.Unlock()

	// ensure we UNLISTEN and clean up listen-related fields when ListenAndBroadcast exits
	defer func() {
		_, unlistenErr := lisConn.Exec(context.Background(), "UNLISTEN "+pgx.Identifier{h.dbListenChannel}.Sanitize())
		if unlistenErr != nil {
			h.errorLog.Error("UNLISTEN failed", "channel", h.dbListenChannel, "error", unlistenErr)
		}

		h.listenMu.Lock()
		close(done)
		h.listenCancel = nil
		h.listenDone = nil
		h.listenMu.Unlock()
	}()

	// LISTEN to receive notifications on the dbListenChannel
	_, err = lisConn.Exec(lisCtx, "LISTEN "+pgx.Identifier{h.dbListenChannel}.Sanitize())
	if err != nil {
		return fmt.Errorf("lisConn.Exec (LISTEN) failed on channel %s: %w", h.dbListenChannel, err)
	}

	type messagePayload struct {
		Type string `json:"type"`
		Body string `json:"body"`
	}

	// wait for notifications or context cancellation
	for {
		not, err := lisConn.Conn().WaitForNotification(lisCtx)
		if err != nil {
			if lisCtx.Err() != nil {
				return nil
			}
			return fmt.Errorf("lisConn.Conn().WaitForNotification failed: %w", err)
		}

		// payload needs to be the notification ID int64 to be looked up by selectFunc
		notId, err := strconv.ParseInt(not.Payload, 10, 64)
		if err != nil {
			h.errorLog.Error("strconv.ParseInt failed", "payload", not.Payload, "error", err)
			continue
		}

		// select the notification details
		userId, notType, message, err := selectFunc(ctx, h.db, notId)
		if err != nil {
			h.errorLog.Error("selectFunc failed", "notification_id", notId, "error", err)
			continue
		}

		// prepare JSON payload
		payload := messagePayload{
			Type: notType,
			Body: message,
		}
		msgBytes, err := json.Marshal(payload)
		if err != nil {
			h.errorLog.Error("json.Marshal failed", "payload", payload, "error", err)
			continue
		}

		// broadcast the message to the user's active connections
		if err := h.BroadcastE(userId, msgBytes); err != nil {
			h.errorLog.Error("h.BroadcastE failed", "user_id", userId, "error", err)
		}

	} // end for
}

// Register adds a WebSocket connection for a given user ID.
func (h *NotificationHub) Register(userID int64, c *websocket.Conn) error {

	if c == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	var shouldClose bool
	var err error

	h.mu.Lock()

	switch {

	// reject if hub is closed
	case h.closed.Load():
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
func (h *NotificationHub) Unregister(userID int64, c *websocket.Conn) {
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

// UserConnCount returns the number of active connections for a given user ID.
func (h *NotificationHub) UserConnCount(userID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if conns, exists := h.conns[userID]; exists {
		return len(conns)
	}
	return 0
}
