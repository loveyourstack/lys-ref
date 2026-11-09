package main

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
)

type httpServerApplication struct {
	*cmd.Application

	GetOptions  lysapi.GetOptions
	PostOptions lysapi.PostOptions

	// add login attempts so that suppsrv can use rejectBlockedIp middleware
	// at the moment, suppsrv restart is needed to refresh the internal blocked IPs. In real app, could add a db listener
	LoginAttempts   *lysauth.AppLoginAttempts
	UseRealIpHeader bool

	RateLimits *lysauth.AppRateLimits

	BlockedIPStore sysblockedip.Store
	SrvLogStore    syssrvreq.Store
}

type clientIpCtxKeyType struct{}

var clientIpCtxKey = clientIpCtxKeyType{}

// initClientIP resolves the client IP once for the rest of the request chain.
func (srvApp *httpServerApplication) initClientIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		remoteHostIP, err := lysauth.GetIpAddrFromReq(r, srvApp.UseRealIpHeader)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("initClientIP: lysauth.GetIpAddrFromReq failed: %w", err), srvApp.Logger, w)
			return
		}
		ctx = context.WithValue(ctx, clientIpCtxKey, remoteHostIP)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func clientIPFromContext(ctx context.Context) (netip.Addr, error) {
	ip, ok := ctx.Value(clientIpCtxKey).(netip.Addr)
	if !ok || !ip.IsValid() {
		return netip.Addr{}, fmt.Errorf("client IP missing or invalid in context")
	}
	return ip, nil
}

// limit is middleware that applies rate limiting to all requests based on IP
func (srvApp *httpServerApplication) limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get remote IP from context
		remoteHostIP, err := clientIPFromContext(ctx)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("limit: %w", err), srvApp.Logger, w)
			return
		}

		// check rate limit
		if !srvApp.RateLimits.Allow(remoteHostIP.String()) {
			lys.HandleUserError(lysauth.ErrTooManyRequests, w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// logRequest is middleware that logs requests in the app's info log
func (srvApp *httpServerApplication) logRequest(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get remote IP from context
		remoteHostIP, err := clientIPFromContext(ctx)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("logRequest: %w", err), srvApp.Logger, w)
			return
		}

		// create user info and attach to context, to be filled in by authenticate and login later in the chain
		// must be a pointer so that it can be overwritten
		userInfo := &ReqUserInfo{}
		ctx = lys.WithUserInfo(ctx, userInfo)

		// use a custom writer that captures the response status code
		sw := &lys.StatusWriter{ResponseWriter: w}

		start := time.Now()
		next.ServeHTTP(sw, r.WithContext(ctx))
		duration := time.Since(start)

		srvApp.Logger.Info(fmt.Sprintf("%s - %s - %s %s %s - %d - %dms",
			remoteHostIP, cmp.Or(userInfo.UserName, "-"), r.Proto, r.Method, r.URL.RequestURI(), sw.Status, duration.Milliseconds()))

		// in real app: also log to db for activity monitoring as shown in refsrv
	})
}

// secureHeaders is middleware that add XSS protection headers to responses
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}
