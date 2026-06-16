package main

import (
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
)

func (srvApp *httpServerApplication) wsHubStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get userId -> conns count map
	statusMap := srvApp.NotificationHub.Status()

	// get userId -> name map
	userIdNameMap, err := srvApp.UserStore.SelectIdNameMap(ctx)
	if err != nil {
		lys.HandleInternalError(ctx, fmt.Errorf("wsHubStatus: srvApp.UserStore.SelectIdNameMap failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// convert statusMap userIds to user names
	nameConnsMap := make(map[string]int, len(statusMap))
	for userId, conns := range statusMap {
		if name, ok := userIdNameMap[userId]; ok {
			nameConnsMap[name] = conns
		} else {
			nameConnsMap[fmt.Sprintf("userId %d", userId)] = conns
		}
	}

	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   nameConnsMap,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

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
