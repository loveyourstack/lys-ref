package lysws

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
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

func TestRegisterRejectsNilConnection(t *testing.T) {
	hub := NewNotificationHub(testLogger())

	err := hub.Register(1, nil)
	if err == nil {
		t.Fatal("expected error for nil connection")
	}
	if !strings.Contains(err.Error(), "connection cannot be nil") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterAndUnregister(t *testing.T) {
	hub := NewNotificationHub(testLogger())
	userID := int64(99)

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	if err := hub.Register(userID, serverConn); err != nil {
		t.Fatalf("register connection: %v", err)
	}
	if got := hub.UserConnCount(userID); got != 1 {
		t.Fatalf("expected 1 registered connection, got %d", got)
	}

	hub.Unregister(userID, serverConn)
	if got := hub.UserConnCount(userID); got != 0 {
		t.Fatalf("expected 0 connections after unregister, got %d", got)
	}

	if hub.UserExists(userID) {
		t.Fatal("expected user entry to be removed after last connection unregister")
	}
}

func TestRegisterRejectsWhenHubClosed(t *testing.T) {
	hub := NewNotificationHub(testLogger())

	if err := hub.Close(); err != nil {
		t.Fatalf("close hub: %v", err)
	}

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	err := hub.Register(7, serverConn)
	if err == nil {
		t.Fatal("expected error when registering on closed hub")
	}
	if !strings.Contains(err.Error(), "closed") {
		t.Fatalf("unexpected error: %v", err)
	}

	if writeErr := serverConn.WriteMessage(websocket.TextMessage, []byte("x")); writeErr == nil {
		t.Fatal("expected registered connection to be closed when hub is closed")
	}
}

func TestRegisterEnforcesMaxUserConnections(t *testing.T) {
	hub := NewNotificationHub(testLogger())
	userID := int64(123)

	cleanups := make([]func(), 0, MaxUserConnections+1)
	t.Cleanup(func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	})

	for i := 0; i < MaxUserConnections; i++ {
		serverConn, _, cleanup := newWebsocketPair(t)
		cleanups = append(cleanups, cleanup)

		if err := hub.Register(userID, serverConn); err != nil {
			t.Fatalf("register connection %d: %v", i, err)
		}
	}

	extraConn, _, cleanup := newWebsocketPair(t)
	cleanups = append(cleanups, cleanup)

	err := hub.Register(userID, extraConn)
	if err == nil {
		t.Fatal("expected max connections error")
	}
	if !strings.Contains(err.Error(), "maximum active connections") {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := hub.UserConnCount(userID); got != MaxUserConnections {
		t.Fatalf("expected %d connections after max enforcement, got %d", MaxUserConnections, got)
	}

	if writeErr := extraConn.WriteMessage(websocket.TextMessage, []byte("x")); writeErr == nil {
		t.Fatal("expected extra connection to be closed by Register")
	}
}

func TestBroadcastEUnregistersFailedConnectionAndReturnsError(t *testing.T) {
	hub := NewNotificationHub(testLogger())
	userID := int64(42)

	goodServerConn, goodClientConn, cleanupGood := newWebsocketPair(t)
	t.Cleanup(cleanupGood)

	badServerConn, _, cleanupBad := newWebsocketPair(t)
	t.Cleanup(cleanupBad)

	if err := hub.Register(userID, goodServerConn); err != nil {
		t.Fatalf("register good connection: %v", err)
	}
	if err := hub.Register(userID, badServerConn); err != nil {
		t.Fatalf("register bad connection: %v", err)
	}

	// Force write failure for one registered connection.
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
	hub := NewNotificationHub(testLogger())
	userID := int64(77)

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	if err := hub.Register(userID, serverConn); err != nil {
		t.Fatalf("register connection: %v", err)
	}

	if err := hub.Close(); err != nil {
		t.Fatalf("close hub: %v", err)
	}

	hub.mu.RLock()
	closed := hub.closed
	connCount := len(hub.conns)
	hub.mu.RUnlock()

	if !closed {
		t.Fatal("expected hub to be marked closed")
	}
	if connCount != 0 {
		t.Fatalf("expected connection map to be cleared; got %d user entries", connCount)
	}
}
