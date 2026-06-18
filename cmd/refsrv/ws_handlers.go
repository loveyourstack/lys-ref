package main

import (
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
)

func (srvApp *httpServerApplication) wsNotificationsRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get user from request
	userInfo, ok := ctx.Value(lys.UserInfoCtxKey).(ReqUserInfo)
	if !ok {
		lys.HandleInternalError(ctx, fmt.Errorf("wsNotificationsRegister: user not authenticated"), srvApp.ErrorLog, w)
		return
	}

	// register this ws connection to receive notifications for this user
	if err := srvApp.NotificationHub.ServeUserSocket(ctx, w, r, userInfo.UserId); err != nil {
		srvApp.ErrorLog.Error("wsNotificationsRegister: srvApp.NotificationHub.ServeUserSocket failed", "user_id", userInfo.UserId, "error", err)
	}
}
