package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys/lysauth"
)

type httpServerApplication struct {
	*cmd.Application

	GetOptions  lys.GetOptions
	PostOptions lys.PostOptions

	// add login attempts so that suppsrv can use rejectBlockedIp middleware
	// at the moment, suppsrv restart is needed to refresh the internal blocked IPs. In real app, could add a db listener
	LoginAttempts    *lysauth.AppLoginAttempts
	UseXForwardedFor bool
	XForwardedForIdx int

	RateLimits *lysauth.AppRateLimits

	BlockedIPStore sysblockedip.Store
	SrvLogStore    syssrvreq.Store
}

// limit is middleware that applies rate limiting to all requests based on IP
func (srvApp *httpServerApplication) limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// get remote IP
		ip, err := lysauth.GetRemoteHostIP(r, srvApp.UseXForwardedFor, srvApp.XForwardedForIdx)
		if err != nil {
			lys.HandleInternalError(r.Context(), fmt.Errorf("limit: lysauth.GetRemoteHostIP failed: %w", err), srvApp.ErrorLog, w)
			return
		}

		// check rate limit
		if !srvApp.RateLimits.Allow(ip.String()) {
			lys.HandleUserError(lysauth.ErrTooManyRequests, w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// logAuthedRequest is middleware that logs requests in the app's info log
func (srvApp *httpServerApplication) logAuthedRequest(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get user from request
		userInfo, ok := ctx.Value(lys.UserInfoCtxKey).(ReqUserInfo)
		if !ok {
			lys.HandleInternalError(ctx, fmt.Errorf("logAuthedRequest: user not authenticated"), srvApp.ErrorLog, w)
			return
		}

		// get remote ip
		remoteHostIP, err := lysauth.GetRemoteHostIP(r, srvApp.UseXForwardedFor, srvApp.XForwardedForIdx)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("logAuthedRequest: lysauth.GetRemoteHostIP failed: %w", err), srvApp.ErrorLog, w)
			return
		}

		// use a custom writer that captures the response status code
		sw := &lys.StatusWriter{ResponseWriter: w}

		start := time.Now()
		next.ServeHTTP(sw, r)
		duration := time.Since(start)

		srvApp.InfoLog.Info(fmt.Sprintf("%s - %s - %s %s %s - %d - %dms",
			remoteHostIP, userInfo.UserName, r.Proto, r.Method, r.URL.RequestURI(), sw.Status, duration.Milliseconds()))

		// in real app: also log to db for activity monitoring as shown in refsrv
	})
}

func (srvApp *httpServerApplication) runRateLimitCleanup(ctx context.Context, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			srvApp.RateLimits.CleanupExpired()
		case <-ctx.Done():
			return
		}
	}
}

// secureHeaders is middleware that add XSS protection headers to responses
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}
