package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd/rpcsrv/pb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys-ref/pkg/lysp"
	"github.com/loveyourstack/lys/lysset"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type dmVertServer struct {
	pb.UnimplementedDmVerticalSvcServer
	*gRpcServerApplication

	store            dmvertical.Store
	dbNames          lysset.Set[string]
	jsonKeyDbNameMap map[string]string
}

func (s *dmVertServer) Create(ctx context.Context, req *pb.DmVerticalInput) (*pb.DmVertical, error) {

	input := dmvertical.Input{
		Name: req.Name,
	}

	item, err := s.store.InsertSelect(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("s.store.InsertSelect failed: %w", err)
	}

	return dmVerticalToPb(item), nil
}

func (s *dmVertServer) Delete(ctx context.Context, req *pb.DeleteByIdReq) (*emptypb.Empty, error) {

	err := s.store.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "vertical not found")
		}
		return nil, fmt.Errorf("s.store.Delete failed: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *dmVertServer) Get(ctx context.Context, req *pb.GetByIdReq) (*pb.DmVertical, error) {

	item, err := s.store.SelectById(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "vertical not found")
		}
		return nil, fmt.Errorf("s.store.SelectById failed: %w", err)
	}

	return dmVerticalToPb(item), nil
}

func (s *dmVertServer) List(ctx context.Context, req *pb.ListReq) (*pb.ListDmVerticalsResp, error) {
	return lysp.RpcList(ctx, s.store, req, s.GetOptions, dmVerticalToPb,
		func(data []*pb.DmVertical, meta lys.GetMetadata) *pb.ListDmVerticalsResp {
			return &pb.ListDmVerticalsResp{Data: data, Metadata: toPbMetadata(meta, len(data))}
		})
}

func (s *dmVertServer) Update(ctx context.Context, req *pb.UpdateDmVerticalReq) (*pb.DmVertical, error) {

	if req.Input == nil {
		return nil, status.Error(codes.InvalidArgument, "input is required")
	}
	input := req.Input

	if req.FieldMask == nil {
		return nil, status.Error(codes.InvalidArgument, "field mask is required")
	}
	paths := req.FieldMask.Paths

	assMap := map[string]any{}

	if slices.Contains(paths, "name") {
		assMap["name"] = input.Name
	}

	if len(assMap) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	err := s.store.UpdatePartial(ctx, assMap, req.Id)
	if err != nil {
		return nil, fmt.Errorf("s.store.UpdatePartial failed: %w", err)
	}

	item, err := s.store.SelectById(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("s.store.SelectById failed: %w", err)
	}

	return dmVerticalToPb(item), nil
}

func dmVerticalToPb(m dmvertical.Model) *pb.DmVertical {
	return &pb.DmVertical{
		Id:            m.Id,
		CampaignCount: int32(m.CampaignCount),
		CreatedAt:     timestamppb.New(time.Time(m.CreatedAt)),
		UpdatedAt:     timestamppb.New(time.Time(m.UpdatedAt)),

		Name: m.Name,
	}
}
