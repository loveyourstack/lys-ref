package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyserr"
	"golang.org/x/crypto/bcrypt"
)

type loginResponse struct {
	ForcePasswordChange bool       `json:"force_password_change"`
	GeoIpCountryIsoCode string     `json:"geo_ip_country_iso_code"`
	GeoIpLocation       string     `json:"geo_ip_location"`
	Ip                  netip.Addr `json:"ip"`
	Roles               []string   `json:"roles"`
	SessionToken        string     `json:"token"`
	UserId              int64      `json:"user_id"`
	UserName            string     `json:"user_name"`
}

func (srvApp *httpServerApplication) authBlockSessionIp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get ip param
	ipStr := mux.Vars(r)["ip"]
	if ipStr == "" {
		lys.HandleUserError(lyserr.User{Message: "ip is missing"}, w)
		return
	}

	// parse ip param
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		lys.HandleUserError(lyserr.User{Message: "invalid ip"}, w)
		return
	}

	// add to app blocked IPs
	srvApp.LoginAttempts.Block(ip)

	// persist blocked IP to db
	_, err = srvApp.BlockedIPStore.Insert(ctx, sysblockedip.Input{Ip: ip})
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("authBlockSessionIp: srvApp.BlockedIPStore.Insert failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// delete any existing sessions for this IP
	err = srvApp.Sessions.DeleteByIp(ip)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("authBlockSessionIp: srvApp.Sessions.DeleteByIp failed for IP %s: %w", ip, err), srvApp.ErrorLog, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   "success",
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

// authGetLoginAttempts returns the current app login attempts.
func (srvApp *httpServerApplication) authGetLoginAttempts(w http.ResponseWriter, r *http.Request) {

	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   srvApp.LoginAttempts.ListByCreatedAt(false),
		GetMetadata: &lys.GetMetadata{
			Count:      srvApp.LoginAttempts.Count(),
			TotalCount: int64(srvApp.LoginAttempts.Count()),
		},
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

// authGetSessions returns the current app sessions.
func (srvApp *httpServerApplication) authGetSessions(w http.ResponseWriter, r *http.Request) {

	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   srvApp.Sessions.ListByLastAccessAt(false),
		GetMetadata: &lys.GetMetadata{
			Count:      srvApp.Sessions.Count(),
			TotalCount: int64(srvApp.Sessions.Count()),
		},
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func authGetSignInResponse(forcePasswordChange bool, roles []string, userId int64, geoIpCountryIsoCode, geoIpLocation string, ip netip.Addr, sessToken, userName string) loginResponse {
	loginResp := loginResponse{
		ForcePasswordChange: forcePasswordChange,
		GeoIpCountryIsoCode: geoIpCountryIsoCode,
		GeoIpLocation:       geoIpLocation,
		Ip:                  ip,
		Roles:               roles,
		SessionToken:        sessToken,
		UserId:              userId,
		UserName:            userName,
	}
	return loginResp
}

// authLogin handles user/password login.
func (srvApp *httpServerApplication) authLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// track login attempt and get remote IP
	remoteHostIP, ok := srvApp.authTrackLoginAttempt(ctx, w, r, "authLogin")
	if !ok {
		return
	}

	// get login credentials from request
	creds, err := srvApp.getCredentialsFromRequest(r)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("authLogin: getCredentialsFromRequest failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// try to select a user record for the supplied username
	sysUser, err := srvApp.UserStore.SelectByName(ctx, creds.UserName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			lys.HandleUserError(lyserr.User{Message: "invalid credentials", StatusCode: http.StatusForbidden}, w)
		} else {
			lys.HandleError(ctx, fmt.Errorf("authLogin: userStore.SelectByName failed for username %v: %w", creds.UserName, err), srvApp.ErrorLog, w)
		}
		return
	}

	// reject deactivated user
	if sysUser.IsDeactivated {
		lys.HandleUserError(lyserr.User{Message: "deactivated", StatusCode: http.StatusForbidden}, w)
		return
	}

	// validate password
	err = bcrypt.CompareHashAndPassword([]byte(sysUser.HashedPw), []byte(creds.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			lys.HandleUserError(lyserr.User{Message: "invalid credentials", StatusCode: http.StatusForbidden}, w)
		} else {
			lys.HandleInternalError(ctx, fmt.Errorf("authLogin: bcrypt.CompareHashAndPassword failed for username %s: %w", creds.UserName, err), srvApp.ErrorLog, w)
		}
		return
	}

	// success

	// delete login attempt for the user's IP
	found, err := srvApp.LoginAttempts.DeleteByIp(remoteHostIP)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("authLogin: srvApp.LoginAttempts.DeleteByIp failed for username %v: IP %s: %w", creds.UserName, remoteHostIP, err), srvApp.ErrorLog, w)
		return
	}
	if !found {
		lys.HandleInternalError(ctx, fmt.Errorf("authLogin: srvApp.LoginAttempts.DeleteByIp: IP not found for username %v: IP %s", creds.UserName, remoteHostIP), srvApp.ErrorLog, w)
		return
	}

	// if the user doesn't allow multiple sessions, any existing sessions for this user will be deleted in the Add method below, so archive them first
	if !sysUser.AllowMultipleSessions {
		existingSessions := srvApp.Sessions.GetByUserId(sysUser.Id)
		if len(existingSessions) > 0 {
			err = srvApp.archiveSessions(ctx, existingSessions)
			if err != nil {
				lys.HandleInternalError(ctx, fmt.Errorf("authLogin: srvApp.archiveSessions failed for username %v: %w", creds.UserName, err), srvApp.ErrorLog, w)
				return
			}
		}
	}

	// try to get geo info for the user's IP
	geoIpLocation, geoIpCountryIsoCode, err := srvApp.authGetGeoIp(ctx, remoteHostIP.String())
	if err != nil {
		lys.HandleInternalError(ctx, fmt.Errorf("authLogin: srvApp.authGetGeoIp failed for username %v: %w", creds.UserName, err), srvApp.ErrorLog, w)
		return
	}

	// create a session for this user
	sessInput := lysauth.SessionInput{
		AllowMultipleSessions: sysUser.AllowMultipleSessions,
		FamilyName:            sysUser.FamilyName,
		ForcePasswordChange:   sysUser.ForcePasswordChange,
		GivenName:             sysUser.GivenName,
		GeoIpCountryIsoCode:   geoIpCountryIsoCode,
		GeoIpLocation:         geoIpLocation,
		Ip:                    remoteHostIP,
		Roles:                 sysrole.ToStringSlice(sysUser.Roles),
		UserAgent:             r.UserAgent(),
		UserId:                sysUser.Id,
		UserName:              sysUser.Name,
	}
	sessToken, err := srvApp.Sessions.Add(sessInput)
	if err != nil {
		lys.HandleInternalError(ctx, fmt.Errorf("authLogin: srvApp.Sessions.Add failed for username %v: %w", creds.UserName, err), srvApp.ErrorLog, w)
		return
	}

	// build the reponse to be returned to the user
	loginResp := authGetSignInResponse(sysUser.ForcePasswordChange, sysrole.ToStringSlice(sysUser.Roles), sysUser.Id,
		geoIpCountryIsoCode, geoIpLocation, remoteHostIP, sessToken, sysUser.Name)

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   loginResp,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

// authLogout handles user logout by deleting and archiving the session.
func (srvApp *httpServerApplication) authLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get session
	sess, err := srvApp.Sessions.FromRequest(r, srvApp.InfoLog)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("authLogout: srvApp.Sessions.FromRequest failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// archive session to db
	err = srvApp.archiveSessions(ctx, []lysauth.Session{sess})
	if err != nil {
		lys.HandleInternalError(ctx, fmt.Errorf("authLogout: srvApp.archiveSessions failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// delete session from app
	srvApp.Sessions.DeleteByTokens([]string{sess.Token})

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   "success",
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

// authSessionTokenLogin allows the user to log in via his existing session if it is shown to come from the
// same IP and user agent as the session stored in the application.
// Intended to make a full reload (F5) work in the browser without needing to authenticate again.
func (srvApp *httpServerApplication) authSessionTokenLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// track login attempt and get remote IP
	remoteHostIP, ok := srvApp.authTrackLoginAttempt(ctx, w, r, "authSessionTokenLogin")
	if !ok {
		return
	}

	// get session associated with this request, if any
	sess, err := srvApp.Sessions.FromRequest(r, srvApp.InfoLog)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("authSessionTokenLogin: srvApp.Sessions.FromRequest failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// delete login attempt for the user's IP
	found, err := srvApp.LoginAttempts.DeleteByIp(remoteHostIP)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("authSessionTokenLogin: srvApp.LoginAttempts.DeleteByIp failed for username %v: IP %s: %w", sess.UserName, remoteHostIP, err), srvApp.ErrorLog, w)
		return
	}
	if !found {
		lys.HandleInternalError(ctx, fmt.Errorf("authSessionTokenLogin: srvApp.LoginAttempts.DeleteByIp: IP not found for username %v: IP %s", sess.UserName, remoteHostIP), srvApp.ErrorLog, w)
		return
	}

	// build the reponse to be returned to the UI
	signInResp := authGetSignInResponse(sess.ForcePasswordChange, sess.Roles, sess.UserId,
		sess.GeoIpCountryIsoCode, sess.GeoIpLocation, sess.Ip, sess.Token, sess.UserName)

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   signInResp,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

// authTrackLoginAttempt gets the remote IP, logs the attempt, and records a login attempt.
func (srvApp *httpServerApplication) authTrackLoginAttempt(ctx context.Context, w http.ResponseWriter, r *http.Request, caller string) (remoteHostIP netip.Addr, ok bool) {

	// get remote ip
	remoteHostIP, err := lysauth.GetRemoteHostIP(r, srvApp.UseXForwardedFor, srvApp.XForwardedForIdx)
	if err != nil {
		lys.HandleInternalError(ctx, fmt.Errorf("%s: lysauth.GetRemoteHostIP failed: %w", caller, err), srvApp.ErrorLog, w)
		return netip.Addr{}, false
	}

	// add login attempt from this IP
	err = srvApp.LoginAttempts.Add(remoteHostIP)
	if err != nil {

		// reject if blocked (should be caught by the rejectBlockedIp middleware, but just in case)
		if errors.Is(err, lysauth.ErrBlocked) {
			lys.HandleUserError(lysauth.ErrBlocked, w)
			srvApp.InfoLog.Debug(fmt.Sprintf("authTrackLoginAttempt: blocked request from IP %s - %s %s %s", remoteHostIP, r.Proto, r.Method, r.URL.RequestURI()))
			return netip.Addr{}, false
		}

		// if max attempts exceeded
		if errors.Is(err, lysauth.ErrMaxAttemptsExceeded) {

			// reject
			lys.HandleUserError(lysauth.ErrMaxAttemptsExceeded, w)
			srvApp.InfoLog.Debug(fmt.Sprintf("authTrackLoginAttempt: max attempts exceeded for IP %s - %s %s %s", remoteHostIP, r.Proto, r.Method, r.URL.RequestURI()))

			// persist blocked IP to db
			_, err := srvApp.BlockedIPStore.Insert(ctx, sysblockedip.Input{Ip: remoteHostIP})
			if err != nil {
				lys.HandleError(ctx, fmt.Errorf("%s: srvApp.BlockedIPStore.Insert failed: %w", caller, err), srvApp.ErrorLog, w)
			}
			return netip.Addr{}, false
		}

		// other error: shouldn't happen
		lys.HandleError(ctx, fmt.Errorf("%s: srvApp.LoginAttempts.Add failed: %w", caller, err), srvApp.ErrorLog, w)
		return netip.Addr{}, false
	}

	return remoteHostIP, true
}

func (srvApp *httpServerApplication) authUnblockIp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get ip param
	ipStr := mux.Vars(r)["ip"]
	if ipStr == "" {
		lys.HandleUserError(lyserr.User{Message: "ip is missing"}, w)
		return
	}

	// validate ip param
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		lys.HandleUserError(lyserr.User{Message: "invalid ip"}, w)
		return
	}

	// remove from login attempts
	_, err = srvApp.LoginAttempts.DeleteByIp(ip)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("authUnblockIp: srvApp.LoginAttempts.DeleteByIp failed for IP %s: %w", ip, err), srvApp.ErrorLog, w)
		return
	}

	// delete from db blocked IPs
	err = srvApp.BlockedIPStore.DeleteByIp(ctx, ip)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("authUnblockIp: srvApp.BlockedIPStore.DeleteByIp failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   "success",
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}
