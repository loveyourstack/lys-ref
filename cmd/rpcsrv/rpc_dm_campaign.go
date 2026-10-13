package main

import (
	"context"
	"time"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd/rpcsrv/pb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/pkg/lysp"
	"google.golang.org/protobuf/types/known/emptypb"
)

type dmCampSrv struct {
	pb.UnimplementedDmCampaignSvcServer
	*gRpcServerApplication

	store            dmcampaign.Store
	jsonKeyDbNameMap map[string]string
}

func (s *dmCampSrv) Create(ctx context.Context, req *pb.DmCampaignInput) (*pb.DmCampaign, error) {
	return lysp.RpcCreate(ctx, req, s.store, dmCampPbToInput, dmCampToPb)
}

func (s *dmCampSrv) Delete(ctx context.Context, req *pb.DeleteByIdReq) (*emptypb.Empty, error) {
	return lysp.RpcDelete(ctx, req.Id, s.store)
}

func (s *dmCampSrv) Get(ctx context.Context, req *pb.GetByIdReq) (*pb.DmCampaign, error) {
	return lysp.RpcGetById(ctx, req.Id, s.store, dmCampToPb)
}

func (s *dmCampSrv) List(ctx context.Context, req *pb.ListReq) (*pb.ListDmCampaignsResp, error) {
	return lysp.RpcList(ctx, s.store, req, s.GetOptions, dmCampToPb,
		func(data []*pb.DmCampaign, meta lys.GetMetadata) *pb.ListDmCampaignsResp {
			return &pb.ListDmCampaignsResp{Data: data, Metadata: toPbMetadata(meta, len(data))}
		})
}

func (s *dmCampSrv) Update(ctx context.Context, req *pb.UpdateDmCampaignReq) (*pb.DmCampaign, error) {
	return lysp.RpcUpdate(ctx, req.Input, req.FieldMask, req.Id, s.jsonKeyDbNameMap, s.store, dmCampToPb)
}

func dmCampPbToInput(pb *pb.DmCampaignInput) dmcampaign.Input {
	return dmcampaign.Input{
		CountryFk:      pb.CountryFk,
		DailyBudgetEur: pb.DailyBudgetEur,
		IsActive:       pb.IsActive,
		Manager:        pb.Manager,
		Name:           pb.Name,
		VerticalFk:     pb.VerticalFk,
	}
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
