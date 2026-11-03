package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/netip"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys-ref/internal/testutil"
	"github.com/loveyourstack/lys/lysauth"
	"golang.org/x/crypto/bcrypt"
)

const testPassword = "correct-horse-battery-staple"

// authTestEnv is an app on the test db plus an owner connection for fixtures and assertions.
type authTestEnv struct {
	srvApp  *httpServerApplication
	router  http.Handler
	ownerDb *pgxpool.Pool
}

func newAuthTestEnv(t *testing.T) *authTestEnv {
	t.Helper()

	srvApp := newTestApp(t)
	srvApp.UnauthedRateLimits = lysauth.NewAppRateLimits(1e6, 1e6, time.Hour) // rate limiting is not under test here
	env := &authTestEnv{
		srvApp:  srvApp,
		router:  srvApp.getRouter(),
		ownerDb: testutil.NewOwnerPool(t, srvApp.Config),
	}

	// blocking the test client persists its IP: make sure no run is affected by a previous one
	deleteBlockedIp := func() {
		if _, err := env.ownerDb.Exec(context.Background(), "DELETE FROM system.blocked_ip WHERE ip = $1::inet", testRemoteIp()); err != nil {
			t.Errorf("deleting blocked ip: %v", err)
		}
	}
	deleteBlockedIp()
	t.Cleanup(deleteBlockedIp)

	return env
}

func testRemoteIp() string {
	return testRemoteAddr[:strings.LastIndex(testRemoteAddr, ":")]
}

type testUser struct {
	id       int64
	name     string
	password string
}

// newUser inserts a Viewer user with a unique name and removes it and everything created for it when the test ends.
func (env *authTestEnv) newUser(t *testing.T, allowMultipleSessions, isDeactivated bool) testUser {
	t.Helper()

	ctx := context.Background()

	hashedPw, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() error = %v", err)
	}

	name := fmt.Sprintf("authtest_%08x", rand.Uint32())
	id, err := sysuser.Store{Db: env.ownerDb}.Insert(ctx, sysuser.Input{
		AllowMultipleSessions: allowMultipleSessions,
		Email:                 name + "@example.com",
		FamilyName:            "Test",
		GivenName:             "Test",
		HashedPw:              string(hashedPw),
		Name:                  name,
		Roles:                 []sysrole.Enum{sysrole.Viewer},
	})
	if err != nil {
		t.Fatalf("inserting user: %v", err)
	}

	t.Cleanup(func() {
		for _, stmt := range []struct {
			sql string
			arg any
		}{
			{"DELETE FROM system.session_history WHERE user_fk = $1", id},
			{"DELETE FROM system.srv_request WHERE user_name = $1", name},
			{"DELETE FROM system.user WHERE id = $1", id},
		} {
			if _, err := env.ownerDb.Exec(ctx, stmt.sql, stmt.arg); err != nil {
				t.Errorf("cleanup %q: %v", stmt.sql, err)
			}
		}
	})

	if isDeactivated {
		// Store.Deactivate also renames the user, which would hide the deactivated check from the login handler
		if _, err := env.ownerDb.Exec(ctx, "UPDATE system.user SET is_deactivated = true WHERE id = $1", id); err != nil {
			t.Fatalf("deactivating user: %v", err)
		}
	}

	return testUser{id: id, name: name, password: testPassword}
}

func (env *authTestEnv) countHistory(t *testing.T, userId int64) (n int) {
	t.Helper()

	err := env.ownerDb.QueryRow(context.Background(), "SELECT count(*) FROM system.session_history WHERE user_fk = $1", userId).Scan(&n)
	if err != nil {
		t.Fatalf("counting session history: %v", err)
	}
	return n
}

func (env *authTestEnv) countBlockedIp(t *testing.T) (n int) {
	t.Helper()

	err := env.ownerDb.QueryRow(context.Background(), "SELECT count(*) FROM system.blocked_ip WHERE ip = $1::inet", testRemoteIp()).Scan(&n)
	if err != nil {
		t.Fatalf("counting blocked ips: %v", err)
	}
	return n
}

// post sends a JSON body (or raw string) to the router, optionally authenticated with a session token.
func (env *authTestEnv) post(t *testing.T, path string, body any, token string) *testResponse {
	t.Helper()

	var b []byte
	switch v := body.(type) {
	case nil:
	case string:
		b = []byte(v)
	default:
		var err error
		if b, err = json.Marshal(v); err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
	}

	req, err := http.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return env.send(t, req)
}

func (env *authTestEnv) send(t *testing.T, req *http.Request) *testResponse {
	t.Helper()

	rec := testServe(t, env.router, req)

	resp := &testResponse{code: rec.Code}
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp.body); err != nil {
			t.Fatalf("response body is not a StdResponse (status %d): %q", rec.Code, rec.Body.String())
		}
	}
	return resp
}

type testResponse struct {
	code int
	body struct {
		Status         string          `json:"status"`
		Data           json.RawMessage `json:"data"`
		ErrDescription string          `json:"err_description"`
	}
}

func (r *testResponse) login(t *testing.T) loginResponse {
	t.Helper()

	var lr loginResponse
	if err := json.Unmarshal(r.body.Data, &lr); err != nil {
		t.Fatalf("decoding login response %s: %v", r.body.Data, err)
	}
	return lr
}

// login logs the user in and returns the session token.
func (env *authTestEnv) login(t *testing.T, u testUser) string {
	t.Helper()

	resp := env.post(t, "/login", LoginInput{UserName: u.name, Password: u.password}, "")
	if resp.code != http.StatusOK {
		t.Fatalf("login status = %d (%s), want 200", resp.code, resp.body.ErrDescription)
	}
	return resp.login(t).SessionToken
}

func TestAuthLogin(t *testing.T) {
	env := newAuthTestEnv(t)

	active := env.newUser(t, false, false)
	deactivated := env.newUser(t, false, true)

	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string // substring of err_description
		wantUser   *testUser
	}{
		{name: "valid credentials", body: LoginInput{UserName: active.name, Password: active.password}, wantStatus: http.StatusOK, wantUser: &active},
		{name: "missing username", body: LoginInput{UserName: "", Password: "wrong"}, wantStatus: http.StatusBadRequest, wantErr: "UserName is required"},
		{name: "missing password", body: LoginInput{UserName: active.name, Password: ""}, wantStatus: http.StatusBadRequest, wantErr: "Password is required"},
		{name: "wrong password", body: LoginInput{UserName: active.name, Password: "wrong"}, wantStatus: http.StatusUnauthorized, wantErr: "invalid credentials"},
		{name: "unknown user", body: LoginInput{UserName: "authtest_nobody", Password: active.password}, wantStatus: http.StatusUnauthorized, wantErr: "invalid credentials"},
		{name: "deactivated user", body: LoginInput{UserName: deactivated.name, Password: deactivated.password}, wantStatus: http.StatusUnauthorized, wantErr: "invalid credentials"},
		{name: "malformed json", body: `{"user_name":`, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// fresh in-memory state per case
			env.srvApp.attachAuth()

			resp := env.post(t, "/login", tt.body, "")

			if resp.code != tt.wantStatus {
				t.Fatalf("status = %d (%s), want %d", resp.code, resp.body.ErrDescription, tt.wantStatus)
			}
			if !strings.Contains(resp.body.ErrDescription, tt.wantErr) {
				t.Errorf("err_description = %q, want it to contain %q", resp.body.ErrDescription, tt.wantErr)
			}

			if tt.wantUser == nil {
				if resp.body.Status == lys.ReqSucceeded {
					t.Errorf("status field = %q for a failed login", resp.body.Status)
				}
				if n := env.srvApp.Sessions.Count(); n != 0 {
					t.Errorf("sessions after failed login = %d, want 0", n)
				}
				return
			}

			if resp.body.Status != lys.ReqSucceeded {
				t.Errorf("status field = %q, want %q", resp.body.Status, lys.ReqSucceeded)
			}
			lr := resp.login(t)
			if lr.SessionToken == "" {
				t.Error("session token is empty")
			}
			if lr.UserId != tt.wantUser.id || lr.UserName != tt.wantUser.name {
				t.Errorf("user = %d %q, want %d %q", lr.UserId, lr.UserName, tt.wantUser.id, tt.wantUser.name)
			}
			if !slices.Equal(lr.Roles, []string{sysrole.Viewer.String()}) {
				t.Errorf("roles = %v, want [%s]", lr.Roles, sysrole.Viewer)
			}
			if n := env.srvApp.Sessions.Count(); n != 1 {
				t.Errorf("sessions after login = %d, want 1", n)
			}
			if n := len(env.srvApp.LoginAttempts.All()); n != 0 {
				t.Errorf("login attempts after successful login = %d, want 0", n)
			}
		})
	}
}

func TestAuthLoginAttempts(t *testing.T) {
	env := newAuthTestEnv(t)
	u := env.newUser(t, false, false)

	bad := LoginInput{UserName: u.name, Password: "wrong"}
	good := LoginInput{UserName: u.name, Password: u.password}

	t.Run("success resets attempts", func(t *testing.T) {
		env.srvApp.attachAuth()

		for range 3 {
			if resp := env.post(t, "/login", bad, ""); resp.code != http.StatusUnauthorized {
				t.Fatalf("bad login status = %d, want 401", resp.code)
			}
		}
		if resp := env.post(t, "/login", good, ""); resp.code != http.StatusOK {
			t.Fatalf("good login status = %d, want 200", resp.code)
		}

		if n := len(env.srvApp.LoginAttempts.All()); n != 0 {
			t.Errorf("login attempts = %d, want 0", n)
		}
	})

	t.Run("blocked after max attempts", func(t *testing.T) {
		env.srvApp.attachAuth()

		// 5 attempts are allowed
		for i := 1; i <= 5; i++ {
			if resp := env.post(t, "/login", bad, ""); resp.code != http.StatusUnauthorized {
				t.Fatalf("attempt %d: status = %d, want 401", i, resp.code)
			}
		}
		if n := env.countBlockedIp(t); n != 0 {
			t.Fatalf("blocked ips before max attempts exceeded = %d, want 0", n)
		}

		// the 6th blocks the IP and persists the block
		resp := env.post(t, "/login", bad, "")
		if resp.code != http.StatusForbidden || !strings.Contains(resp.body.ErrDescription, "Max login attempts") {
			t.Fatalf("attempt 6: status = %d (%s), want 403 max attempts exceeded", resp.code, resp.body.ErrDescription)
		}
		if n := env.countBlockedIp(t); n != 1 {
			t.Errorf("blocked ips after max attempts exceeded = %d, want 1", n)
		}

		// even correct credentials are now rejected
		resp = env.post(t, "/login", good, "")
		if resp.code != http.StatusForbidden || !strings.Contains(resp.body.ErrDescription, "blocked") {
			t.Fatalf("login while blocked: status = %d (%s), want 403 blocked", resp.code, resp.body.ErrDescription)
		}
		if n := env.srvApp.Sessions.Count(); n != 0 {
			t.Errorf("sessions while blocked = %d, want 0", n)
		}
	})
}

func TestAuthLoginSessions(t *testing.T) {
	tests := []struct {
		name                  string
		allowMultipleSessions bool
		wantSessions          int
		wantArchived          int
	}{
		{name: "single session user: first session is archived", allowMultipleSessions: false, wantSessions: 1, wantArchived: 1},
		{name: "multiple session user: both sessions are kept", allowMultipleSessions: true, wantSessions: 2, wantArchived: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newAuthTestEnv(t)
			u := env.newUser(t, tt.allowMultipleSessions, false)

			first := env.login(t, u)
			second := env.login(t, u)

			if first == second {
				t.Error("both logins returned the same token")
			}
			if n := env.srvApp.Sessions.Count(); n != tt.wantSessions {
				t.Errorf("sessions = %d, want %d", n, tt.wantSessions)
			}
			if n := env.countHistory(t, u.id); n != tt.wantArchived {
				t.Errorf("archived sessions = %d, want %d", n, tt.wantArchived)
			}
		})
	}
}

func TestAuthLogout(t *testing.T) {
	env := newAuthTestEnv(t)
	u := env.newUser(t, false, false)
	token := env.login(t, u)

	if resp := env.post(t, "/a/logout", nil, ""); resp.code != http.StatusForbidden {
		t.Errorf("logout without token: status = %d, want 403", resp.code)
	}
	if n := env.srvApp.Sessions.Count(); n != 1 {
		t.Fatalf("sessions after rejected logout = %d, want 1", n)
	}

	resp := env.post(t, "/a/logout", nil, token)
	if resp.code != http.StatusOK {
		t.Fatalf("logout status = %d (%s), want 200", resp.code, resp.body.ErrDescription)
	}
	if n := env.srvApp.Sessions.Count(); n != 0 {
		t.Errorf("sessions after logout = %d, want 0", n)
	}
	if n := env.countHistory(t, u.id); n != 1 {
		t.Errorf("archived sessions after logout = %d, want 1", n)
	}

	// the token is no longer valid
	if resp := env.post(t, "/a/logout", nil, token); resp.code != http.StatusForbidden {
		t.Errorf("logout with used token: status = %d, want 403", resp.code)
	}
}

func TestAuthSessionTokenLogin(t *testing.T) {
	env := newAuthTestEnv(t)
	u := env.newUser(t, false, false)
	token := env.login(t, u)

	tests := []struct {
		name       string
		token      string
		userAgent  string
		remoteAddr string
		wantStatus int
	}{
		{name: "valid session", token: token, wantStatus: http.StatusOK},
		{name: "no token", wantStatus: http.StatusForbidden},
		{name: "unknown token", token: "0123456789012345678901234", wantStatus: http.StatusForbidden},
		{name: "different user agent", token: token, userAgent: "other-browser", wantStatus: http.StatusForbidden},
		{name: "different ip", token: token, remoteAddr: "192.0.2.99:5000", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/session-token-login", nil)
			if err != nil {
				t.Fatalf("http.NewRequest() error = %v", err)
			}
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			if tt.userAgent != "" {
				req.Header.Set("User-Agent", tt.userAgent)
			}
			req.RemoteAddr = tt.remoteAddr

			resp := env.send(t, req)
			if resp.code != tt.wantStatus {
				t.Fatalf("status = %d (%s), want %d", resp.code, resp.body.ErrDescription, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				lr := resp.login(t)
				if lr.SessionToken != token || lr.UserId != u.id || lr.UserName != u.name {
					t.Errorf("response = token %q user %d %q, want the existing session of %d %q", lr.SessionToken, lr.UserId, lr.UserName, u.id, u.name)
				}
			}
		})
	}
}

func TestAuthGetSignInResponse(t *testing.T) {
	ip := netip.MustParseAddr("192.0.2.10")

	got := authGetSignInResponse("de", true, "DE", "Berlin, Germany", true, ip, "pic.png", []string{"Tech", "Viewer"}, "tok", 42, "alice")
	want := loginResponse{
		DefaultLocale:       "de",
		ForcePasswordChange: true,
		HasAwsSgRules:       true,
		GeoIpCountryIsoCode: "DE",
		GeoIpLocation:       "Berlin, Germany",
		Ip:                  ip,
		ProfilePic:          "pic.png",
		Roles:               []string{"Tech", "Viewer"},
		SessionToken:        "tok",
		UserId:              42,
		UserName:            "alice",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("authGetSignInResponse() = %+v, want %+v", got, want)
	}

	// the token is the only credential: it must be sent as "token", the user id as "user_id"
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(b, &fields); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if fields["token"] != "tok" || fields["user_id"] != float64(42) || fields["ip"] != "192.0.2.10" {
		t.Errorf("json fields = %v, want token, user_id and ip set", fields)
	}
}
