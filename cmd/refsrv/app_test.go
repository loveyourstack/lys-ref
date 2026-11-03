package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/testutil"
	"github.com/loveyourstack/lys/lysws"
)

const (
	testUserAgent  = "lys-ref-test"
	testRemoteAddr = "192.0.2.10:5000"
)

// newTestAppNoDb returns an app with authentication enabled and no db or external connections.
// Requests that reach a store will panic, so it is only suitable for requests rejected before the handler.
func newTestAppNoDb(t testing.TB) *httpServerApplication {
	t.Helper()

	conf := &myapp.Config{}
	conf.General.Env = appenv.Dev
	conf.Api.UseAuthentication = true // disable the dev auth bypass
	conf.UI.Url = "http://localhost:3000"

	srvApp, err := newBaseApplication(conf)
	if err != nil {
		t.Fatalf("newBaseApplication() error = %v", err)
	}
	srvApp.Logger = slog.New(slog.DiscardHandler)

	return srvApp
}

// newTestApp returns an app whose stores use the test database. Cloud clients, the RPC client and services are not attached.
func newTestApp(t testing.TB) *httpServerApplication {
	t.Helper()

	conf := testutil.LoadConfig(t)
	conf.Api.UseAuthentication = true // disable the dev auth bypass

	srvApp, err := newBaseApplication(conf)
	if err != nil {
		t.Fatalf("newBaseApplication() error = %v", err)
	}
	srvApp.Logger = slog.New(slog.DiscardHandler)

	srvApp.Db = testutil.NewServerPool(t, conf)
	srvApp.attachStores()

	// the hub only listens when ListenAndBroadcast is called, which the tests don't do
	srvApp.NotificationHub, err = lysws.NewNotificationHub(context.Background(), srvApp.Db, "system.notification", 1,
		conf.UI.Url, srvApp.Logger)
	if err != nil {
		t.Fatalf("lysws.NewNotificationHub() error = %v", err)
	}

	return srvApp
}

// testServe sends a request to h and returns the recorded response. Unless the request sets them, it comes from
// the default test client address and user agent.
// A panic in a handler fails the test instead of crashing the test binary.
func testServe(t testing.TB, h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	if req.RemoteAddr == "" {
		req.RemoteAddr = testRemoteAddr
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", testUserAgent)
	}

	rec := httptest.NewRecorder()
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("%s %s: handler panicked: %v", req.Method, req.URL.Path, r)
			}
		}()
		h.ServeHTTP(rec, req)
	}()

	return rec
}
