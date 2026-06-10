package main

import (
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys/lyserr"
)

func (srvApp *httpServerApplication) awsUpdateUserSecurityGroupRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get session associated with this request, if any, and validate it
	sess, err := srvApp.Sessions.FromRequest(r, srvApp.InfoLog)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.Sessions.FromRequest failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// user IP is only correct in prod, so exit if not in prod
	if srvApp.Config.General.Env != appenv.Prod {
		lys.HandleUserError(lyserr.User{Message: "only call this endpoint in prod env"}, w)
		return
	}

	// for dev testing: comment out the prod env check, and comment in the IP hardcoding
	// don't forget to change it back
	//sess.Ip, _ = netip.ParseAddr("79.224.139.176")

	// update firewall rules for this user
	err = srvApp.AwsSvc.UpdateUserSecurityGroupRules(ctx, sess.UserName, sess.Ip)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.AwsSvc.UpdateUserSecurityGroupRules failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// send mail to user
	// TODO

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   "updated",
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}
