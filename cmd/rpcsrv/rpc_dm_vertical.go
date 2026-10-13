package main

import (
	"context"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd/rpcsrv/pb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys-ref/pkg/lysp"
	"google.golang.org/protobuf/types/known/emptypb"
)

type dmVertServer struct {
	pb.UnimplementedDmVerticalSvcServer
	*gRpcServerApplication

	store            dmvertical.Store
	jsonKeyDbNameMap map[string]string
}

func (s *dmVertServer) Create(ctx context.Context, req *pb.DmVerticalInput) (*pb.DmVertical, error) {
	return lysp.RpcCreate(ctx, req, s.store, dmvertical.InputFromPbFunc, dmvertical.ModelToPbFunc)
}

func (s *dmVertServer) Delete(ctx context.Context, req *pb.DeleteByIdReq) (*emptypb.Empty, error) {
	return lysp.RpcDelete(ctx, req.Id, s.store)
}

func (s *dmVertServer) Get(ctx context.Context, req *pb.GetByIdReq) (*pb.DmVertical, error) {
	return lysp.RpcGetById(ctx, req.Id, s.store, dmvertical.ModelToPbFunc)
}

func (s *dmVertServer) List(ctx context.Context, req *pb.ListReq) (*pb.ListDmVerticalsResp, error) {
	return lysp.RpcList(ctx, s.store, req, s.GetOptions, dmvertical.ModelToPbFunc,
		func(data []*pb.DmVertical, meta lys.GetMetadata) *pb.ListDmVerticalsResp {
			return &pb.ListDmVerticalsResp{Data: data, Metadata: toPbMetadata(meta, len(data))}
		})
}

func (s *dmVertServer) Update(ctx context.Context, req *pb.UpdateDmVerticalReq) (*pb.DmVertical, error) {
	return lysp.RpcUpdate(ctx, req.Input, req.FieldMask, req.Id, s.jsonKeyDbNameMap, s.store, dmvertical.ModelToPbFunc)
}
