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

func TestClientIPConsumersRejectInvalidContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "missing", ctx: context.Background()},
		{name: "invalid address", ctx: lysauth.WithClientIp(context.Background(), netip.Addr{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestAppNoDb(t)
			handlers := map[string]func(http.Handler) http.Handler{
				"logRequest":      srvApp.logRequest,
				"rejectBlockedIp": lys.RejectBlockedIp(srvApp.LoginAttempts, srvApp.Logger),
				"limitUnauthed":   lys.LimitUnauthed(srvApp.UnauthedRateLimits, srvApp.Logger),
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
	chain := srvApp.logRequest(lys.RejectBlockedIp(srvApp.LoginAttempts, srvApp.Logger)(
		lys.LimitUnauthed(srvApp.UnauthedRateLimits, srvApp.Logger)(next),
	))
	mutateRequest := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = "198.51.100.7:1234"
		r.Header.Set("X-Real-IP", "198.51.100.7")
		chain.ServeHTTP(w, r)
	})
	h := lys.SetClientIp(srvApp.UseRealIpHeader, srvApp.Logger)(mutateRequest)
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
			ctx := lysauth.WithClientIp(req.Context(), netip.MustParseAddr("192.0.2.10"))
			next := &nextRecorder{}
			rec := testServe(t, lys.RejectBlockedIp(srvApp.LoginAttempts, srvApp.Logger)(next), req.WithContext(ctx))
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
		return lysauth.WithClientIp(context.Background(), netip.MustParseAddr(ip))
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
			ctx: lysauth.WithClientIp(metadata.AppendToOutgoingContext(context.Background(), "x-other", "v"),
				netip.MustParseAddr("192.0.2.10")),
			wantCalled: true,
			wantMd:     map[string][]string{"x-real-ip": {"192.0.2.10"}, "x-other": {"v"}},
		},
		{
			name:    "client ip missing: invoker not called",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "client ip is invalid",
			ctx:     lysauth.WithClientIp(context.Background(), netip.Addr{}),
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
