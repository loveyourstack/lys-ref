package main

import (
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
)

func (srvApp *httpServerApplication) wsRegisterForNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get user from request
	userInfo, ok := ctx.Value(lys.UserInfoCtxKey).(ReqUserInfo)
	if !ok {
		lys.HandleInternalError(ctx, fmt.Errorf("wsRegisterForNotifications: user not authenticated"), srvApp.ErrorLog, w)
		return
	}

	// upgrade http connection to websocket
	wsConn, err := srvApp.NotificationHub.UpgradeHttpRequest(w, r)
	if err != nil {
		lys.HandleError(r.Context(), fmt.Errorf("srvApp.NotificationHub.UpgradeHttpRequest failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// register websocket connection for user in notification hub
	err = srvApp.NotificationHub.Register(userInfo.UserId, wsConn)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.NotificationHub.Register failed: %w", err), srvApp.ErrorLog, w)
		return
	}
}

func (srvApp *httpServerApplication) wsUnregisterForNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get user from request
	userInfo, ok := ctx.Value(lys.UserInfoCtxKey).(ReqUserInfo)
	if !ok {
		lys.HandleInternalError(ctx, fmt.Errorf("wsUnregisterForNotifications: user not authenticated"), srvApp.ErrorLog, w)
		return
	}

	// get this user's active ws connections
	wsConns := srvApp.NotificationHub.GetUserConns(userInfo.UserId)

	// unregister user conns
	for _, wsConn := range wsConns {
		srvApp.NotificationHub.Unregister(userInfo.UserId, wsConn)
	}
}
