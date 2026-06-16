package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/loveyourstack/lys"
)

const (
	wsPongWait      = 60 * time.Second
	wsPingInterval  = 54 * time.Second // less than wsPongWait
	wsPingWriteWait = 10 * time.Second
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

	// upgrade http connection to websocket
	wsConn, err := srvApp.NotificationHub.UpgradeHttpRequest(w, r)
	if err != nil {
		lys.HandleError(r.Context(), fmt.Errorf("srvApp.NotificationHub.UpgradeHttpRequest failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// register websocket connection for user in notification hub
	srvApp.NotificationHub.Register(userInfo.UserId, wsConn)

	// stop heartbeat before unregister/close
	done := make(chan struct{})
	defer srvApp.NotificationHub.Unregister(userInfo.UserId, wsConn)
	defer close(done)

	// initial read deadline
	if err := wsConn.SetReadDeadline(time.Now().Add(wsPongWait)); err != nil {
		return
	}

	// each pong extends read deadline
	wsConn.SetPongHandler(func(string) error {
		return wsConn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	// periodic ping heartbeat
	go func() {
		ticker := time.NewTicker(wsPingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				deadline := time.Now().Add(wsPingWriteWait)
				if err := wsConn.WriteControl(websocket.PingMessage, []byte("hb"), deadline); err != nil {
					_ = wsConn.Close() // force read loop to exit -> defer unregister runs
					return
				}
			}
		}
	}()

	// keep connection open until client disconnects, e.g. with a browser page reload
	defer srvApp.NotificationHub.Unregister(userInfo.UserId, wsConn)
	for {
		if _, _, err := wsConn.ReadMessage(); err != nil {
			break
		}
	}
}
