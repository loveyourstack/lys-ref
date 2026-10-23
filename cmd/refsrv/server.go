package main

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/loveyourstack/connectors/aws/stores/awsusersgrule"
	"github.com/loveyourstack/connectors/maxmind/stores/mmlocation"
	"github.com/loveyourstack/connectors/maxmind/stores/mmnetwork"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/connectors/awsbedrockapi"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssession"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssessionhist"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lysws"
)

type httpServerApplication struct {
	*cmd.Application

	GetOptions  lysapi.GetOptions
	PostOptions lysapi.PostOptions

	// auth-related
	LoginAttempts    *lysauth.AppLoginAttempts
	Sessions         *lysauth.AppSessions
	UseXForwardedFor bool // whether to use X-Forwarded-For header to determine IP
	XForwardedForIdx int  // if using X-Forwarded-For, which index to use (0 for first, etc)

	// rate limits
	AuthedRateLimits   *lysauth.AppRateLimits // more generous than unauthed
	UnauthedRateLimits *lysauth.AppRateLimits

	// ws notifications
	NotificationHub *lysws.NotificationHub

	// stores
	AwsUserSgRuleStore awsusersgrule.Store
	BlockedIPStore     sysblockedip.Store
	CountryStore       geocountry.Store
	GeoLocationStore   mmlocation.Store
	GeoNetworkStore    mmnetwork.Store
	LoginAttemptStore  sysloginattempt.Store
	SessionStore       syssession.Store
	SessionHistStore   syssessionhist.Store
	SrvLogStore        syssrvreq.Store
	UserStore          sysuser.Store

	// clients
	AwsBedrockClient *awsbedrockapi.Client
}

// limitAuthed is middleware that applies rate limiting to authenticated users based on user ID
func (srvApp *httpServerApplication) limitAuthed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// get user info from context (set by authenticate middleware)
		userInfo, ok := r.Context().Value(lys.UserInfoCtxKey).(*ReqUserInfo)
		if !ok {
			lys.HandleUserError(lys.ErrUserInfoMissing, w)
			return
		}
		key := strconv.FormatInt(userInfo.UserId, 10)

		// check rate limit
		if !srvApp.AuthedRateLimits.Allow(key) {
			lys.HandleUserError(lysauth.ErrTooManyRequests, w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// limitUnauthed is middleware that applies rate limiting to unauthenticated users based on IP
func (srvApp *httpServerApplication) limitUnauthed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// get remote IP
		ip, err := lysauth.GetIpAddrFromReq(r, srvApp.UseXForwardedFor, srvApp.XForwardedForIdx)
		if err != nil {
			lys.HandleInternalError(r.Context(), fmt.Errorf("limitUnauthed: lysauth.GetIpAddrFromReq failed: %w", err), srvApp.Logger, w)
			return
		}

		// check rate limit
		if !srvApp.UnauthedRateLimits.Allow(ip.String()) {
			lys.HandleUserError(lysauth.ErrTooManyRequests, w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// logRequest is middleware that logs requests in the app's info log and the database for activity monitoring.
func (srvApp *httpServerApplication) logRequest(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// create user info and attach to context, to be filled in by authenticate and login later in the chain
		// must be a pointer so that it can be overwritten
		userInfo := &ReqUserInfo{}
		ctx = context.WithValue(ctx, lys.UserInfoCtxKey, userInfo)

		// exclude websocket registration: request is long-lived
		if r.URL.Path == "/a/ws/notifications/register" {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// don't log the sessions monitor (auto-refreshes in UI)
		if r.URL.Path == "/a/tech/sessions" && r.Method == http.MethodGet {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// get remote ip
		remoteHostIP, err := lysauth.GetIpAddrFromReq(r, srvApp.UseXForwardedFor, srvApp.XForwardedForIdx)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("logRequest: lysauth.GetIpAddrFromReq failed: %w", err), srvApp.Logger, w)
			return
		}

		// use a custom writer that captures the response status code
		sw := &lys.StatusWriter{ResponseWriter: w}

		start := time.Now()
		next.ServeHTTP(sw, r.WithContext(ctx)) // don't return responses after this point: log errors only
		duration := time.Since(start)

		status := sw.Status
		// default to 200 if status is not set, e.g. for websocket upgrade requests
		if status == 0 {
			status = http.StatusOK
		}

		// for ws: strip out token url param
		urlVals := r.URL.Query()
		urlVals.Del("token")
		r.URL.RawQuery = urlVals.Encode()

		// interesting: use client such as Postman to compare the logged duration with the actual duration to measure the middleware+logging overhead (c. 8-12ms on dev)

		// log to file
		srvApp.Logger.Info(fmt.Sprintf("%s - %s - %s %s %s - %d - %dms",
			remoteHostIP, cmp.Or(userInfo.UserName, "-"), r.Proto, r.Method, r.URL.RequestURI(), status, duration.Milliseconds()))

		// exit if user is not authenticated
		if userInfo.UserId == 0 {
			return
		}

		// unescape url for readability
		endpoint, err := url.QueryUnescape(r.URL.RequestURI())
		if err != nil {
			srvApp.Logger.Error("logRequest: url.QueryUnescape failed", "error", err)
		}

		// log to db for activity monitoring

		// use separate context to ensure logging happens even if request context is cancelled
		logCtx := context.Background()
		_, err = srvApp.SrvLogStore.Insert(logCtx, syssrvreq.Input{
			DurationMs: duration.Milliseconds(),
			Endpoint:   endpoint,
			Ip:         remoteHostIP,
			Method:     r.Method,
			StatusCode: status,
			UserName:   userInfo.UserName,
		})
		if err != nil {
			srvApp.Logger.Error("logRequest: failed to insert log", "error", err)
		}
	})
}

func (srvApp *httpServerApplication) runRateLimitCleanup(ctx context.Context, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			srvApp.AuthedRateLimits.CleanupExpired()
			srvApp.UnauthedRateLimits.CleanupExpired()
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
