package main

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"net/url"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type httpServerApplication struct {
	*cmd.Application

	GetOptions  lysapi.GetOptions
	PostOptions lysapi.PostOptions

	// auth-related
	LoginAttempts   *lysauth.AppLoginAttempts
	Sessions        *lysauth.AppSessions
	UseRealIpHeader bool // whether to use Real-IP header to determine IP

	// rate limits
	AuthedRateLimits   *lysauth.AppRateLimits // more generous than unauthed
	UnauthedRateLimits *lysauth.AppRateLimits

	// connection to internal RPC server
	rpcClient *grpc.ClientConn
	RpcApiKey string

	// ws notifications
	NotificationHub *lysws.NotificationHub

	ShutdownTimeouts shutdownTimeouts

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

// logRequest is middleware that logs requests in the app's info log and the database for activity monitoring.
func (srvApp *httpServerApplication) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get remote IP from context
		remoteHostIP, err := lysauth.ClientIpFromCtx(ctx)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("logRequest: %w", err), srvApp.Logger, w)
			return
		}

		// create user info and attach to context, to be filled in by authenticate and login later in the chain
		// must be a pointer so that it can be overwritten
		userInfo := &ReqUserInfo{}
		ctx = lys.WithUserInfo(ctx, userInfo)

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
		if userInfo.UserName == "" {
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

// rpcDefaultDeadlineUnaryInterceptor applies a timeout to RPC calls whose context has no deadline yet.
// A deadline set by the caller is left alone.
func rpcDefaultDeadlineUnaryInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// rpcForwardIpUnaryInterceptor sets x-real-ip on outgoing RPCs to the end user's IP, which the RPC server requires in prod
func rpcForwardIpUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ip, err := lysauth.ClientIpFromCtx(ctx)
		if err != nil {
			return fmt.Errorf("rpcForwardIpUnaryInterceptor: %w", err)
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", ip.String())
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// rpcKeyCredentials attaches the API key to every outgoing RPC as metadata.
type rpcKeyCredentials struct{ key string }

func (c rpcKeyCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"x-api-key": c.key}, nil
}

// RequireTransportSecurity is false because the RPC server is only reached over loopback.
func (c rpcKeyCredentials) RequireTransportSecurity() bool { return false }
