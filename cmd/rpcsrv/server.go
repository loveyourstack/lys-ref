package main

import (
	"cmp"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/loveyourstack/connectors/maxmind/stores/mmlocation"
	"github.com/loveyourstack/connectors/maxmind/stores/mmnetwork"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcreq"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsession"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsessionhist"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssvcaccount"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type gRpcServerApplication struct {
	*cmd.Application

	GetOptions lysapi.GetOptions

	// auth-related
	LoginAttempts   *lysauth.AppLoginAttempts
	Sessions        *lysauth.AppSessions
	UseRealIpHeader bool // whether to use Real-IP header to determine IP

	// rate limits
	AuthedRateLimits   *lysauth.AppRateLimits // more generous than unauthed
	UnauthedRateLimits *lysauth.AppRateLimits

	ShutdownTimeouts shutdownTimeouts

	// stores
	BlockedIPStore    sysblockedip.Store
	CountryStore      geocountry.Store
	GeoLocationStore  mmlocation.Store
	GeoNetworkStore   mmnetwork.Store
	LoginAttemptStore sysrpcloginattempt.Store
	SessionStore      sysrpcsession.Store
	SessionHistStore  sysrpcsessionhist.Store
	SrvLogStore       sysrpcreq.Store
	SvcAccountStore   syssvcaccount.Store
	UserStore         sysuser.Store
}

// limitUnauthed rate limits public methods by IP. Authenticated methods are limited by user ID in limitAuthed
func (rpcApp *gRpcServerApplication) limitUnauthed(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {

	// only apply rate limiting to public methods
	// continue on methods with no policy: authenticate will handle them
	if policy, ok := methodPolicies[info.FullMethod]; !ok || !policy.public {
		return handler(ctx, req)
	}

	// get remote IP from context
	remoteHostIp, err := lysauth.ClientIpFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("lysauth.ClientIpFromCtx failed: %w", err)
	}

	if !rpcApp.UnauthedRateLimits.Allow(remoteHostIp.String()) {
		return nil, lysauth.ErrTooManyRequests
	}

	return handler(ctx, req)
}

// limitAuthed rate limits authenticated requests by user ID. Must run after authenticate
func (rpcApp *gRpcServerApplication) limitAuthed(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {

	userInfo, ok := lys.GetUserInfoFromCtx(ctx).(*ReqUserInfo)
	if !ok {
		return nil, lys.ErrUserInfoMissing
	}

	// public methods skip authentication and were already limited by limitUnauthed
	if userInfo.UserId == 0 {
		return handler(ctx, req)
	}

	if !rpcApp.AuthedRateLimits.Allow(strconv.FormatInt(userInfo.UserId, 10)) {
		return nil, lysauth.ErrTooManyRequests
	}

	return handler(ctx, req)
}

func (rpcApp *gRpcServerApplication) logRequest(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {

	// get remote IP from context
	remoteHostIp, err := lysauth.ClientIpFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("lysauth.ClientIpFromCtx failed: %w", err)
	}

	// create user info and attach to context, to be filled in by authenticate and login later in the chain
	// must be a pointer so that it can be overwritten
	userInfo := &ReqUserInfo{}
	ctx = lys.WithUserInfo(ctx, userInfo)

	// run handler with timing
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	// code as determined by mapError
	code := status.Code(err)

	// log to file
	rpcApp.Logger.Info(fmt.Sprintf("%s - %s - %s - %s - %dms",
		remoteHostIp, cmp.Or(userInfo.UserName, "-"), info.FullMethod, code, duration.Milliseconds()))

	// exit if user is not authenticated
	if userInfo.UserName == "" {
		return resp, err
	}

	// log to db for activity monitoring

	// separate context so logging happens even if the request context is cancelled
	_, insErr := rpcApp.SrvLogStore.Insert(context.Background(), sysrpcreq.Input{
		DurationMs: duration.Milliseconds(),
		Endpoint:   info.FullMethod,
		Ip:         remoteHostIp,
		Status:     code.String(),
		UserName:   userInfo.UserName,
	})
	if insErr != nil {
		rpcApp.Logger.Error("logRequest: SrvLogStore.Insert failed", "error", insErr)
	}

	return resp, err
}
