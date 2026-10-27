package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/pb"
	"github.com/loveyourstack/lys/lysapi"
	"google.golang.org/protobuf/encoding/protojson"
)

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

func (srvApp *httpServerApplication) techRpcLoginAttempts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rpcResp, err := pb.NewAuthSvcClient(srvApp.rpcClient).ListLoginAttempts(ctx, &pb.EmptyReq{})
	if err != nil {
		lys.HandleInternalError(ctx, fmt.Errorf("techRpcLoginAttempts: AuthSvc.ListLoginAttempts failed: %w", err), srvApp.Logger, w)
		return
	}

	// convert rpc response data to JSON raw messages so that timestamps are encoded
	jsonItems := make([]json.RawMessage, len(rpcResp.GetData()))
	for i, session := range rpcResp.GetData() {
		data, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(session)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("techRpcLoginAttempts: protojson.Marshal session failed: %w", err), srvApp.Logger, w)
			return
		}
		jsonItems[i] = data
	}

	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   jsonItems,
	}
	if metadata := rpcResp.GetMetadata(); metadata != nil {
		resp.GetMetadata = &lysapi.SelectMetadata{
			Count:                 int(metadata.Count),
			TotalCount:            metadata.TotalCount,
			TotalCountIsEstimated: metadata.TotalCountIsEstimated,
		}
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) techRpcSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rpcResp, err := pb.NewAuthSvcClient(srvApp.rpcClient).ListSessions(ctx, &pb.EmptyReq{})
	if err != nil {
		lys.HandleInternalError(ctx, fmt.Errorf("techRpcSessions: AuthSvc.ListSessions failed: %w", err), srvApp.Logger, w)
		return
	}

	jsonItems := make([]json.RawMessage, len(rpcResp.GetData()))
	for i, session := range rpcResp.GetData() {
		data, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(session)
		if err != nil {
			lys.HandleInternalError(ctx, fmt.Errorf("techRpcSessions: protojson.Marshal session failed: %w", err), srvApp.Logger, w)
			return
		}
		jsonItems[i] = data
	}

	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   jsonItems,
	}
	if metadata := rpcResp.GetMetadata(); metadata != nil {
		resp.GetMetadata = &lysapi.SelectMetadata{
			Count:                 int(metadata.Count),
			TotalCount:            metadata.TotalCount,
			TotalCountIsEstimated: metadata.TotalCountIsEstimated,
		}
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}
