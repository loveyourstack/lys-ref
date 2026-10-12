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
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/pkg/lysp"
	"github.com/loveyourstack/lys/lysset"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type dmCampSrv struct {
	pb.UnimplementedDmCampaignSvcServer
	*gRpcServerApplication

	store            dmcampaign.Store
	dbNames          lysset.Set[string]
	jsonKeyDbNameMap map[string]string
}

func (s *dmCampSrv) Create(ctx context.Context, req *pb.DmCampaignInput) (*pb.DmCampaign, error) {

	input := dmcampaign.Input{
		CountryFk:      req.CountryFk,
		DailyBudgetEur: req.DailyBudgetEur,
		IsActive:       req.IsActive,
		Manager:        req.Manager,
		Name:           req.Name,
		VerticalFk:     req.VerticalFk,
	}

	item, err := s.store.InsertSelect(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("s.store.InsertSelect failed: %w", err)
	}

	return dmCampToPb(item), nil
}

func (s *dmCampSrv) Delete(ctx context.Context, req *pb.DeleteByIdReq) (*emptypb.Empty, error) {

	err := s.store.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "campaign not found")
		}
		return nil, fmt.Errorf("s.store.Delete failed: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *dmCampSrv) Get(ctx context.Context, req *pb.GetByIdReq) (*pb.DmCampaign, error) {

	item, err := s.store.SelectById(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "campaign not found")
		}
		return nil, fmt.Errorf("s.store.SelectById failed: %w", err)
	}

	return dmCampToPb(item), nil
}

func (s *dmCampSrv) List(ctx context.Context, req *pb.ListReq) (*pb.ListDmCampaignsResp, error) {
	return lysp.RpcList(ctx, s.store, req, s.GetOptions, dmCampToPb,
		func(data []*pb.DmCampaign, meta lys.GetMetadata) *pb.ListDmCampaignsResp {
			return &pb.ListDmCampaignsResp{Data: data, Metadata: toPbMetadata(meta, len(data))}
		})
}

func (s *dmCampSrv) Update(ctx context.Context, req *pb.UpdateDmCampaignReq) (*pb.DmCampaign, error) {

	if req.Input == nil {
		return nil, status.Error(codes.InvalidArgument, "input is required")
	}
	input := req.Input

	if req.FieldMask == nil {
		return nil, status.Error(codes.InvalidArgument, "field mask is required")
	}
	paths := req.FieldMask.Paths

	assMap := map[string]any{}
	if slices.Contains(paths, "country_fk") {
		assMap["country_fk"] = input.CountryFk
	}
	if slices.Contains(paths, "daily_budget_eur") {
		assMap["daily_budget_eur"] = input.DailyBudgetEur
	}
	if slices.Contains(paths, "is_active") {
		assMap["is_active"] = input.IsActive
	}
	if slices.Contains(paths, "manager") {
		assMap["manager"] = input.Manager
	}
	if slices.Contains(paths, "name") {
		assMap["name"] = input.Name
	}
	if slices.Contains(paths, "vertical_fk") {
		assMap["vertical_fk"] = input.VerticalFk
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

	return dmCampToPb(item), nil
}

func dmCampToPb(m dmcampaign.Model) *pb.DmCampaign {
	return &pb.DmCampaign{
		Id:               m.Id,
		Country:          m.Country,
		CountryIso2:      m.CountryIso2,
		CreatedAt:        tsToPb(time.Time(m.CreatedAt)),
		PerformanceRange: m.PerformanceRange,
		UpdatedAt:        tsToPb(time.Time(m.UpdatedAt)),
		Vertical:         m.Vertical,

		CountryFk:      m.CountryFk,
		DailyBudgetEur: m.DailyBudgetEur,
		IsActive:       m.IsActive,
		Manager:        m.Manager,
		Name:           m.Name,
		VerticalFk:     m.VerticalFk,
	}
}
