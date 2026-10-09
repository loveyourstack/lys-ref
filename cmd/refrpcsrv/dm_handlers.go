package main

import (
	"context"
	"time"

	"github.com/loveyourstack/lys-ref/cmd/refrpcsrv/pb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys/lyspg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type campaignServer struct {
	pb.UnimplementedCampaignServiceServer
	*gRpcServerApplication
}

func (s *campaignServer) List(ctx context.Context, req *pb.ListCampaignsRequest) (*pb.ListCampaignsResponse, error) {

	campStore := dmcampaign.Store{Db: s.Db}
	items, _, err := campStore.Select(ctx, lyspg.SelectParams{})
	if err != nil {
		s.Logger.Error("dmcampaign.Select failed", "error", err)
		return nil, status.Error(codes.Internal, "select failed")
	}

	resp := &pb.ListCampaignsResponse{Items: make([]*pb.DmCampaign, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, dmCampaignToPb(item))
	}

	return resp, nil
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
}

func (s *verticalServer) List(ctx context.Context, req *pb.ListVerticalsRequest) (*pb.ListVerticalsResponse, error) {

	// select verticals from db
	vertStore := dmvertical.Store{Db: s.Db}
	items, _, err := vertStore.Select(ctx, lyspg.SelectParams{})
	if err != nil {
		s.Logger.Error("dmvertical.Select failed", "error", err)
		return nil, status.Error(codes.Internal, "select failed")
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
