package main

import (
	"context"
	"net/http"
	"net/netip"
	"reflect"
	"testing"
	"time"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssession"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lystype"
)

// withUserInfo mimics logRequest: it puts an empty *ReqUserInfo in the request context for the middleware under test to fill in.
func withUserInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), lys.UserInfoCtxKey, &ReqUserInfo{})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// nextRecorder is a final handler that records whether it was reached and the user info it saw.
type nextRecorder struct {
	called   bool
	userInfo *ReqUserInfo
}

func (n *nextRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.called = true
	n.userInfo, _ = r.Context().Value(lys.UserInfoCtxKey).(*ReqUserInfo)
	w.WriteHeader(http.StatusNoContent)
}

func newTestRequest(t testing.TB) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	return req
}

// addTestSession adds a session for the default test client and returns its token.
func addTestSession(t testing.TB, srvApp *httpServerApplication, userId int64, userName string, roles []string) string {
	t.Helper()

	token, err := srvApp.Sessions.Add(lysauth.SessionInput{
		Email:               userName + "@example.com",
		FamilyName:          "Test",
		GeoIpCountryIsoCode: "DE",
		GeoIpLocation:       "Berlin, Germany",
		GivenName:           "Test",
		Ip:                  netip.MustParseAddr(testRemoteIp()),
		Roles:               roles,
		UserAgent:           testUserAgent,
		UserId:              userId,
		UserName:            userName,
	})
	if err != nil {
		t.Fatalf("Sessions.Add() error = %v", err)
	}
	return token
}

func TestAppSessionsToInputs(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	expires := created.Add(10 * time.Hour)
	lastAccess := created.Add(time.Hour)
	ip := netip.MustParseAddr("192.0.2.10")

	full := lysauth.Session{
		CreatedAt:    lystype.Datetime(created),
		ExpiresAt:    lystype.Datetime(expires),
		LastAccessAt: lystype.Datetime(lastAccess),
		Token:        "token-1",
		SessionInput: lysauth.SessionInput{
			AllowMultipleSessions: true,
			Email:                 "a@example.com",
			FamilyName:            "Family",
			ForcePasswordChange:   true,
			GeoIpCountryIsoCode:   "DE",
			GeoIpLocation:         "Berlin, Germany",
			GivenName:             "Given",
			Ip:                    ip,
			ProfilePic:            "pic.png",
			Roles:                 []string{"Tech", "Viewer"},
			UserAgent:             "agent",
			UserId:                7,
			UserName:              "alice",
		},
	}
	fullWant := syssession.Input{
		AllowMultipleSessions: true,
		CreatedAt:             lystype.Datetime(created),
		Email:                 "a@example.com",
		ExpiresAt:             lystype.Datetime(expires),
		FamilyName:            "Family",
		ForcePasswordChange:   true,
		GivenName:             "Given",
		GeoIpCountryIsoCode:   "DE",
		GeoIpLocation:         "Berlin, Germany",
		Ip:                    ip,
		LastAccessAt:          lystype.Datetime(lastAccess),
		ProfilePic:            "pic.png",
		Roles:                 []string{"Tech", "Viewer"},
		Token:                 "token-1",
		UserAgent:             "agent",
		UserFk:                7,
		UserName:              "alice",
	}

	second := full
	second.Token = "token-2"
	second.UserId = 8
	second.ForcePasswordChange = false
	secondWant := fullWant
	secondWant.Token = "token-2"
	secondWant.UserFk = 8
	secondWant.ForcePasswordChange = false

	tests := []struct {
		name string
		in   []lysauth.Session
		want []syssession.Input
	}{
		{name: "nil", in: nil, want: []syssession.Input{}},
		{name: "one session, all fields", in: []lysauth.Session{full}, want: []syssession.Input{fullWant}},
		{name: "order is kept", in: []lysauth.Session{second, full}, want: []syssession.Input{secondWant, fullWant}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appSessionsToInputs(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("appSessionsToInputs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAppLoginAttemptsToInputs(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	ip1 := netip.MustParseAddr("192.0.2.10")
	ip2 := netip.MustParseAddr("2001:db8::1")

	tests := []struct {
		name string
		in   []lysauth.LoginAttempt
		want []sysloginattempt.Input
	}{
		{name: "nil", in: nil, want: []sysloginattempt.Input{}},
		{
			name: "all fields, order kept",
			in: []lysauth.LoginAttempt{
				{CreatedAt: lystype.Datetime(created), Ip: ip1, IsBlocked: true, NumAttempts: 5},
				{CreatedAt: lystype.Datetime(created.Add(time.Minute)), Ip: ip2, NumAttempts: 2},
			},
			want: []sysloginattempt.Input{
				{CreatedAt: lystype.Datetime(created), Ip: ip1, IsBlocked: true, NumAttempts: 5},
				{CreatedAt: lystype.Datetime(created.Add(time.Minute)), Ip: ip2, NumAttempts: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appLoginAttemptsToInputs(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("appLoginAttemptsToInputs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAuthorizeRole(t *testing.T) {
	tests := []struct {
		name        string
		userInfo    *ReqUserInfo // nil: not in context
		allowed     []sysrole.Enum
		wantStatus  int
		wantReached bool
	}{
		{name: "user info missing", userInfo: nil, allowed: []sysrole.Enum{sysrole.Viewer}, wantStatus: http.StatusUnauthorized},
		{name: "user info missing, no roles required", userInfo: nil, allowed: nil, wantStatus: http.StatusUnauthorized},
		{name: "no roles required", userInfo: &ReqUserInfo{}, allowed: nil, wantStatus: http.StatusNoContent, wantReached: true},
		{name: "matching role", userInfo: &ReqUserInfo{Roles: []sysrole.Enum{sysrole.Standard}}, allowed: []sysrole.Enum{sysrole.Standard, sysrole.Tech}, wantStatus: http.StatusNoContent, wantReached: true},
		{name: "one of several roles matches", userInfo: &ReqUserInfo{Roles: []sysrole.Enum{sysrole.Viewer, sysrole.Tech}}, allowed: []sysrole.Enum{sysrole.Tech}, wantStatus: http.StatusNoContent, wantReached: true},
		{name: "no matching role", userInfo: &ReqUserInfo{Roles: []sysrole.Enum{sysrole.Viewer}}, allowed: []sysrole.Enum{sysrole.Standard, sysrole.Tech}, wantStatus: http.StatusForbidden},
		{name: "user has no roles", userInfo: &ReqUserInfo{}, allowed: []sysrole.Enum{sysrole.Viewer}, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := &nextRecorder{}
			h := authorizeRole(tt.allowed)(next)

			req := newTestRequest(t)
			if tt.userInfo != nil {
				req = req.WithContext(context.WithValue(req.Context(), lys.UserInfoCtxKey, tt.userInfo))
			}

			rec := testServe(t, h, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if next.called != tt.wantReached {
				t.Errorf("next called = %v, want %v", next.called, tt.wantReached)
			}
		})
	}
}

func TestRejectBlockedIp(t *testing.T) {
	blockedIp := netip.MustParseAddr(testRemoteIp())
	realIp := netip.MustParseAddr("198.51.100.7")

	tests := []struct {
		name            string
		useRealIpHeader bool
		blocked         []netip.Addr
		remoteAddr      string // default: the test client
		realIpHeader    string
		wantStatus      int
	}{
		{name: "not blocked", wantStatus: http.StatusNoContent},
		{name: "blocked", blocked: []netip.Addr{blockedIp}, wantStatus: http.StatusForbidden},
		{name: "another ip is blocked", blocked: []netip.Addr{realIp}, wantStatus: http.StatusNoContent},
		{name: "invalid remote addr", remoteAddr: "bogus", wantStatus: http.StatusInternalServerError},
		{
			name: "real ip header: blocked client behind proxy", useRealIpHeader: true, blocked: []netip.Addr{realIp},
			remoteAddr: "127.0.0.1:4000", realIpHeader: realIp.String(), wantStatus: http.StatusForbidden,
		},
		{
			name: "real ip header: proxy ip is not what is checked", useRealIpHeader: true, blocked: []netip.Addr{netip.MustParseAddr("127.0.0.1")},
			remoteAddr: "127.0.0.1:4000", realIpHeader: realIp.String(), wantStatus: http.StatusNoContent,
		},
		{
			name: "real ip header: header missing", useRealIpHeader: true,
			remoteAddr: "127.0.0.1:4000", wantStatus: http.StatusInternalServerError,
		},
		{
			name: "real ip header: remote addr is not loopback", useRealIpHeader: true,
			realIpHeader: realIp.String(), wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestAppNoDb(t)
			srvApp.UseRealIpHeader = tt.useRealIpHeader
			for _, ip := range tt.blocked {
				if err := srvApp.LoginAttempts.Block(ip); err != nil {
					t.Fatalf("LoginAttempts.Block() error = %v", err)
				}
			}

			next := &nextRecorder{}
			h := srvApp.rejectBlockedIp(next)

			req := newTestRequest(t)
			req.RemoteAddr = tt.remoteAddr
			if tt.realIpHeader != "" {
				req.Header.Set("X-Real-IP", tt.realIpHeader)
			}

			rec := testServe(t, h, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if wantReached := tt.wantStatus == http.StatusNoContent; next.called != wantReached {
				t.Errorf("next called = %v, want %v", next.called, wantReached)
			}
		})
	}
}

func TestAuthenticateDevBypass(t *testing.T) {
	tests := []struct {
		name     string
		env      appenv.Enum
		useAuth  bool
		withInfo bool // whether logRequest has put user info in the context
		header   string
		want     int
		wantUser *ReqUserInfo
	}{
		{
			name: "dev without authentication: dummy user", env: appenv.Dev, useAuth: false, withInfo: true,
			want: http.StatusNoContent, wantUser: &ReqUserInfo{Roles: []sysrole.Enum{sysrole.Tech}, UserId: 1, UserName: "Unauthed Dev"},
		},
		{name: "dev without authentication: user info missing", env: appenv.Dev, useAuth: false, withInfo: false, want: http.StatusInternalServerError},
		{name: "dev with authentication: no token", env: appenv.Dev, useAuth: true, withInfo: true, want: http.StatusForbidden},
		{name: "prod ignores the bypass setting: no token", env: appenv.Prod, useAuth: false, withInfo: true, want: http.StatusForbidden},
		{name: "prod ignores the bypass setting: token not found", env: appenv.Prod, useAuth: false, withInfo: true, header: "Bearer nope", want: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestAppNoDb(t)
			srvApp.Config.General.Env = tt.env
			srvApp.Config.Api.UseAuthentication = tt.useAuth

			next := &nextRecorder{}
			var h http.Handler = srvApp.authenticate(next)
			if tt.withInfo {
				h = withUserInfo(h)
			}

			req := newTestRequest(t)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			rec := testServe(t, h, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
			if wantReached := tt.want == http.StatusNoContent; next.called != wantReached {
				t.Errorf("next called = %v, want %v", next.called, wantReached)
			}
			if tt.wantUser != nil && !reflect.DeepEqual(next.userInfo, tt.wantUser) {
				t.Errorf("user info = %+v, want %+v", next.userInfo, tt.wantUser)
			}
		})
	}
}

func TestAuthenticateSession(t *testing.T) {
	tests := []struct {
		name       string
		roles      []string
		header     func(token string) string
		userAgent  string // default: the test client
		remoteAddr string // default: the test client
		withInfo   bool
		want       int
		wantUser   *ReqUserInfo
	}{
		{
			name: "valid session", roles: []string{"Standard", "Viewer"}, withInfo: true,
			header: func(token string) string { return "Bearer " + token },
			want:   http.StatusNoContent,
			wantUser: &ReqUserInfo{
				Roles: []sysrole.Enum{sysrole.Standard, sysrole.Viewer}, UserId: 42, UserName: "alice",
			},
		},
		{name: "no header", roles: []string{"Viewer"}, withInfo: true, header: func(string) string { return "" }, want: http.StatusForbidden},
		{name: "not a bearer header", roles: []string{"Viewer"}, withInfo: true, header: func(token string) string { return "Basic " + token }, want: http.StatusForbidden},
		{name: "unknown token", roles: []string{"Viewer"}, withInfo: true, header: func(string) string { return "Bearer unknown" }, want: http.StatusForbidden},
		{
			name: "different user agent", roles: []string{"Viewer"}, withInfo: true, userAgent: "other-browser",
			header: func(token string) string { return "Bearer " + token }, want: http.StatusForbidden,
		},
		{
			name: "different ip", roles: []string{"Viewer"}, withInfo: true, remoteAddr: "192.0.2.99:5000",
			header: func(token string) string { return "Bearer " + token }, want: http.StatusForbidden,
		},
		{
			name: "invalid role in session", roles: []string{"NotARole"}, withInfo: true,
			header: func(token string) string { return "Bearer " + token }, want: http.StatusInternalServerError,
		},
		{
			name: "user info missing from context", roles: []string{"Viewer"}, withInfo: false,
			header: func(token string) string { return "Bearer " + token }, want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srvApp := newTestAppNoDb(t)
			token := addTestSession(t, srvApp, 42, "alice", tt.roles)

			next := &nextRecorder{}
			var h http.Handler = srvApp.authenticate(next)
			if tt.withInfo {
				h = withUserInfo(h)
			}

			req := newTestRequest(t)
			if hdr := tt.header(token); hdr != "" {
				req.Header.Set("Authorization", hdr)
			}
			if tt.userAgent != "" {
				req.Header.Set("User-Agent", tt.userAgent)
			}
			req.RemoteAddr = tt.remoteAddr

			rec := testServe(t, h, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
			if wantReached := tt.want == http.StatusNoContent; next.called != wantReached {
				t.Errorf("next called = %v, want %v", next.called, wantReached)
			}
			if tt.wantUser != nil && !reflect.DeepEqual(next.userInfo, tt.wantUser) {
				t.Errorf("user info = %+v, want %+v", next.userInfo, tt.wantUser)
			}
		})
	}
}
