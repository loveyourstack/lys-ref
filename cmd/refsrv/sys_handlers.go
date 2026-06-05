package main

import (
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
)

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
