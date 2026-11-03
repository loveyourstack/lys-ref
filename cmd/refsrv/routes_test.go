package main

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys-ref/internal/testutil"
	"github.com/loveyourstack/lys/lysauth"
)

type testRoute struct {
	method string
	path   string // concrete path: route variables replaced with a value
}

func (r testRoute) String() string { return r.method + " " + r.path }

var routeVarRe = regexp.MustCompile(`\{[^}]*\}`)

// walkRoutes returns every method + path registered on the router.
func walkRoutes(t testing.TB, router *mux.Router) []testRoute {
	t.Helper()

	var routes []testRoute
	err := router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {

		// subrouter containers have no path template or methods of their own
		tmpl, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}

		for _, method := range methods {
			routes = append(routes, testRoute{method: method, path: routeVarRe.ReplaceAllString(tmpl, "1")})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("router.Walk() error = %v", err)
	}
	if len(routes) == 0 {
		t.Fatal("router has no routes")
	}

	return routes
}

func isAuthedRoute(r testRoute) bool {
	return strings.HasPrefix(r.path, "/a/")
}

func isRejected(status int) bool {
	return status == http.StatusUnauthorized || status == http.StatusForbidden
}

// Every route outside the public allowlist must be behind the authenticated subrouter.
func TestRoutesRequireAuth(t *testing.T) {
	srvApp := newTestAppNoDb(t)
	router := srvApp.getMuxRouter()
	routes := walkRoutes(t, router)

	// routes that are intentionally reachable without a session
	wantPublic := []string{"POST /login", "POST /session-token-login"}

	var gotPublic []string
	for _, route := range routes {
		if !isAuthedRoute(route) {
			gotPublic = append(gotPublic, route.String())
		}
	}
	slices.Sort(gotPublic)
	if !slices.Equal(gotPublic, wantPublic) {
		t.Errorf("public routes = %v, want %v", gotPublic, wantPublic)
	}

	tests := []struct {
		name       string
		authHeader string
	}{
		{name: "no auth header"},
		{name: "malformed auth header", authHeader: "Token abc"},
		{name: "unknown token", authHeader: "Bearer 0123456789012345678901234"},
	}

	for _, route := range routes {
		if !isAuthedRoute(route) {
			continue
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s/%s", route, tt.name), func(t *testing.T) {
				req, _ := http.NewRequest(route.method, route.path, nil)
				if tt.authHeader != "" {
					req.Header.Set("Authorization", tt.authHeader)
				}

				rec := testServe(t, router, req)
				if !isRejected(rec.Code) {
					t.Errorf("status = %d, want 401 or 403", rec.Code)
				}
			})
		}
	}
}

// Routes which change data must reject users without a writer role, and the tech-only subroutes must reject non-tech users.
func TestRoutesRoleRestrictions(t *testing.T) {
	srvApp := newTestApp(t)
	srvApp.AuthedRateLimits = lysauth.NewAppRateLimits(1e6, 1e6, time.Hour) // one session sends hundreds of requests
	router := srvApp.getMuxRouter()
	routes := walkRoutes(t, router)

	ownerDb := testutil.NewOwnerPool(t, srvApp.Config)

	// subroutes restricted to the Tech role
	techOnlyPrefixes := []string{"/a/pg-monitor/", "/a/tech/"}

	// authed routes which use a mutating method but are intentionally open to non-writers
	nonWriterMutations := map[string]bool{
		"POST /a/logout":                             true,
		"POST /a/aws/bedrock-list-image-models":      true, // read-only listing
		"POST /a/digmark/mcp-query":                  true, // read-only query
		"POST /a/system/notifications/add-fake":      true, // only affects the user's own notifications
		"PATCH /a/system/notifications/set-all-read": true, // only affects the user's own notifications
		"PATCH /a/system/notifications/set-read":     true, // only affects the user's own notifications
	}

	sessionFor := func(userId int64, userName string, role sysrole.Enum) string {
		t.Helper()

		// requests by an authenticated user are logged to the db: remove them afterwards
		t.Cleanup(func() {
			if _, err := ownerDb.Exec(context.Background(), "DELETE FROM system.srv_request WHERE user_name = $1", userName); err != nil {
				t.Errorf("deleting srv_request rows: %v", err)
			}
		})

		token, err := srvApp.Sessions.Add(lysauth.SessionInput{
			Email:               userName + "@example.com",
			FamilyName:          "Test",
			GeoIpCountryIsoCode: "ZZ",
			GeoIpLocation:       "Test",
			GivenName:           "Test",
			Ip:                  netip.MustParseAddrPort(testRemoteAddr).Addr(),
			Roles:               []string{role.String()},
			UserAgent:           testUserAgent,
			UserId:              userId,
			UserName:            userName,
		})
		if err != nil {
			t.Fatalf("Sessions.Add() error = %v", err)
		}
		return token
	}

	viewerToken := sessionFor(1, "route-test-viewer", sysrole.Viewer)
	standardToken := sessionFor(2, "route-test-standard", sysrole.Standard)

	isMutation := func(method string) bool {
		return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
	}
	isTechOnly := func(path string) bool {
		return slices.ContainsFunc(techOnlyPrefixes, func(p string) bool { return strings.HasPrefix(path, p) })
	}

	checked := 0
	for _, route := range routes {
		if !isAuthedRoute(route) || nonWriterMutations[route.String()] {
			continue
		}

		var token, who string
		switch {
		case isTechOnly(route.path):
			token, who = standardToken, "standard user on tech-only route"
		case isMutation(route.method):
			token, who = viewerToken, "viewer on writer-only route"
		default:
			continue
		}
		checked++

		t.Run(fmt.Sprintf("%s/%s", route, who), func(t *testing.T) {
			req, _ := http.NewRequest(route.method, route.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			rec := testServe(t, router, req)
			if rec.Code != http.StatusForbidden {
				t.Errorf("status = %d, want 403", rec.Code)
			}
		})
	}
	if checked == 0 {
		t.Fatal("no routes were checked")
	}
}
