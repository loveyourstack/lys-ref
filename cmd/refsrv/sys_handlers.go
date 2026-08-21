package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysnotification"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysformfile"
)

func (srvApp *httpServerApplication) sysAddFakeNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get user from request
	userInfo, ok := ctx.Value(lys.UserInfoCtxKey).(ReqUserInfo)
	if !ok {
		lys.HandleInternalError(ctx, fmt.Errorf("sysAddFakeNotification: user not authenticated"), srvApp.Logger, w)
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
		lys.HandleInternalError(ctx, fmt.Errorf("sysAddFakeNotification: sysNotStore.Insert failed: %w", err), srvApp.Logger, w)
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

	uiStoreData, err := srvApp.SysSvc.SelectUiStoreData(ctx, srvApp.Db)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.SysSvc.SelectUiStoreData failed: %w", err), srvApp.Logger, w)
		return
	}

	// return success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   uiStoreData,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) sysGetUserUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	notsStore := sysnotification.Store{Db: srvApp.Db}
	cnt, err := notsStore.SelectUserUnreadCount(ctx)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("notsStore.SelectUserUnreadCount failed: %w", err), srvApp.Logger, w)
		return
	}

	// return success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   cnt,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) sysSetAllNotificationsToRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// set all notifications to read for the authenticated user
	notsStore := sysnotification.Store{Db: srvApp.Db}
	err := notsStore.SetAllUsersToRead(ctx)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("notsStore.SetAllUsersToRead failed: %w", err), srvApp.Logger, w)
		return
	}

	// return success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   "updated",
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) sysSetNotificationsToRead(env lys.Env) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		type input struct {
			Ids []int64 `json:"ids"`
		}

		// get req body
		body, err := lys.ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("sysSetNotificationsToRead: ExtractJsonBody failed: %w", err), env.Logger, w)
			return
		}

		// unmarshal the body
		var inp input
		err = json.Unmarshal(body, &inp)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("sysSetNotificationsToRead: json.Unmarshal failed: %w", err), env.Logger, w)
			return
		}

		// set these ids to read
		notsStore := sysnotification.Store{Db: srvApp.Db}
		err = notsStore.SetUsersToRead(ctx, inp.Ids)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("notsStore.SetUsersToRead failed: %w", err), srvApp.Logger, w)
			return
		}

		// return success
		resp := lys.StdResponse{
			Status: lys.ReqSucceeded,
			Data:   "updated",
		}
		lys.JsonResponse(resp, http.StatusOK, w)
	}
}

func (srvApp *httpServerApplication) sysSetUserProfilePic(env lys.Env) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// extract the image from the request
		uploadFiles, err := lysformfile.ExtractFromRequest(r, lysformfile.ExtractParams{
			AllowedMimeTypes: lysformfile.ImageMimeTypes,
			ImgMaxHeightPx:   new(400),
			ImgMaxWidthPx:    new(400),
			ImgMinHeightPx:   new(400),
			ImgMinWidthPx:    new(400),
			MaxSizePerFile:   500 * 1024, // 500 KB
		})
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("sysSetUserProfilePic: lysformfile.ExtractFromRequest failed: %w", err), env.Logger, w)
			return
		}

		if len(uploadFiles) != 1 {
			lys.HandleUserError(lyserr.User{Message: "exactly one image must be uploaded"}, w)
			return
		}

		// get user id from context
		userId := lys.GetUserIdFromCtx(ctx)
		if userId == 0 {
			lys.HandleUserError(lyserr.User{Message: "user not authenticated", StatusCode: http.StatusForbidden}, w)
			return
		}

		// store image and update user record
		storedFileName, err := srvApp.SysSvc.SetUserProfilePic(ctx, userId, uploadFiles[0], srvApp.Config.General.UploadsPath)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("srvApp.SysSvc.SetUserProfilePic failed: %w", err), env.Logger, w)
			return
		}

		// update user session(s) with new profile pic
		srvApp.Sessions.UpdateProfilePicByUserId(userId, storedFileName)

		// return success
		resp := lys.StdResponse{
			Status: lys.ReqSucceeded,
			Data:   storedFileName,
		}
		lys.JsonResponse(resp, http.StatusOK, w)
	}
}
