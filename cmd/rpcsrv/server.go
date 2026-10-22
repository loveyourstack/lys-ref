package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyserr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRpcServerApplication struct {
	*cmd.Application

	GetOptions lysapi.GetOptions

	// auth-related
	LoginAttempts    *lysauth.AppLoginAttempts
	Sessions         *lysauth.AppSessions
	UseXForwardedFor bool // whether to use X-Forwarded-For header to determine IP
	XForwardedForIdx int  // if using X-Forwarded-For, which index to use (0 for first, etc)

	// rate limits
	AuthedRateLimits   *lysauth.AppRateLimits // more generous than unauthed
	UnauthedRateLimits *lysauth.AppRateLimits

	// stores
	BlockedIPStore    sysblockedip.Store
	CountryStore      geocountry.Store
	GeoLocationStore  mmlocation.Store
	GeoNetworkStore   mmnetwork.Store
	LoginAttemptStore sysrpcloginattempt.Store
	SessionStore      sysrpcsession.Store
	SessionHistStore  sysrpcsessionhist.Store
	SrvLogStore       sysrpcreq.Store
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

	ip, err := getRemoteHostIp(ctx, rpcApp.UseXForwardedFor, rpcApp.XForwardedForIdx)
	if err != nil {
		return nil, fmt.Errorf("limitUnauthed: getRemoteHostIp failed: %w", err)
	}

	if !rpcApp.UnauthedRateLimits.Allow(ip.String()) {
		return nil, lysauth.ErrTooManyRequests
	}

	return handler(ctx, req)
}

// limitAuthed rate limits authenticated requests by user ID. Must run after authenticate
func (rpcApp *gRpcServerApplication) limitAuthed(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {

	userInfo, ok := ctx.Value(lys.UserInfoCtxKey).(*ReqUserInfo)
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

	// create user info and attach to context, to be filled in by authenticate and login later in the chain
	// must be a pointer so that it can be overwritten
	userInfo := &ReqUserInfo{}
	ctx = context.WithValue(ctx, lys.UserInfoCtxKey, userInfo)

	// run handler with timing
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	// code as determined by mapError
	code := status.Code(err)

	// get remote ip
	remoteHostIp, ipErr := getRemoteHostIp(ctx, rpcApp.UseXForwardedFor, rpcApp.XForwardedForIdx)
	if ipErr != nil {
		rpcApp.Logger.Error("logRequest: getRemoteHostIp failed", "error", ipErr)
		return resp, err
	}

	// log to file
	rpcApp.Logger.Info(fmt.Sprintf("%s - %s - %s - %s - %dms",
		remoteHostIp, cmp.Or(userInfo.UserName, "-"), info.FullMethod, code, duration.Milliseconds()))

	// exit if user is not authenticated
	if userInfo.UserId == 0 {
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

func (rpcApp *gRpcServerApplication) mapError(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {

	// call the handler and return if no error occurred
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	// an error occurred

	code := status.Code(err)
	errMsg := err.Error()

	// see if err can be unwrapped to a pgx PgError
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {

		switch pgErr.Code {

		// define text and status for errors that can be attributed to bad requests or conflicts

		case pgerrcode.CheckViolation:
			code = codes.InvalidArgument
			errMsg = fmt.Sprintf("check constraint violation: %s", pgErr.ConstraintName)

		case pgerrcode.ExclusionViolation:
			code = codes.InvalidArgument
			errMsg = fmt.Sprintf("exclusion constraint violation: %s", pgErr.Detail)

		case pgerrcode.ForeignKeyViolation:
			code = codes.InvalidArgument
			errMsg = fmt.Sprintf("foreign key violation: %s", pgErr.Detail)

		case pgerrcode.InvalidTextRepresentation: // e.g. enum value does not exist
			code = codes.InvalidArgument
			errMsg = fmt.Sprintf("invalid text: %s", pgErr.Message)

		case pgerrcode.NotNullViolation:
			code = codes.InvalidArgument
			errMsg = fmt.Sprintf("missing required field: %s", pgErr.ColumnName)

		case pgerrcode.StringDataRightTruncationDataException: // e.g. text too long for varchar(i) column
			code = codes.InvalidArgument
			errMsg = pgErr.Message

		case pgerrcode.UndefinedObject: // e.g. enum type does not exist
			code = codes.InvalidArgument
			errMsg = fmt.Sprintf("undefined object: %s", pgErr.Message)

		case pgerrcode.UniqueViolation:
			code = codes.AlreadyExists
			errStr := pgErr.ConstraintName // single column unique key
			if pgErr.Detail != "" {
				errStr = pgErr.Detail // is better, but only filled on multiple column unique keys
			}
			errMsg = fmt.Sprintf("unique constraint violation: %s", errStr)

		default:
			// other pgx error, no special treatment
		}
	}

	// user-facing errors (lysauth.Err*, bad credentials, etc): map HTTP status to gRPC code
	if userErr, ok := errors.AsType[lyserr.User](err); ok {
		code = userErrCode(userErr.StatusCode)
		errMsg = userErr.Message
	}

	// distinguish between error types
	switch code {

	// Expected client errors: return, but don't log
	case codes.AlreadyExists,
		codes.Canceled,
		codes.DeadlineExceeded,
		codes.FailedPrecondition,
		codes.InvalidArgument,
		codes.NotFound,
		codes.PermissionDenied,
		codes.ResourceExhausted,
		codes.Unauthenticated:
		return resp, status.Error(code, errMsg)

	default:
		// Unexpected server errors: log and return generic error to the client.
		rpcApp.Logger.Error("gRPC method failed",
			"method", info.FullMethod,
			"code", code.String(),
			"error", err,
		)
		return resp, status.Error(codes.Internal, "internal server error")
	}
}

func (rpcApp *gRpcServerApplication) runRateLimitCleanup(ctx context.Context, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rpcApp.AuthedRateLimits.CleanupExpired()
			rpcApp.UnauthedRateLimits.CleanupExpired()
		case <-ctx.Done():
			return
		}
	}
}

// userErrCode maps HTTP status codes to gRPC codes.
func userErrCode(httpStatus int) codes.Code {
	switch httpStatus {
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	default: // 400 and any other user error
		return codes.InvalidArgument
	}
}
