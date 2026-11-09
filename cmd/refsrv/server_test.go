package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/netip"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys/lysauth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestSecureHeaders(t *testing.T) {
	next := &nextRecorder{}
	rec := testServe(t, secureHeaders(next), newTestRequest(t))

	if !next.called {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	for header, want := range map[string]string{
		"X-XSS-Protection": "1; mode=block",
		"X-Frame-Options":  "deny",
	} {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestLimitAuthed(t *testing.T) {
	tests := []struct {
		name     string
		burst    int
		requests []*ReqUserInfo // one request each, in order; nil: no user info in context
		want     []int
	}{
		{
			name: "user info missing", burst: 2,
			requests: []*ReqUserInfo{nil},
			want:     []int{http.StatusUnauthorized},
		},
		{
			name: "within burst", burst: 2,
			requests: []*ReqUserInfo{{UserId: 1}, {UserId: 1}},
			want:     []int{http.StatusNoContent, http.StatusNoContent},
		},
		{
			name: "burst exceeded", burst: 2,
			requests: []*ReqUserInfo{{UserId: 1}, {UserId: 1}, {UserId: 1}},
			want:     []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests},
		},
		{
			name: "limit is per user", burst: 1,
			requests: []*ReqUserInfo{{UserId: 1}, {UserId: 2}, {UserId: 1}, {UserId: 2}},
			want:     []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests, http.StatusTooManyRequests},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestAppNoDb(t)
			srvApp.AuthedRateLimits = lysauth.NewAppRateLimits(0, tt.burst, time.Hour) // no refill: only the burst is allowed

			next := &nextRecorder{}
			h := srvApp.limitAuthed(next)

			for i, userInfo := range tt.requests {
				req := newTestRequest(t)
				if userInfo != nil {
					req = req.WithContext(context.WithValue(req.Context(), lys.UserInfoCtxKey, userInfo))
				}

				next.called = false
				rec := testServe(t, h, req)

				if rec.Code != tt.want[i] {
					t.Errorf("request %d: status = %d, want %d", i, rec.Code, tt.want[i])
				}
				if wantReached := tt.want[i] == http.StatusNoContent; next.called != wantReached {
					t.Errorf("request %d: next called = %v, want %v", i, next.called, wantReached)
				}
			}
		})
	}
}

func TestLimitUnauthed(t *testing.T) {
	type request struct {
		remoteAddr   string // default: the test client
		realIpHeader string
		want         int
	}

	tests := []struct {
		name            string
		useRealIpHeader bool
		burst           int
		requests        []request
	}{
		{
			name: "within burst", burst: 2,
			requests: []request{{want: http.StatusNoContent}, {want: http.StatusNoContent}},
		},
		{
			name: "burst exceeded", burst: 2,
			requests: []request{{want: http.StatusNoContent}, {want: http.StatusNoContent}, {want: http.StatusTooManyRequests}},
		},
		{
			name: "limit is per ip", burst: 1,
			requests: []request{
				{remoteAddr: "192.0.2.10:1000", want: http.StatusNoContent},
				{remoteAddr: "192.0.2.11:1000", want: http.StatusNoContent},
				{remoteAddr: "192.0.2.10:2000", want: http.StatusTooManyRequests}, // same ip, different port
			},
		},
		{
			name: "invalid remote addr", burst: 1,
			requests: []request{{remoteAddr: "bogus", want: http.StatusInternalServerError}},
		},
		{
			name: "real ip header: limit is per client behind the proxy", useRealIpHeader: true, burst: 1,
			requests: []request{
				{remoteAddr: "127.0.0.1:4000", realIpHeader: "198.51.100.1", want: http.StatusNoContent},
				{remoteAddr: "127.0.0.1:4000", realIpHeader: "198.51.100.2", want: http.StatusNoContent},
				{remoteAddr: "127.0.0.1:4001", realIpHeader: "198.51.100.1", want: http.StatusTooManyRequests},
			},
		},
		{
			name: "real ip header: missing header", useRealIpHeader: true, burst: 1,
			requests: []request{{remoteAddr: "127.0.0.1:4000", want: http.StatusInternalServerError}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestAppNoDb(t)
			srvApp.UseRealIpHeader = tt.useRealIpHeader
			srvApp.UnauthedRateLimits = lysauth.NewAppRateLimits(0, tt.burst, time.Hour) // no refill: only the burst is allowed

			next := &nextRecorder{}
			h := srvApp.initClientIP(srvApp.limitUnauthed(next))

			for i, r := range tt.requests {
				req := newTestRequest(t)
				req.RemoteAddr = r.remoteAddr
				if r.realIpHeader != "" {
					req.Header.Set("X-Real-IP", r.realIpHeader)
				}

				next.called = false
				rec := testServe(t, h, req)

				if rec.Code != r.want {
					t.Errorf("request %d: status = %d, want %d", i, rec.Code, r.want)
				}
				if wantReached := r.want == http.StatusNoContent; next.called != wantReached {
					t.Errorf("request %d: next called = %v, want %v", i, next.called, wantReached)
				}
			}
		})
	}
}

func TestInitClientIP(t *testing.T) {
	tests := []struct {
		name            string
		useRealIpHeader bool
		remoteAddr      string
		realIpHeader    string
		wantIP          string
		wantStatus      int
	}{
		{name: "remote address", remoteAddr: "192.0.2.10:5000", wantIP: "192.0.2.10", wantStatus: http.StatusNoContent},
		{name: "ipv6", remoteAddr: "[2001:db8::1]:5000", wantIP: "2001:db8::1", wantStatus: http.StatusNoContent},
		{name: "trusted proxy", useRealIpHeader: true, remoteAddr: "127.0.0.1:5000", realIpHeader: "198.51.100.7", wantIP: "198.51.100.7", wantStatus: http.StatusNoContent},
		{name: "invalid address", remoteAddr: "bogus", wantStatus: http.StatusInternalServerError},
		{name: "missing header", useRealIpHeader: true, remoteAddr: "127.0.0.1:5000", wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestAppNoDb(t)
			srvApp.UseRealIpHeader = tt.useRealIpHeader
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				ip, err := clientIPFromContext(r.Context())
				if err != nil || ip.String() != tt.wantIP {
					t.Errorf("clientIPFromContext() = %v, %v, want %s", ip, err, tt.wantIP)
				}
				w.WriteHeader(http.StatusNoContent)
			})
			req := newTestRequest(t)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Real-IP", tt.realIpHeader)
			rec := testServe(t, srvApp.initClientIP(next), req)
			if rec.Code != tt.wantStatus || called != (tt.wantStatus == http.StatusNoContent) {
				t.Errorf("status = %d, called = %t; want %d, %t", rec.Code, called, tt.wantStatus, tt.wantStatus == http.StatusNoContent)
			}
		})
	}
}

func TestClientIPConsumersRejectInvalidContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "missing", ctx: context.Background()},
		{name: "wrong type", ctx: context.WithValue(context.Background(), clientIpCtxKey, "192.0.2.10")},
		{name: "invalid address", ctx: context.WithValue(context.Background(), clientIpCtxKey, netip.Addr{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestAppNoDb(t)
			handlers := map[string]func(http.Handler) http.Handler{
				"logRequest":      srvApp.logRequest,
				"rejectBlockedIp": srvApp.rejectBlockedIp,
				"limitUnauthed":   srvApp.limitUnauthed,
			}
			for name, middleware := range handlers {
				t.Run(name, func(t *testing.T) {
					next := &nextRecorder{}
					rec := testServe(t, middleware(next), newTestRequest(t).WithContext(tt.ctx))
					if rec.Code != http.StatusInternalServerError || next.called {
						t.Errorf("status = %d, next called = %t; want 500, false", rec.Code, next.called)
					}
				})
			}
			t.Run("authTrackLoginAttempt", func(t *testing.T) {
				req := newTestRequest(t).WithContext(tt.ctx)
				rec := testServe(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ok := srvApp.authTrackLoginAttempt(r.Context(), w, r, netip.MustParseAddr("192.0.2.10"), "test")
					if ok {
						t.Error("authTrackLoginAttempt succeeded")
					}
				}), req)
				if rec.Code != http.StatusInternalServerError || srvApp.LoginAttempts.Count() != 0 {
					t.Errorf("status = %d, attempts = %d; want 500, 0", rec.Code, srvApp.LoginAttempts.Count())
				}
			})
		})
	}
}

func TestClientIPUsedThroughoutChain(t *testing.T) {
	srvApp := newTestAppNoDb(t)
	srvApp.UnauthedRateLimits = lysauth.NewAppRateLimits(0, 1, time.Hour)
	var logs bytes.Buffer
	srvApp.Logger = slog.New(slog.NewTextHandler(&logs, nil))
	wantIP := netip.MustParseAddr("192.0.2.10")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !srvApp.authTrackLoginAttempt(r.Context(), w, r, wantIP, "test") {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	chain := srvApp.logRequest(srvApp.rejectBlockedIp(srvApp.limitUnauthed(next)))
	mutateRequest := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = "198.51.100.7:1234"
		r.Header.Set("X-Real-IP", "198.51.100.7")
		chain.ServeHTTP(w, r)
	})
	h := srvApp.initClientIP(mutateRequest)
	for _, want := range []int{http.StatusNoContent, http.StatusTooManyRequests} {
		req := newTestRequest(t)
		req.RemoteAddr = "192.0.2.10:5000"
		if rec := testServe(t, h, req); rec.Code != want {
			t.Errorf("status = %d, want %d", rec.Code, want)
		}
	}
	attempts := srvApp.LoginAttempts.All()
	if len(attempts) != 1 || attempts[0].Ip != wantIP {
		t.Errorf("login attempts = %v, want one for %v", attempts, wantIP)
	}
	if !strings.Contains(logs.String(), wantIP.String()+" - ") {
		t.Errorf("logs do not contain client IP %v: %s", wantIP, logs.String())
	}
}

func TestRejectBlockedIpUsesContext(t *testing.T) {
	tests := []struct {
		name       string
		blockedIP  string
		wantStatus int
	}{
		{name: "context IP blocked", blockedIP: "192.0.2.10", wantStatus: http.StatusForbidden},
		{name: "request IP blocked", blockedIP: "198.51.100.7", wantStatus: http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestAppNoDb(t)
			if err := srvApp.LoginAttempts.Block(netip.MustParseAddr(tt.blockedIP)); err != nil {
				t.Fatal(err)
			}
			req := newTestRequest(t)
			req.RemoteAddr = "198.51.100.7:1234"
			ctx := context.WithValue(req.Context(), clientIpCtxKey, netip.MustParseAddr("192.0.2.10"))
			next := &nextRecorder{}
			rec := testServe(t, srvApp.rejectBlockedIp(next), req.WithContext(ctx))
			if rec.Code != tt.wantStatus || next.called != (tt.wantStatus == http.StatusNoContent) {
				t.Errorf("status = %d, next called = %t; want %d", rec.Code, next.called, tt.wantStatus)
			}
		})
	}
}

// fakeInvoker is a grpc.UnaryInvoker that records the context it was called with.
type fakeInvoker struct {
	called bool
	ctx    context.Context
	err    error
}

func (f *fakeInvoker) invoke(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	f.called = true
	f.ctx = ctx
	return f.err
}

func TestRpcDefaultDeadlineUnaryInterceptor(t *testing.T) {
	const timeout = time.Minute

	callerDeadline := time.Now().Add(time.Hour)
	withCallerDeadline, cancel := context.WithDeadline(context.Background(), callerDeadline)
	defer cancel()

	invokerErr := errors.New("invoker failed")

	tests := []struct {
		name         string
		ctx          context.Context
		invokerErr   error
		wantDeadline func(got time.Time) bool
	}{
		{
			name: "no deadline: default applied",
			ctx:  context.Background(),
			wantDeadline: func(got time.Time) bool {
				return time.Until(got) > timeout-10*time.Second && time.Until(got) <= timeout
			},
		},
		{
			name:         "caller deadline kept",
			ctx:          withCallerDeadline,
			wantDeadline: func(got time.Time) bool { return got.Equal(callerDeadline) },
		},
		{
			name:         "invoker error is returned",
			ctx:          context.Background(),
			invokerErr:   invokerErr,
			wantDeadline: func(got time.Time) bool { return !got.IsZero() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &fakeInvoker{err: tt.invokerErr}

			err := rpcDefaultDeadlineUnaryInterceptor(timeout)(tt.ctx, "/svc/Method", nil, nil, nil, inv.invoke)

			if !errors.Is(err, tt.invokerErr) {
				t.Errorf("error = %v, want %v", err, tt.invokerErr)
			}
			if !inv.called {
				t.Fatal("invoker was not called")
			}
			deadline, ok := inv.ctx.Deadline()
			if !ok {
				t.Fatal("invoker context has no deadline")
			}
			if !tt.wantDeadline(deadline) {
				t.Errorf("invoker context deadline = %v, in %v", deadline, time.Until(deadline))
			}
		})
	}
}

func TestRpcForwardIpUnaryInterceptor(t *testing.T) {
	ipCtx := func(ip string) context.Context {
		return context.WithValue(context.Background(), clientIpCtxKey, netip.MustParseAddr(ip))
	}

	tests := []struct {
		name       string
		ctx        context.Context
		invokerErr error
		wantCalled bool
		wantErr    bool
		wantMd     map[string][]string // outgoing metadata seen by the invoker
	}{
		{
			name:       "ipv4",
			ctx:        ipCtx("192.0.2.10"),
			wantCalled: true,
			wantMd:     map[string][]string{"x-real-ip": {"192.0.2.10"}},
		},
		{
			name:       "ipv6",
			ctx:        ipCtx("2001:db8::1"),
			wantCalled: true,
			wantMd:     map[string][]string{"x-real-ip": {"2001:db8::1"}},
		},
		{
			name: "existing metadata kept",
			ctx: context.WithValue(metadata.AppendToOutgoingContext(context.Background(), "x-other", "v"),
				clientIpCtxKey, netip.MustParseAddr("192.0.2.10")),
			wantCalled: true,
			wantMd:     map[string][]string{"x-real-ip": {"192.0.2.10"}, "x-other": {"v"}},
		},
		{
			name:    "client ip missing: invoker not called",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "client ip has wrong type",
			ctx:     context.WithValue(context.Background(), clientIpCtxKey, "192.0.2.10"),
			wantErr: true,
		},
		{
			name:    "client ip is invalid",
			ctx:     context.WithValue(context.Background(), clientIpCtxKey, netip.Addr{}),
			wantErr: true,
		},
		{
			name:       "invoker error is returned",
			ctx:        ipCtx("192.0.2.10"),
			invokerErr: errors.New("invoker failed"),
			wantCalled: true,
			wantErr:    true,
			wantMd:     map[string][]string{"x-real-ip": {"192.0.2.10"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &fakeInvoker{err: tt.invokerErr}

			err := rpcForwardIpUnaryInterceptor()(tt.ctx, "/svc/Method", nil, nil, nil, inv.invoke)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.invokerErr != nil && !errors.Is(err, tt.invokerErr) {
				t.Errorf("error = %v, want %v", err, tt.invokerErr)
			}
			if inv.called != tt.wantCalled {
				t.Fatalf("invoker called = %v, want %v", inv.called, tt.wantCalled)
			}
			if !tt.wantCalled {
				return
			}
			md, _ := metadata.FromOutgoingContext(inv.ctx)
			if !reflect.DeepEqual(map[string][]string(md), tt.wantMd) {
				t.Errorf("outgoing metadata = %v, want %v", md, tt.wantMd)
			}
		})
	}
}

func TestRpcKeyCredentials(t *testing.T) {
	creds := rpcKeyCredentials{key: "secret-key"}

	got, err := creds.GetRequestMetadata(context.Background(), "https://example.com/svc")
	if err != nil {
		t.Fatalf("GetRequestMetadata() error = %v", err)
	}
	want := map[string]string{"x-api-key": "secret-key"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetRequestMetadata() = %v, want %v", got, want)
	}

	if creds.RequireTransportSecurity() {
		t.Error("RequireTransportSecurity() = true, want false: the RPC server is only reached over loopback")
	}
}
