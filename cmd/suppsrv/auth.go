package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppemployee"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspgdb"
)

const CompanyCtxKey lyspgdb.ContextKey = "CompanyKey"

// authenticate is middleware that authenticates the user and adds his information to request context
func (srvApp *httpServerApplication) authenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get employee from email in Employee-Email header
		employee, err := srvApp.getEmployeeFromReq(r)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("authenticate: getEmployeeFromReq failed: %w", err), srvApp.Logger, w)
			return
		}

		// add the employee info to the request context
		reqUserInfo := ReqUserInfo{
			UserId:   employee.Id,
			UserName: employee.GivenName + " " + employee.FamilyName,
		}
		ctx = context.WithValue(ctx, lys.UserInfoCtxKey, reqUserInfo)

		// add the company id under a separate ctx key for use by lyspgdb.GetPoolWithCtxSetting (for RLS)
		ctx = context.WithValue(ctx, CompanyCtxKey, employee.CompanyFk)

		// continue
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getEmployeeFromReq returns the employee from the request
// it expects the email address in the Employee-Email header (for testing only)
func (srvApp *httpServerApplication) getEmployeeFromReq(r *http.Request) (emp suppemployee.Model, err error) {
	ctx := r.Context()

	authHeaderFound := false

	for name, vals := range r.Header {

		if name != "Employee-Email" {
			continue
		}

		authHeaderFound = true
		email := vals[0]
		if email == "" {
			return suppemployee.Model{}, lyserr.User{Message: "Email missing from Employee-Email header", StatusCode: http.StatusForbidden}
		}
		empStore := suppemployee.Store{Db: srvApp.OwnerDb}
		emp, err = empStore.SelectByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return suppemployee.Model{}, lyserr.User{Message: "Invalid email", StatusCode: http.StatusForbidden}
			}
			return suppemployee.Model{}, fmt.Errorf("empStore.SelectByEmail failed: %w", err)
		}
		break
	}

	if !authHeaderFound {
		return suppemployee.Model{}, lyserr.User{Message: "Employee-Email header missing from request", StatusCode: http.StatusForbidden}
	}

	return emp, nil
}

func (srvApp *httpServerApplication) loadBlockedIPsFromDb(ctx context.Context) (err error) {

	// read blocked IPs from db
	dbBlockedIPs, err := srvApp.BlockedIPStore.SelectAll(ctx)
	if err != nil {
		return fmt.Errorf("loadBlockedIPsFromDb: BlockedIPStore.SelectAll failed: %w", err)
	}

	if len(dbBlockedIPs) == 0 {
		return nil
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

	// no other login attempts: this server has no user login process

	// load login attempts into app
	err = srvApp.LoginAttempts.Load(appLAs)
	if err != nil {
		return fmt.Errorf("loadBlockedIPsFromDb: LoginAttempts.Load failed: %w", err)
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
			lys.HandleInternalError(ctx, fmt.Errorf("rejectBlockedIp: lysauth.GetRemoteHostIP failed: %w", err), srvApp.Logger, w)
			return
		}

		isBlocked, err := srvApp.LoginAttempts.IsBlocked(remoteHostIP)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("rejectBlockedIp: srvApp.LoginAttempts.IsBlocked failed for IP %s: %w", remoteHostIP, err), srvApp.Logger, w)
			return
		}

		// reject if blocked
		if isBlocked {
			lys.HandleUserError(lysauth.ErrBlocked, w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
