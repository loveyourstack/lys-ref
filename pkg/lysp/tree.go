package lysp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
)

func GetItem[T any](env lys.Env, selectFunc func(ctx context.Context) (T, error)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// select item from Db
		item, err := selectFunc(ctx)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("GetItem: selectFunc failed: %w", err), env.Logger, w)
			return
		}

		// success
		resp := lys.StdResponse{
			Status: lys.ReqSucceeded,
			Data:   item,
		}
		lys.JsonResponse(resp, http.StatusOK, w)
	}
}
