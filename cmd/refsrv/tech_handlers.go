package main

import (
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
)

func (srvApp *httpServerApplication) techHubStatus(w http.ResponseWriter, r *http.Request) {
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
