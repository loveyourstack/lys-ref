package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/stores/lysinc/lysincemployee"
	"github.com/loveyourstack/lys-ref/pkg/lysp"
)

func (srvApp *httpServerApplication) GetEmployeeTree(env lys.Env, selectFunc func(ctx context.Context) (*lysp.TreeNode[lysincemployee.Model], error)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// select items from Db
		items, err := selectFunc(ctx)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("GetEmployeeTree: selectFunc failed: %w", err), env.Logger, w)
			return
		}

		// success
		resp := lys.StdResponse{
			Status: lys.ReqSucceeded,
			Data:   items,
		}
		lys.JsonResponse(resp, http.StatusOK, w)
	}
}
