package lysws

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testMaxUserConnections = 5

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestHub() *NotificationHub {
	return &NotificationHub{
		conns:              make(map[int64][]*websocket.Conn),
		errorLog:           testLogger(),
		infoLog:            testLogger(),
		maxUserConnections: testMaxUserConnections,
	}
}

func newWebsocketPair(t *testing.T) (*websocket.Conn, *websocket.Conn, func()) {
	t.Helper()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}

	serverConnCh := make(chan *websocket.Conn, 1)
	errCh := make(chan error, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			errCh <- err
			return
		}
		serverConnCh <- conn
	}))

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		ts.Close()
		t.Fatalf("dial websocket: %v", err)
	}

	var serverConn *websocket.Conn
	select {
	case serverConn = <-serverConnCh:
	case err = <-errCh:
		_ = clientConn.Close()
		ts.Close()
		t.Fatalf("upgrade websocket: %v", err)
	case <-time.After(2 * time.Second):
		_ = clientConn.Close()
		ts.Close()
		t.Fatal("timed out waiting for server websocket connection")
	}

	cleanup := func() {
		_ = serverConn.Close()
		_ = clientConn.Close()
		ts.Close()
	}

	return serverConn, clientConn, cleanup
}

func TestNewNotificationHubValidation(t *testing.T) {
	tests := []struct {
		name               string
		db                 *pgxpool.Pool
		channel            string
		maxUserConnections int
		infoLog            *slog.Logger
		errorLog           *slog.Logger
		wantErr            string
	}{
		{
			name:               "nil db",
			db:                 nil,
			channel:            "chan_notifications",
			maxUserConnections: testMaxUserConnections,
			infoLog:            testLogger(),
			errorLog:           testLogger(),
			wantErr:            "db is required",
		},
		{
			name:               "empty channel",
			db:                 &pgxpool.Pool{},
			channel:            "",
			maxUserConnections: testMaxUserConnections,
			infoLog:            testLogger(),
			errorLog:           testLogger(),
			wantErr:            "dbListenChannel is required",
		},
		{
			name:               "maxUserConnections zero",
			db:                 &pgxpool.Pool{},
			channel:            "chan_notifications",
			maxUserConnections: 0,
			infoLog:            testLogger(),
			errorLog:           testLogger(),
			wantErr:            "maxUserConnections must be greater than 0",
		},
		{
			name:               "maxUserConnections negative",
			db:                 &pgxpool.Pool{},
			channel:            "chan_notifications",
			maxUserConnections: -1,
			infoLog:            testLogger(),
			errorLog:           testLogger(),
			wantErr:            "maxUserConnections must be greater than 0",
		},
		{
			name:               "nil infoLog",
			db:                 &pgxpool.Pool{},
			channel:            "chan_notifications",
			maxUserConnections: testMaxUserConnections,
			infoLog:            nil,
			errorLog:           testLogger(),
			wantErr:            "infoLog is required",
		},
		{
			name:               "nil errorLog",
			db:                 &pgxpool.Pool{},
			channel:            "chan_notifications",
			maxUserConnections: testMaxUserConnections,
			infoLog:            testLogger(),
			errorLog:           nil,
			wantErr:            "errorLog is required",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			hub, err := NewNotificationHub(context.Background(), tc.db, tc.channel, tc.maxUserConnections, "*", tc.infoLog, tc.errorLog)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if hub != nil {
				t.Fatal("expected nil hub on constructor validation failure")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error: got %q, want contains %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestListenAndBroadcastValidation(t *testing.T) {
	dummySelect := func(ctx context.Context, db *pgxpool.Pool, notID int64) (int64, string, string, error) {
		return 1, "x", "y", nil
	}

	t.Run("nil select func", func(t *testing.T) {
		hub := newTestHub()
		err := hub.ListenAndBroadcast(context.Background(), nil)
		if err == nil || !strings.Contains(err.Error(), "selectFunc is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("closed hub", func(t *testing.T) {
		hub := newTestHub()
		hub.closed.Store(true)

		err := hub.ListenAndBroadcast(context.Background(), dummySelect)
		if err == nil || !strings.Contains(err.Error(), "notification hub is closed") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing listen conn", func(t *testing.T) {
		hub := newTestHub()
		hub.dbListenChannel = "chan_notifications"

		err := hub.ListenAndBroadcast(context.Background(), dummySelect)
		if err == nil || !strings.Contains(err.Error(), "dbListenConn is not initialized") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRegisterNilConnectionIsIgnored(t *testing.T) {
	hub := newTestHub()

	hub.Register(1, nil)

	if got := hub.UserConnCount(1); got != 0 {
		t.Fatalf("expected nil connection to be ignored, got %d", got)
	}
}

func TestRegisterAndUnregister(t *testing.T) {
	hub := newTestHub()
	userID := int64(99)

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	hub.Register(userID, serverConn)
	if got := hub.UserConnCount(userID); got != 1 {
		t.Fatalf("expected 1 registered connection, got %d", got)
	}

	hub.Unregister(userID, serverConn)
	if got := hub.UserConnCount(userID); got != 0 {
		t.Fatalf("expected 0 connections after unregister, got %d", got)
	}

	hub.mu.RLock()
	_, exists := hub.conns[userID]
	hub.mu.RUnlock()
	if exists {
		t.Fatal("expected user entry to be removed after last connection unregister")
	}
}

func TestRegisterRejectsWhenHubClosed(t *testing.T) {
	hub := newTestHub()
	hub.closed.Store(true)

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	hub.Register(7, serverConn)

	if got := hub.UserConnCount(7); got != 0 {
		t.Fatalf("expected no registration when hub is closed, got %d", got)
	}

	if writeErr := serverConn.WriteMessage(websocket.TextMessage, []byte("x")); writeErr == nil {
		t.Fatal("expected rejected connection to be closed when hub is closed")
	}
}

func TestRegisterEnforcesMaxUserConnections(t *testing.T) {
	hub := newTestHub()
	userID := int64(123)

	cleanups := make([]func(), 0, testMaxUserConnections+1)
	t.Cleanup(func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	})

	for i := 0; i < testMaxUserConnections; i++ {
		serverConn, _, cleanup := newWebsocketPair(t)
		cleanups = append(cleanups, cleanup)

		hub.Register(userID, serverConn)
	}

	extraConn, extraClientConn, cleanup := newWebsocketPair(t)
	cleanups = append(cleanups, cleanup)

	hub.Register(userID, extraConn)

	if got := hub.UserConnCount(userID); got != testMaxUserConnections {
		t.Fatalf("expected %d connections after max enforcement, got %d", testMaxUserConnections, got)
	}

	if err := extraClientConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	_, _, readErr := extraClientConn.ReadMessage()
	if readErr == nil {
		t.Fatal("expected close error from extra rejected connection")
	}

	closeErr, ok := readErr.(*websocket.CloseError)
	if !ok {
		t.Fatalf("expected websocket close error, got %T (%v)", readErr, readErr)
	}

	if closeErr.Code != 4429 {
		t.Fatalf("expected close code 4429, got %d", closeErr.Code)
	}
	if closeErr.Text != "max connections reached" {
		t.Fatalf("expected close text %q, got %q", "max connections reached", closeErr.Text)
	}
}

func TestBroadcastEUnregistersFailedConnectionAndReturnsError(t *testing.T) {
	hub := newTestHub()
	userID := int64(42)

	goodServerConn, goodClientConn, cleanupGood := newWebsocketPair(t)
	t.Cleanup(cleanupGood)

	badServerConn, _, cleanupBad := newWebsocketPair(t)
	t.Cleanup(cleanupBad)

	hub.Register(userID, goodServerConn)
	hub.Register(userID, badServerConn)

	if err := badServerConn.Close(); err != nil {
		t.Fatalf("close bad connection: %v", err)
	}

	msg := []byte("hello")
	err := hub.BroadcastE(userID, msg)
	if err == nil {
		t.Fatal("expected broadcast error when one connection is closed")
	}

	if got := hub.UserConnCount(userID); got != 1 {
		t.Fatalf("expected failed connection to be unregistered; got %d connections", got)
	}

	if err := goodClientConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	msgType, gotMsg, readErr := goodClientConn.ReadMessage()
	if readErr != nil {
		t.Fatalf("read broadcast message from good connection: %v", readErr)
	}
	if msgType != websocket.TextMessage {
		t.Fatalf("expected text message type, got %d", msgType)
	}
	if string(gotMsg) != string(msg) {
		t.Fatalf("unexpected message body: got %q, want %q", string(gotMsg), string(msg))
	}
}

func TestCloseMarksHubClosedAndClearsConnections(t *testing.T) {
	hub := newTestHub()
	userID := int64(77)

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	hub.Register(userID, serverConn)

	if err := hub.Close(); err != nil {
		t.Fatalf("close hub: %v", err)
	}

	if !hub.closed.Load() {
		t.Fatal("expected hub to be marked closed")
	}

	hub.mu.RLock()
	connCount := len(hub.conns)
	hub.mu.RUnlock()
	if connCount != 0 {
		t.Fatalf("expected connection map to be cleared; got %d user entries", connCount)
	}
}
