package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssession"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lysset"
	"github.com/loveyourstack/lys/lysslice"
)

// LoginInput contains fields needed for login.
type LoginInput struct {
	UserName string `json:"user_name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (srvApp *httpServerApplication) archiveExpiredSessions(ctx context.Context) {
	expiredSessions := srvApp.Sessions.GetExpired()
	if len(expiredSessions) == 0 {
		return
	}

	runCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	err := srvApp.archiveSessions(runCtx, expiredSessions)
	cancel()
	if err != nil {
		srvApp.ErrorLog.Error(fmt.Sprintf("archiveExpiredSessions: archive failed for %d sessions: %v", len(expiredSessions), err))
		return
	}

	tokens := make([]string, len(expiredSessions))
	for i, sess := range expiredSessions {
		tokens[i] = sess.Token
	}
	srvApp.Sessions.DeleteByTokens(tokens)

	srvApp.InfoLog.Debug(fmt.Sprintf("archiveExpiredSessions: archived %d expired sessions", len(expiredSessions)))
}

// archiveSessions saves the supplied appSessions to the session_history table.
func (srvApp *httpServerApplication) archiveSessions(ctx context.Context, appSessions []lysauth.Session) (err error) {

	inputs := appSessionsToInputs(appSessions)

	_, err = srvApp.SessionHistStore.BulkInsert(ctx, inputs)
	if err != nil {
		return fmt.Errorf("archiveSessions: srvApp.SessionHistStore.BulkInsert failed: %w", err)
	}

	return nil
}

// authenticate is middleware that authenticates the user and adds his information to request context
func (srvApp *httpServerApplication) authenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// allow skipping auth in dev env by using a dummy user
		if srvApp.Config.General.Env == appenv.Dev && !srvApp.Config.API.UseAuthentication {
			reqUserInfo := ReqUserInfo{
				Roles:    []sysrole.Enum{sysrole.Tech},
				UserId:   1,
				UserName: "Unauthed Dev",
			}

			// add the dummy user info to the request context, then continue
			ctx = context.WithValue(ctx, lys.UserInfoCtxKey, reqUserInfo)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// auth required

		// get session associated with this request, if any
		sess, err := srvApp.Sessions.FromRequest(r, srvApp.InfoLog)
		if err != nil {
			lys.HandleError(r.Context(), fmt.Errorf("authenticate: srvApp.Sessions.FromRequest failed: %w", err), srvApp.ErrorLog, w)
			return
		}

		// convert session roles from string to enum
		sysRoles, err := sysrole.FromStringSlice(sess.Roles)
		if err != nil {
			lys.HandleInternalError(r.Context(), fmt.Errorf("authenticate: sysrole.FromStringSlice failed: %w", err), srvApp.ErrorLog, w)
			return
		}

		// write fields that need to be attached to the request
		reqUserInfo := ReqUserInfo{
			Roles:    sysRoles,
			UserId:   sess.UserId,
			UserName: sess.UserName,
		}

		// add the user info to the request context, then continue
		ctx = context.WithValue(r.Context(), lys.UserInfoCtxKey, reqUserInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srvApp *httpServerApplication) getCredentialsFromRequest(r *http.Request) (creds LoginInput, err error) {

	// get req body
	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		return LoginInput{}, fmt.Errorf("lys.ExtractJsonBody failed: %w", err)
	}

	// unmarshal the body
	creds, err = lys.DecodeJsonBody[LoginInput](body)
	if err != nil {
		return LoginInput{}, fmt.Errorf("lys.DecodeJsonBody failed: %w", err)
	}

	// validate input
	if err = lysmeta.Validate(srvApp.Validate, creds); err != nil {
		return LoginInput{}, fmt.Errorf("lysmeta.Validate failed: %w", err)
	}

	return creds, nil
}

func (srvApp *httpServerApplication) authGetGeoIp(ctx context.Context, remoteAddr string) (geoIpLocation, geoIpCountryIsoCode string, err error) {

	network, err := srvApp.GeoNetworkStore.SelectByIp(ctx, remoteAddr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "Unknown network", "ZZ", nil
		}
		return "", "", fmt.Errorf("srvApp.GeoNetworkStore.SelectByIp failed: %w", err)
	}

	loc, err := srvApp.GeoLocationStore.SelectByGeonameId(ctx, network.GeonameId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "Unknown location", "ZZ", nil
		}
		return "", "", fmt.Errorf("srvApp.GeoLocationStore.SelectByGeonameId failed: %w", err)
	}

	return fmt.Sprintf("%s, %s", loc.CityName, loc.CountryName), loc.CountryIsoCode, nil
}

func (srvApp *httpServerApplication) loadLoginAttemptsFromDb(ctx context.Context) (err error) {

	// read blocked IPs from db
	dbBlockedIPs, err := srvApp.BlockedIPStore.SelectAll(ctx)
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
	dbLAs, err := srvApp.LoginAttemptStore.SelectAll(ctx)
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
	err = srvApp.LoginAttempts.Load(appLAs)
	if err != nil {
		return fmt.Errorf("loadLoginAttemptsFromDb: LoginAttempts.Load failed: %w", err)
	}

	return nil
}

func (srvApp *httpServerApplication) loadSessionsFromDb(ctx context.Context) (err error) {

	// read sessions from db
	dbSessions, err := srvApp.SessionStore.SelectAll(ctx)
	if err != nil {
		return fmt.Errorf("loadSessionsFromDb: srvApp.SessionStore.SelectAll failed: %w", err)
	}

	if len(dbSessions) == 0 {
		return nil
	}

	// convert to app sessions
	appSessions := make([]lysauth.Session, len(dbSessions))
	for i, dbSess := range dbSessions {
		appSessInput := lysauth.SessionInput{
			AllowMultipleSessions: dbSess.AllowMultipleSessions,
			FamilyName:            dbSess.FamilyName,
			ForcePasswordChange:   dbSess.ForcePasswordChange,
			GivenName:             dbSess.GivenName,
			GeoIpCountryIsoCode:   dbSess.GeoIpCountryIsoCode,
			GeoIpLocation:         dbSess.GeoIpLocation,
			Ip:                    dbSess.Ip,
			Roles:                 dbSess.Roles,
			UserAgent:             dbSess.UserAgent,
			UserId:                dbSess.UserFk,
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
	err = srvApp.Sessions.Load(appSessions)
	if err != nil {
		return fmt.Errorf("loadSessionsFromDb: srvApp.Sessions.Load failed: %w", err)
	}

	return nil
}

// persistLoginAttempts saves the supplied appLoginAttempts to the login attempts table.
func (srvApp *httpServerApplication) persistLoginAttempts(ctx context.Context) (err error) {

	// get all login attempts
	appLoginAttempts := srvApp.LoginAttempts.All()

	if len(appLoginAttempts) == 0 {
		return nil
	}

	// convert to login attempt inputs for bulk insert
	inputs := appLoginAttemptsToInputs(appLoginAttempts)

	// truncate (to ensure empty table and prevent dups) then bulk insert in tx

	tx, err := srvApp.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("persistLoginAttempts: Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	err = srvApp.LoginAttemptStore.TruncateTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("persistLoginAttempts: LoginAttemptStore.TruncateTx failed: %w", err)
	}

	_, err = srvApp.LoginAttemptStore.BulkInsertTx(ctx, tx, inputs)
	if err != nil {
		return fmt.Errorf("persistLoginAttempts: BulkInsertTx failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("persistLoginAttempts: Tx.Commit failed: %w", err)
	}

	return nil
}

// persistSessions saves the supplied appSessions to the session table.
func (srvApp *httpServerApplication) persistSessions(ctx context.Context) (err error) {

	// get all sessions
	appSessions := srvApp.Sessions.All()

	if len(appSessions) == 0 {
		return nil
	}

	// convert to session inputs for bulk insert
	inputs := appSessionsToInputs(appSessions)

	// truncate (to ensure empty table and prevent dups) then bulk insert in tx

	tx, err := srvApp.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("persistSessions: Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	err = srvApp.SessionStore.TruncateTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("persistSessions: SessionStore.TruncateTx failed: %w", err)
	}

	_, err = srvApp.SessionStore.BulkInsertTx(ctx, tx, inputs)
	if err != nil {
		return fmt.Errorf("persistSessions: BulkInsertTx failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("persistSessions: Tx.Commit failed: %w", err)
	}

	return nil
}

// rejectBlockedIp is middleware that rejects requests from blocked IPs
func (srvApp *httpServerApplication) rejectBlockedIp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get remote ip
		remoteHostIP, err := lysauth.GetRemoteHostIP(r, srvApp.UseXForwardedFor, srvApp.XForwardedForIdx)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("rejectBlockedIp: lysauth.GetRemoteHostIP failed: %w", err), srvApp.ErrorLog, w)
			return
		}

		isBlocked, err := srvApp.LoginAttempts.IsBlocked(remoteHostIP)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("rejectBlockedIp: srvApp.LoginAttempts.IsBlocked failed for IP %s: %w", remoteHostIP, err), srvApp.ErrorLog, w)
			return
		}

		// reject if blocked
		if isBlocked {
			lys.HandleUserError(lysauth.ErrBlocked, w)
			srvApp.InfoLog.Debug(fmt.Sprintf("rejectBlockedIp: blocked request from IP %s - %s %s %s", remoteHostIP, r.Proto, r.Method, r.URL.RequestURI()))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (srvApp *httpServerApplication) runExpiredSessionArchiver(ctx context.Context, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			srvApp.archiveExpiredSessions(ctx)
		}
	}
}

func appLoginAttemptsToInputs(appLAs []lysauth.LoginAttempt) []sysloginattempt.Input {
	inputs := make([]sysloginattempt.Input, len(appLAs))
	for i, appAttempt := range appLAs {
		inputs[i] = sysloginattempt.Input{
			CreatedAt:   appAttempt.CreatedAt,
			Ip:          appAttempt.Ip,
			IsBlocked:   appAttempt.IsBlocked,
			NumAttempts: appAttempt.NumAttempts,
		}
	}
	return inputs
}

func appSessionsToInputs(appSessions []lysauth.Session) []syssession.Input {
	inputs := make([]syssession.Input, len(appSessions))
	for i, appSess := range appSessions {
		inputs[i] = syssession.Input{
			AllowMultipleSessions: appSess.AllowMultipleSessions,
			CreatedAt:             appSess.CreatedAt,
			Email:                 "", // not using in this project
			ExpiresAt:             appSess.ExpiresAt,
			FamilyName:            "", // not using in this project
			GivenName:             "", // not using in this project
			GeoIpCountryIsoCode:   appSess.GeoIpCountryIsoCode,
			GeoIpLocation:         appSess.GeoIpLocation,
			Ip:                    appSess.Ip,
			LastAccessAt:          appSess.LastAccessAt,
			Roles:                 appSess.Roles,
			Token:                 appSess.Token,
			UserAgent:             appSess.UserAgent,
			UserFk:                appSess.UserId,
			UserName:              appSess.UserName,
		}
	}
	return inputs
}

// authorizeRole is middleware that checks if the user has at least one of the allowed roles
func authorizeRole(allowedRoles []sysrole.Enum) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// get the user info from context
			userInfo, ok := ctx.Value(lys.UserInfoCtxKey).(ReqUserInfo)
			if !ok {
				lys.HandleUserError(lys.ErrUserInfoMissing, w)
				return
			}

			// check user is authorized to do this
			if len(allowedRoles) > 0 && !lysslice.ContainsAny(userInfo.Roles, allowedRoles) {
				lys.HandleUserError(lys.ErrPermissionDenied, w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
