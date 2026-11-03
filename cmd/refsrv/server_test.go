package main

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"reflect"
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
			h := srvApp.limitUnauthed(next)

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
