package main

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysnotification"
	"github.com/loveyourstack/lys/lyserr"
)

func (srvApp *httpServerApplication) sysAddFakeNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get user from request
	userInfo, ok := ctx.Value(lys.UserInfoCtxKey).(ReqUserInfo)
	if !ok {
		lys.HandleInternalError(ctx, fmt.Errorf("sysAddFakeNotification: user not authenticated"), srvApp.ErrorLog, w)
		return
	}

	// get and validate type= param
	notType := r.URL.Query().Get("type")
	if !slices.Contains([]string{"Info", "Warning"}, notType) {
		lys.HandleUserError(lyserr.User{Message: fmt.Sprintf("invalid type param value: %s", notType)}, w)
		return
	}

	msg := ""
	switch notType {
	case "Info":
		msg = "This is an info notification"
	case "Warning":
		msg = "This is a warning notification"
	}

	// add a test notification for this user
	input := sysnotification.Input{
		IsRead:  false,
		Message: msg,
		NotType: notType,
		UserFk:  userInfo.UserId,
	}
	sysNotStore := sysnotification.Store{Db: srvApp.Db}
	if _, err := sysNotStore.Insert(ctx, input); err != nil {
		lys.HandleInternalError(ctx, fmt.Errorf("sysAddFakeNotification: sysNotStore.Insert failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   "added",
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) sysGetUiStoreData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uiStoreData, err := srvApp.SysSvc.GetUiStoreData(ctx, srvApp.Db)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.SysSvc.GetUiStoreData failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// return success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   uiStoreData,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}
