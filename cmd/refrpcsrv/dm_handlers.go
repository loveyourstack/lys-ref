package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys-ref/cmd/refrpcsrv/pb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys/lyspg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type campaignServer struct {
	pb.UnimplementedCampaignServiceServer
	*gRpcServerApplication

	store dmcampaign.Store
}

func (s *campaignServer) Create(ctx context.Context, req *pb.DmCampaignInput) (*pb.DmCampaign, error) {

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

	return dmCampaignToPb(item), nil
}

func (s *campaignServer) Delete(ctx context.Context, req *pb.DeleteCampaignRequest) (*emptypb.Empty, error) {

	err := s.store.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "campaign not found")
		}
		return nil, fmt.Errorf("s.store.Delete failed: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *campaignServer) Get(ctx context.Context, req *pb.GetCampaignRequest) (*pb.DmCampaign, error) {

	item, err := s.store.SelectById(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "campaign not found")
		}
		return nil, fmt.Errorf("s.store.SelectById failed: %w", err)
	}

	return dmCampaignToPb(item), nil
}

func (s *campaignServer) List(ctx context.Context, req *pb.ListCampaignsRequest) (*pb.ListCampaignsResponse, error) {

	items, _, err := s.store.Select(ctx, lyspg.SelectParams{})
	if err != nil {
		return nil, fmt.Errorf("s.store.Select failed: %w", err)
	}

	resp := &pb.ListCampaignsResponse{Items: make([]*pb.DmCampaign, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, dmCampaignToPb(item))
	}

	return resp, nil
}

func (s *campaignServer) Update(ctx context.Context, req *pb.UpdateCampaignRequest) (*pb.DmCampaign, error) {

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

	return dmCampaignToPb(item), nil
}

func dmCampaignToPb(m dmcampaign.Model) *pb.DmCampaign {
	return &pb.DmCampaign{
		Id:               m.Id,
		Country:          m.Country,
		CountryIso2:      m.CountryIso2,
		CreatedAt:        timestamppb.New(time.Time(m.CreatedAt)),
		PerformanceRange: m.PerformanceRange,
		UpdatedAt:        timestamppb.New(time.Time(m.UpdatedAt)),
		Vertical:         m.Vertical,

		CountryFk:      m.CountryFk,
		DailyBudgetEur: m.DailyBudgetEur,
		IsActive:       m.IsActive,
		Manager:        m.Manager,
		Name:           m.Name,
		VerticalFk:     m.VerticalFk,
	}
}

type verticalServer struct {
	pb.UnimplementedVerticalServiceServer
	*gRpcServerApplication

	store dmvertical.Store
}

func (s *verticalServer) List(ctx context.Context, req *pb.ListVerticalsRequest) (*pb.ListVerticalsResponse, error) {

	// select verticals from db
	items, _, err := s.store.Select(ctx, lyspg.SelectParams{})
	if err != nil {
		return nil, fmt.Errorf("s.store.Select failed: %w", err)
	}

	// convert db models to protobuf response
	resp := &pb.ListVerticalsResponse{Items: make([]*pb.DmVertical, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, dmVerticalToPb(item))
	}

	return resp, nil
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
