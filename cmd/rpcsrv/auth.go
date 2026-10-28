package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsession"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysset"
	"github.com/loveyourstack/lys/lysslice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (rpcApp *gRpcServerApplication) archiveExpiredSessions(ctx context.Context) {
	expiredSessions := rpcApp.Sessions.GetExpired()
	if len(expiredSessions) == 0 {
		return
	}

	runCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	err := rpcApp.archiveSessions(runCtx, expiredSessions)
	cancel()
	if err != nil {
		rpcApp.Logger.Error(fmt.Sprintf("archiveExpiredSessions: archive failed for %d sessions: %v", len(expiredSessions), err))
		return
	}

	tokens := make([]string, len(expiredSessions))
	for i, sess := range expiredSessions {
		tokens[i] = sess.Token
	}
	rpcApp.Sessions.DeleteByTokens(tokens)

	rpcApp.Logger.Debug(fmt.Sprintf("archiveExpiredSessions: archived %d expired sessions", len(expiredSessions)))
}

// archiveSessions saves the supplied appSessions to the session_history table.
func (rpcApp *gRpcServerApplication) archiveSessions(ctx context.Context, appSessions []lysauth.Session) (err error) {

	inputs := appSessionsToInputs(appSessions)

	_, err = rpcApp.SessionHistStore.BulkInsert(ctx, inputs)
	if err != nil {
		return fmt.Errorf("archiveSessions: rpcApp.SessionHistStore.BulkInsert failed: %w", err)
	}

	return nil
}

// authenticate is a unary interceptor that authenticates the caller and adds their information to the context
func (rpcApp *gRpcServerApplication) authenticate(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {

	// get method policy
	policy, ok := methodPolicies[info.FullMethod]
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "RPC method is not authorized")
	}

	// continue without auth if the method is public
	if policy.public {
		return handler(ctx, req)
	}

	var sess lysauth.Session

	// use a dummy user if auth is skipped in dev env
	if rpcApp.Config.General.Env == appenv.Dev && !rpcApp.Config.RpcApi.UseAuthentication {

		sess = lysauth.Session{
			Roles:    []string{"Tech"},
			UserId:   1,
			UserName: "Unauthed Dev",
		}
		// don't continue immediately: allow method policy to be enforced below

	} else {

		// auth required

		// get session associated with this request, if any
		var err error
		sess, err = rpcApp.authGetSession(ctx)
		if err != nil {
			if userErr, ok := errors.AsType[lyserr.User](err); ok {
				return nil, status.Error(codes.Unauthenticated, userErr.Message)
			}
			if errors.Is(err, lysauth.ErrBlocked) {
				return nil, status.Error(codes.PermissionDenied, "IP blocked")
			}
			if errors.Is(err, lysauth.ErrMaxAttemptsExceeded) {
				return nil, status.Error(codes.PermissionDenied, "Max login attempts exceeded")
			}
			return nil, fmt.Errorf("authenticate: rpcApp.authGetSession failed: %w", err)
		}

	}

	// convert session roles from string to enum
	sysRoles, err := sysrole.FromStringSlice(sess.Roles)
	if err != nil {
		return nil, fmt.Errorf("authenticate: sysrole.FromStringSlice failed: %w", err)
	}

	// per-method role check (gRPC equivalent of authorizeRole)
	if len(policy.roles) > 0 && !lysslice.ContainsAny(sysRoles, policy.roles) {
		return nil, status.Error(codes.PermissionDenied, "insufficient role")
	}

	// update the user info in the context (defined in logRequest) with the authenticated session details
	if ui, ok := ctx.Value(lys.UserInfoCtxKey).(*ReqUserInfo); ok {
		ui.Roles = sysRoles
		ui.UserId = sess.UserId
		ui.UserName = sess.UserName
	}

	return handler(ctx, req)
}

func (rpcApp *gRpcServerApplication) authGetGeoIp(ctx context.Context, remoteAddr string) (geoIpLocation, geoIpCountryIsoCode string, err error) {

	network, err := rpcApp.GeoNetworkStore.SelectByIp(ctx, remoteAddr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "Unknown network", "ZZ", nil
		}
		return "", "", fmt.Errorf("rpcApp.GeoNetworkStore.SelectByIp failed: %w", err)
	}

	loc, err := rpcApp.GeoLocationStore.SelectByGeonameId(ctx, network.GeonameId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "Unknown location", "ZZ", nil
		}
		return "", "", fmt.Errorf("rpcApp.GeoLocationStore.SelectByGeonameId failed: %w", err)
	}

	return fmt.Sprintf("%s, %s", loc.CityName, loc.CountryName), loc.CountryIsoCode, nil
}

func (rpcApp *gRpcServerApplication) authGetServiceAccountSession(ctx context.Context, apiKey string, remoteHostIp netip.Addr) (sess lysauth.Session, err error) {

	// note: checking for blocked IP is done by rejectBlockedIp middleware

	// hash API key using SHA-256
	hashedApiKey := sha256.Sum256([]byte(apiKey))
	hash := hex.EncodeToString(hashedApiKey[:])

	// lookup hashed key in DB
	dbSvcAcc, err := rpcApp.SvcAccountStore.SelectByApiKeyHash(ctx, hash)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return lysauth.Session{}, userErr
		}
		if errors.Is(err, pgx.ErrNoRows) {

			// record failed login attempt for this IP and block if necessary
			if tErr := rpcApp.recordFailedAttempt(ctx, remoteHostIp); tErr != nil {
				return lysauth.Session{}, tErr // ErrBlocked / ErrMaxAttemptsExceeded
			}

			return lysauth.Session{}, lyserr.User{Message: "Invalid API key", StatusCode: http.StatusUnauthorized}
		}
		return lysauth.Session{}, fmt.Errorf("authGetServiceAccountSession: SvcAccountStore.SelectByApiKeyHash failed: %w", err)
	}

	// create dummy session for service account
	sess = lysauth.Session{
		UserId:   0,                                     // no user ID
		UserName: "SA: " + dbSvcAcc.Name,                // for logging
		Roles:    sysrole.ToStringSlice(dbSvcAcc.Roles), // for auth
	}

	return sess, nil
}

func (rpcApp *gRpcServerApplication) authGetSession(ctx context.Context) (sess lysauth.Session, err error) {

	// metadata must be present
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return lysauth.Session{}, lyserr.User{Message: "Metadata missing or empty", StatusCode: http.StatusUnauthorized}
	}

	// get ip
	remoteHostIp, err := getRemoteHostIp(ctx, rpcApp.UseXForwardedFor, rpcApp.XForwardedForIdx)
	if err != nil {
		return lysauth.Session{}, fmt.Errorf("authGetSession: getRemoteHostIp failed: %w", err)
	}

	// check for service account access via API key
	apiKey, err := getApiKeyIfPresent(md)
	if err != nil {
		return lysauth.Session{}, fmt.Errorf("authGetSession: getApiKeyIfPresent failed: %w", err)
	}
	if apiKey != "" {
		sess, err = rpcApp.authGetServiceAccountSession(ctx, apiKey, remoteHostIp)
		if err != nil {
			return lysauth.Session{}, fmt.Errorf("authGetSession: authGetServiceAccountSession failed: %w", err)
		}

		// successful service account authentication
		return sess, nil
	}

	// regular user access

	// get token from Authorization header in metadata
	token, err := getBearerToken(md)
	if err != nil {
		return lysauth.Session{}, fmt.Errorf("authGetSession: getBearerToken failed: %w", err)
	}

	// get user agent
	userAgent, err := getUserAgent(ctx)
	if err != nil {
		return lysauth.Session{}, fmt.Errorf("authGetSession: getUserAgent failed: %w", err)
	}

	// get session
	sess, err = rpcApp.Sessions.FromToken(token, remoteHostIp, userAgent, false, rpcApp.Logger)
	if err != nil {
		return lysauth.Session{}, fmt.Errorf("authGetSession: rpcApp.Sessions.FromToken failed: %w", err)
	}

	return sess, nil
}

func (rpcApp *gRpcServerApplication) loadLoginAttemptsFromDb(ctx context.Context) (err error) {

	// read blocked IPs from db
	dbBlockedIPs, err := rpcApp.BlockedIPStore.SelectAll(ctx)
	if err != nil {
		return fmt.Errorf("loadLoginAttemptsFromDb: BlockedIPStore.SelectAll failed: %w", err)
	}

	// make set of blocked IPs for easy lookup
	blockedIPSet := lysset.New[netip.Addr]()
	for _, dbBlockedIP := range dbBlockedIPs {
		blockedIPSet.Add(dbBlockedIP.Ip)
	}

	// create app login attempts from blocked IPs
	appLAs := make([]lysauth.LoginAttempt, len(dbBlockedIPs))
	for i, dbBlockedIP := range dbBlockedIPs {
		appLAs[i] = lysauth.LoginAttempt{
			CreatedAt:   dbBlockedIP.CreatedAt,
			Ip:          dbBlockedIP.Ip,
			IsBlocked:   true,
			NumAttempts: 0, // not currently tracked for blocked IPs
		}
	}

	// read login attempts from db
	dbLAs, err := rpcApp.LoginAttemptStore.SelectAll(ctx)
	if err != nil {
		return fmt.Errorf("loadLoginAttemptsFromDb: LoginAttemptStore.SelectAll failed: %w", err)
	}

	// convert to app login attempts
	for _, dbLA := range dbLAs {

		// skip blocked IPs added above
		if blockedIPSet.Contains(dbLA.Ip) {
			continue
		}

		appLAs = append(appLAs, lysauth.LoginAttempt{
			CreatedAt:   dbLA.CreatedAt,
			Ip:          dbLA.Ip,
			IsBlocked:   dbLA.IsBlocked,
			NumAttempts: dbLA.NumAttempts,
		})
	}

	if len(appLAs) == 0 {
		return nil
	}

	// load login attempts into app
	err = rpcApp.LoginAttempts.Load(appLAs)
	if err != nil {
		return fmt.Errorf("loadLoginAttemptsFromDb: LoginAttempts.Load failed: %w", err)
	}

	return nil
}

func (rpcApp *gRpcServerApplication) loadSessionsFromDb(ctx context.Context) (err error) {

	// read sessions from db
	dbSessions, err := rpcApp.SessionStore.SelectAll(ctx)
	if err != nil {
		return fmt.Errorf("loadSessionsFromDb: rpcApp.SessionStore.SelectAll failed: %w", err)
	}

	if len(dbSessions) == 0 {
		return nil
	}

	// convert to app sessions
	appSessions := make([]lysauth.Session, len(dbSessions))
	for i, dbSess := range dbSessions {
		appSessInput := lysauth.SessionInput{
			AllowMultipleSessions: dbSess.AllowMultipleSessions,
			Email:                 dbSess.Email,
			FamilyName:            dbSess.FamilyName,
			ForcePasswordChange:   dbSess.ForcePasswordChange,
			GivenName:             dbSess.GivenName,
			GeoIpCountryIsoCode:   dbSess.GeoIpCountryIsoCode,
			GeoIpLocation:         dbSess.GeoIpLocation,
			Ip:                    dbSess.Ip,
			ProfilePic:            dbSess.ProfilePic,
			Roles:                 dbSess.Roles,
			UserAgent:             dbSess.UserAgent,
			UserId:                dbSess.UserId,
			UserName:              dbSess.UserName,
		}
		appSessions[i] = lysauth.Session{
			SessionInput: appSessInput,

			CreatedAt:    dbSess.CreatedAt,
			ExpiresAt:    dbSess.ExpiresAt,
			LastAccessAt: dbSess.LastAccessAt,
			Token:        dbSess.Token,
		}
	}

	// load sessions into app
	err = rpcApp.Sessions.Load(appSessions)
	if err != nil {
		return fmt.Errorf("loadSessionsFromDb: rpcApp.Sessions.Load failed: %w", err)
	}

	return nil
}

// persistLoginAttempts saves the supplied appLoginAttempts to the login attempts table.
func (rpcApp *gRpcServerApplication) persistLoginAttempts(ctx context.Context) (err error) {

	// get all login attempts
	appLoginAttempts := rpcApp.LoginAttempts.All()

	// convert to login attempt inputs for bulk insert
	inputs := appLoginAttemptsToInputs(appLoginAttempts)

	// truncate (to ensure empty table and prevent dups) then bulk insert in tx

	tx, err := rpcApp.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("persistLoginAttempts: rpcApp.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	err = rpcApp.LoginAttemptStore.TruncateTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("persistLoginAttempts: rpcApp.LoginAttemptStore.TruncateTx failed: %w", err)
	}

	if len(inputs) > 0 {
		_, err = rpcApp.LoginAttemptStore.BulkInsertTx(ctx, tx, inputs)
		if err != nil {
			return fmt.Errorf("persistLoginAttempts: rpcApp.LoginAttemptStore.BulkInsertTx failed: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("persistLoginAttempts: Tx.Commit failed: %w", err)
	}

	return nil
}

// persistSessions saves the supplied appSessions to the session table.
func (rpcApp *gRpcServerApplication) persistSessions(ctx context.Context) (err error) {

	// get all sessions
	appSessions := rpcApp.Sessions.All()

	// convert to session inputs for bulk insert
	inputs := appSessionsToInputs(appSessions)

	// truncate (to ensure empty table and prevent dups) then bulk insert in tx

	tx, err := rpcApp.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("persistSessions: rpcApp.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	err = rpcApp.SessionStore.TruncateTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("persistSessions: rpcApp.SessionStore.TruncateTx failed: %w", err)
	}

	if len(inputs) > 0 {
		_, err = rpcApp.SessionStore.BulkInsertTx(ctx, tx, inputs)
		if err != nil {
			return fmt.Errorf("persistSessions: rpcApp.SessionStore.BulkInsertTx failed: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("persistSessions: tx.Commit failed: %w", err)
	}

	return nil
}

func (rpcApp *gRpcServerApplication) recordFailedAttempt(ctx context.Context, ip netip.Addr) error {

	err := rpcApp.LoginAttempts.Add(ip)
	if err == nil {
		return nil
	}

	// if max attempts exceeded
	if errors.Is(err, lysauth.ErrMaxAttemptsExceeded) {

		// persist blocked IP to db
		if _, err := rpcApp.BlockedIPStore.Insert(ctx, sysblockedip.Input{Ip: ip}); err != nil {
			return fmt.Errorf("BlockedIPStore.Insert failed: %w", err)
		}

		return lysauth.ErrMaxAttemptsExceeded
	}

	return err // includes ErrBlocked
}

func (rpcApp *gRpcServerApplication) rejectBlockedIp(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {

	// get remote ip
	remoteHostIp, err := getRemoteHostIp(ctx, rpcApp.UseXForwardedFor, rpcApp.XForwardedForIdx)
	if err != nil {
		return nil, fmt.Errorf("getRemoteHostIp failed: %w", err)
	}

	isBlocked, err := rpcApp.LoginAttempts.IsBlocked(remoteHostIp)
	if err != nil {
		return nil, fmt.Errorf("rpcApp.LoginAttempts.IsBlocked failed for IP %s: %w", remoteHostIp, err)
	}

	// reject if blocked
	if isBlocked {
		rpcApp.Logger.Debug(fmt.Sprintf("rejectBlockedIp: blocked request from IP %s", remoteHostIp))
		return nil, status.Error(codes.PermissionDenied, "blocked IP")
	}

	return handler(ctx, req)
}

func (rpcApp *gRpcServerApplication) runExpiredSessionArchiver(ctx context.Context, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rpcApp.archiveExpiredSessions(ctx)
		}
	}
}

func appLoginAttemptsToInputs(appLAs []lysauth.LoginAttempt) []sysrpcloginattempt.Input {
	inputs := make([]sysrpcloginattempt.Input, len(appLAs))
	for i, appAttempt := range appLAs {
		inputs[i] = sysrpcloginattempt.Input{
			CreatedAt:   appAttempt.CreatedAt,
			Ip:          appAttempt.Ip,
			IsBlocked:   appAttempt.IsBlocked,
			NumAttempts: appAttempt.NumAttempts,
		}
	}
	return inputs
}

func appSessionsToInputs(appSessions []lysauth.Session) []sysrpcsession.Input {
	inputs := make([]sysrpcsession.Input, len(appSessions))
	for i, appSess := range appSessions {
		inputs[i] = sysrpcsession.Input{
			AllowMultipleSessions: appSess.AllowMultipleSessions,
			CreatedAt:             appSess.CreatedAt,
			Email:                 appSess.Email,
			ExpiresAt:             appSess.ExpiresAt,
			FamilyName:            appSess.FamilyName,
			GivenName:             appSess.GivenName,
			GeoIpCountryIsoCode:   appSess.GeoIpCountryIsoCode,
			GeoIpLocation:         appSess.GeoIpLocation,
			Ip:                    appSess.Ip,
			LastAccessAt:          appSess.LastAccessAt,
			ProfilePic:            appSess.ProfilePic,
			Roles:                 appSess.Roles,
			Token:                 appSess.Token,
			UserAgent:             appSess.UserAgent,
			UserId:                appSess.UserId,
			UserName:              appSess.UserName,
		}
	}
	return inputs
}
