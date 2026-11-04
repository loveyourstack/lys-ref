package main

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"
)

const lifecycleTestTimeout = 15 * time.Second

// serveResult is the outcome of a serve call running in the background.
type serveResult struct {
	ctx    context.Context
	cancel context.CancelFunc
	addr   string
	done   chan error // receives the error returned by serve
}

// startServe runs srvApp.serve on a random loopback port and waits until it answers requests.
// The server is stopped when the test ends if it hasn't been already.
func startServe(t *testing.T, srvApp *httpServerApplication) *serveResult {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	res := &serveResult{ctx: ctx, cancel: cancel, addr: ln.Addr().String(), done: make(chan error, 1)}
	srv := &http.Server{Handler: srvApp.getRouter()}
	go func() { res.done <- srvApp.serve(ctx, cancel, srv, ln) }()

	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(lifecycleTestTimeout)
	for {
		select {
		case err := <-res.done:
			t.Fatalf("serve returned before serving: %v", err)
		default:
		}

		resp, err := client.Get("http://" + res.addr + "/")
		if err == nil {
			resp.Body.Close()
			return res
		}
		if time.Now().After(deadline) {
			t.Fatalf("server not ready after %v: %v", lifecycleTestTimeout, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// wait returns the error returned by serve, failing the test if it doesn't return in time.
func (res *serveResult) wait(t *testing.T) error {
	t.Helper()

	select {
	case err := <-res.done:
		return err
	case <-time.After(lifecycleTestTimeout):
		t.Fatalf("serve did not return within %v", lifecycleTestTimeout)
		return nil
	}
}

// resetPersistedState empties the tables that serve loads from and persists to, now and when the test ends.
func (env *authTestEnv) resetPersistedState(t *testing.T) {
	t.Helper()

	reset := func() {
		for _, stmt := range []string{"DELETE FROM system.session", "DELETE FROM system.login_attempt"} {
			if _, err := env.ownerDb.Exec(context.Background(), stmt); err != nil {
				t.Errorf("%s: %v", stmt, err)
			}
		}
	}
	reset()
	t.Cleanup(reset) // registered after newUser's cleanup, so runs before the user is deleted
}

func TestServeShutdownPersistsState(t *testing.T) {
	env := newAuthTestEnv(t)
	srvApp := env.srvApp
	u := env.newUser(t, false, false)
	env.resetPersistedState(t)

	res := startServe(t, srvApp)

	// state created while serving
	token := addTestSession(t, srvApp, u.id, u.name, []string{"Viewer"})
	attemptIp := netip.MustParseAddr("192.0.2.77")
	if err := srvApp.LoginAttempts.Add(attemptIp); err != nil {
		t.Fatalf("LoginAttempts.Add() error = %v", err)
	}

	res.cancel()
	if err := res.wait(t); err != nil {
		t.Fatalf("serve() error = %v, want nil after the context was canceled", err)
	}

	// the listener is closed
	if resp, err := (&http.Client{Timeout: time.Second}).Get("http://" + res.addr + "/"); err == nil {
		resp.Body.Close()
		t.Error("server still answers requests after serve returned")
	}

	// the state was written to the db
	var n int
	err := env.ownerDb.QueryRow(context.Background(), "SELECT count(*) FROM system.session WHERE token = $1 AND user_fk = $2", token, u.id).Scan(&n)
	if err != nil || n != 1 {
		t.Errorf("persisted sessions with the token = %d (error %v), want 1", n, err)
	}
	err = env.ownerDb.QueryRow(context.Background(), "SELECT count(*) FROM system.login_attempt WHERE ip = $1::inet", attemptIp.String()).Scan(&n)
	if err != nil || n != 1 {
		t.Errorf("persisted login attempts for %s = %d (error %v), want 1", attemptIp, n, err)
	}

	// a restarted app gets it back
	restarted := newTestApp(t)
	if err := restarted.loadState(context.Background()); err != nil {
		t.Fatalf("loadState() error = %v", err)
	}
	sess, err := restarted.Sessions.FromToken(token, netip.MustParseAddr(testRemoteIp()), testUserAgent, false, restarted.Logger)
	if err != nil {
		t.Fatalf("restored session: Sessions.FromToken() error = %v", err)
	}
	if sess.UserId != u.id || sess.UserName != u.name {
		t.Errorf("restored session user = %d %q, want %d %q", sess.UserId, sess.UserName, u.id, u.name)
	}
	// other login attempts can be loaded too: system.blocked_ip is loaded as login attempts and may contain seed data
	restoredAttempt := false
	for _, la := range restarted.LoginAttempts.All() {
		if la.Ip == attemptIp {
			restoredAttempt = true
		}
	}
	if !restoredAttempt {
		t.Errorf("login attempt for %s was not restored", attemptIp)
	}
}

func TestServeLoadsState(t *testing.T) {
	env := newAuthTestEnv(t)
	u := env.newUser(t, false, false)
	env.resetPersistedState(t)

	// first run: create a session and shut down, which persists it
	first := startServe(t, env.srvApp)
	token := addTestSession(t, env.srvApp, u.id, u.name, []string{"Viewer"})
	first.cancel()
	if err := first.wait(t); err != nil {
		t.Fatalf("first serve() error = %v", err)
	}

	// second run: serve loads the session before it starts serving
	second := newTestApp(t)
	res := startServe(t, second)

	if _, err := second.Sessions.FromToken(token, netip.MustParseAddr(testRemoteIp()), testUserAgent, false, second.Logger); err != nil {
		t.Errorf("restored session: Sessions.FromToken() error = %v", err)
	}

	res.cancel()
	if err := res.wait(t); err != nil {
		t.Fatalf("second serve() error = %v", err)
	}
}

func TestServeReturnsServerError(t *testing.T) {
	env := newAuthTestEnv(t)
	env.resetPersistedState(t)
	srvApp := env.srvApp

	// a closed listener makes srv.Serve fail immediately
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- srvApp.serve(ctx, cancel, &http.Server{Handler: srvApp.getRouter()}, ln) }()

	select {
	case err = <-done:
	case <-time.After(lifecycleTestTimeout):
		t.Fatalf("serve did not return within %v", lifecycleTestTimeout)
	}

	if err == nil || !strings.Contains(err.Error(), "srv.Serve failed") {
		t.Errorf("serve() error = %v, want it to contain %q", err, "srv.Serve failed")
	}
	if ctx.Err() == nil {
		t.Error("stop was not called: background goroutines would keep running")
	}
}

func TestServeLoadStateFails(t *testing.T) {
	srvApp := newTestApp(t)
	srvApp.Db.Close() // every query now fails

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = srvApp.serve(ctx, cancel, &http.Server{Handler: srvApp.getRouter()}, ln)
	if err == nil || !strings.Contains(err.Error(), "loadSessionsFromDb") {
		t.Errorf("serve() error = %v, want it to contain %q", err, "loadSessionsFromDb")
	}

	// the listener passed to serve is closed even when serve fails early
	if conn, err := net.DialTimeout("tcp", ln.Addr().String(), time.Second); err == nil {
		conn.Close()
		t.Error("listener is still open after serve failed")
	}
}

func TestShutdown(t *testing.T) {
	const generous = 10 * time.Second
	const tiny = 20 * time.Millisecond

	tests := []struct {
		name         string
		timeouts     shutdownTimeouts
		inFlight     bool // an http request that never completes is in progress
		hubExits     bool
		wantContains []string
		wantNil      bool
	}{
		{
			name:     "all steps succeed",
			timeouts: shutdownTimeouts{Http: generous, Persist: generous, Hub: generous},
			hubExits: true,
			wantNil:  true,
		},
		{
			name:         "http connections do not drain",
			timeouts:     shutdownTimeouts{Http: tiny, Persist: generous, Hub: generous},
			inFlight:     true,
			hubExits:     true,
			wantContains: []string{"shutdown: srv.Shutdown failed"},
		},
		{
			name:         "persisting times out",
			timeouts:     shutdownTimeouts{Http: generous, Persist: time.Nanosecond, Hub: generous},
			hubExits:     true,
			wantContains: []string{"shutdown: srvApp.persistSessions failed", "shutdown: srvApp.persistLoginAttempts failed"},
		},
		{
			name:         "hub listener does not exit",
			timeouts:     shutdownTimeouts{Http: generous, Persist: generous, Hub: tiny},
			hubExits:     false,
			wantContains: []string{"shutdown: timeout waiting for hub listener to exit"},
		},
		{
			name:     "every step runs even if earlier ones fail",
			timeouts: shutdownTimeouts{Http: tiny, Persist: time.Nanosecond, Hub: tiny},
			inFlight: true,
			hubExits: false,
			wantContains: []string{
				"shutdown: srv.Shutdown failed",
				"shutdown: srvApp.persistSessions failed",
				"shutdown: srvApp.persistLoginAttempts failed",
				"shutdown: timeout waiting for hub listener to exit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestApp(t)
			srvApp.ShutdownTimeouts = tt.timeouts

			// server whose /block handler holds the request until the test ends
			release := make(chan struct{})
			entered := make(chan struct{})
			var enterOnce sync.Once
			mux := http.NewServeMux()
			mux.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
				enterOnce.Do(func() { close(entered) })
				<-release
			})
			srv := &http.Server{Handler: mux}

			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("net.Listen() error = %v", err)
			}
			serveDone := make(chan struct{})
			go func() { defer close(serveDone); srv.Serve(ln) }()

			var clientDone chan struct{}
			t.Cleanup(func() {
				close(release)
				srv.Close()
				<-serveDone
				if clientDone != nil {
					<-clientDone
				}
			})

			if tt.inFlight {
				clientDone = make(chan struct{})
				go func() {
					defer close(clientDone)
					if resp, err := http.Get("http://" + ln.Addr().String() + "/block"); err == nil {
						resp.Body.Close()
					}
				}()
				select {
				case <-entered:
				case <-time.After(lifecycleTestTimeout):
					t.Fatal("in-flight request did not reach the handler")
				}
			}

			bg := &background{hubDone: make(chan struct{})}
			if tt.hubExits {
				close(bg.hubDone)
			}

			err = srvApp.shutdown(srv, bg)

			if tt.wantNil {
				if err != nil {
					t.Fatalf("shutdown() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("shutdown() error = nil, want it to contain %q", tt.wantContains)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("shutdown() error = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}
