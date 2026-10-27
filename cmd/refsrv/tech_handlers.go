package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/pb"
	"github.com/loveyourstack/lys/lysapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// techHandleRpcError translates a failed RPC call into an HTTP response.
// RPC failures are never passed through as 401/403: the UI would misread them as an expired user session.
func (srvApp *httpServerApplication) techHandleRpcError(ctx context.Context, err error, w http.ResponseWriter) {

	// client went away: nothing to send
	if errors.Is(ctx.Err(), context.Canceled) {
		return
	}

	switch status.Code(err) {
	case codes.DeadlineExceeded:
		lys.HandleExtError(ctx, "RPC server timed out", err, srvApp.Logger, w)
	case codes.Unauthenticated, codes.PermissionDenied:
		// misconfigured or expired service account key: an operator problem, not a user problem
		lys.HandleExtError(ctx, "RPC server rejected the service credentials", err, srvApp.Logger, w)
	case codes.Unavailable:
		lys.HandleExtError(ctx, "RPC server is unavailable", err, srvApp.Logger, w)
	default:
		lys.HandleExtError(ctx, "RPC server error", err, srvApp.Logger, w)
	}
}

func (srvApp *httpServerApplication) techHubStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get userId -> conns count map
	statusMap := srvApp.NotificationHub.Status()

	// get userId -> name map
	userIdNameMap, err := srvApp.UserStore.SelectIdNameMap(ctx)
	if err != nil {
		lys.HandleInternalError(ctx, fmt.Errorf("wsHubStatus: srvApp.UserStore.SelectIdNameMap failed: %w", err), srvApp.Logger, w)
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

// techRpcItemsResponse converts a slice of RPC response items to JSON and writes them to the HTTP response. It also includes metadata if provided.
func techRpcItemsResponse[T proto.Message](ctx context.Context, items []T, metadata *pb.GetMetadata, logger *slog.Logger, w http.ResponseWriter) {

	// convert to JSON raw messages so that timestamps are encoded
	jsonItems := make([]json.RawMessage, len(items))
	for i, item := range items {
		data, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(item)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("techRpcItemsResponse: protojson.Marshal item failed: %w", err), logger, w)
			return
		}
		jsonItems[i] = data
	}

	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   jsonItems,
	}
	if metadata != nil {
		resp.GetMetadata = &lysapi.SelectMetadata{
			Count:                 int(metadata.Count),
			TotalCount:            metadata.TotalCount,
			TotalCountIsEstimated: metadata.TotalCountIsEstimated,
		}
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) techRpcLoginAttempts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rpcResp, err := pb.NewAuthSvcClient(srvApp.rpcClient).ListLoginAttempts(ctx, &pb.EmptyReq{})
	if err != nil {
		srvApp.techHandleRpcError(ctx, fmt.Errorf("techRpcLoginAttempts: AuthSvc.ListLoginAttempts failed: %w", err), w)
		return
	}

	techRpcItemsResponse(ctx, rpcResp.GetData(), rpcResp.GetMetadata(), srvApp.Logger, w)
}

func (srvApp *httpServerApplication) techRpcSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rpcResp, err := pb.NewAuthSvcClient(srvApp.rpcClient).ListSessions(ctx, &pb.EmptyReq{})
	if err != nil {
		srvApp.techHandleRpcError(ctx, fmt.Errorf("techRpcSessions: AuthSvc.ListSessions failed: %w", err), w)
		return
	}

	techRpcItemsResponse(ctx, rpcResp.GetData(), rpcResp.GetMetadata(), srvApp.Logger, w)
}
