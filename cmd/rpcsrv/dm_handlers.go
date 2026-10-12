package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd/rpcsrv/pb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysset"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type campaignServer struct {
	pb.UnimplementedCampaignServiceServer
	*gRpcServerApplication

	store            dmcampaign.Store
	dbNames          lysset.Set[string]
	jsonKeyDbNameMap map[string]string
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

func (s *campaignServer) Delete(ctx context.Context, req *pb.DeleteByIdRequest) (*emptypb.Empty, error) {

	err := s.store.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "campaign not found")
		}
		return nil, fmt.Errorf("s.store.Delete failed: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *campaignServer) Get(ctx context.Context, req *pb.GetByIdRequest) (*pb.DmCampaign, error) {

	item, err := s.store.SelectById(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "campaign not found")
		}
		return nil, fmt.Errorf("s.store.SelectById failed: %w", err)
	}

	return dmCampaignToPb(item), nil
}

func (s *campaignServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListCampaignsResponse, error) {

	// parse fields, if set
	fields, err := lys.ExtractFields(s.GetOptions.FieldsParamName, strings.Join(req.Fields, ","), s.dbNames, s.jsonKeyDbNameMap)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, fmt.Errorf("lys.ExtractFields failed: %w", err)
	}

	// parse sorts, if set
	sorts, err := lys.ExtractSorts(s.GetOptions.SortParamName, strings.Join(req.Sorts, ","), s.jsonKeyDbNameMap)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, fmt.Errorf("lys.ExtractSorts failed: %w", err)
	}

	// apply default paging if not set
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PerPage == 0 {
		req.PerPage = int32(s.GetOptions.DefaultPerPage)
	}

	// parse paging
	page, perPage, err := lys.ExtractPaging(s.GetOptions.PageParamName, fmt.Sprintf("%d", req.Page), s.GetOptions.PerPageParamName,
		fmt.Sprintf("%d", req.PerPage), s.GetOptions.DefaultPerPage, s.GetOptions.MaxPerPage)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, fmt.Errorf("lys.ExtractPaging failed: %w", err)
	}

	// get offset from paging params (starts at 0, not 1)
	offset := perPage * (page - 1)

	// parse filters, if set

	// convert req map to url.Values for so library func can be used
	values := url.Values{}
	for k, v := range req.Filters {
		values.Set(k, v)
	}

	conds, err := lys.ExtractFilters(values, s.jsonKeyDbNameMap, nil, nil, s.GetOptions)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, fmt.Errorf("lys.ExtractFilters failed: %w", err)
	}

	// select items from db
	items, _, err := s.store.Select(ctx, lyspg.SelectParams{
		Fields:     fields,
		Conditions: conds,
		Sorts:      sorts,
		Limit:      perPage,
		Offset:     offset,
	})
	if err != nil {
		return nil, fmt.Errorf("s.store.Select failed: %w", err)
	}

	// convert db models to protobuf response
	resp := &pb.ListCampaignsResponse{Items: make([]*pb.DmCampaign, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, dmCampaignToPb(item))
	}

	// when using fields param, Postman still shows fields with zero values, but using protojson.Marshal
	// shows the actual response. Use this to check that fields are omitted correctly.
	//b, _ := protojson.Marshal(resp)
	//fmt.Println(string(b))

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

// tsToPb converts a time.Time to a protobuf Timestamp, returning nil if the time is zero.
// Use this to ensure that zero-value timestamps are omitted when using the fields param.
func tsToPb(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

type verticalServer struct {
	pb.UnimplementedVerticalServiceServer
	*gRpcServerApplication

	store            dmvertical.Store
	dbNames          lysset.Set[string]
	jsonKeyDbNameMap map[string]string
}

func (s *verticalServer) Create(ctx context.Context, req *pb.DmVerticalInput) (*pb.DmVertical, error) {

	input := dmvertical.Input{
		Name: req.Name,
	}

	item, err := s.store.InsertSelect(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("s.store.InsertSelect failed: %w", err)
	}

	return dmVerticalToPb(item), nil
}

func (s *verticalServer) Delete(ctx context.Context, req *pb.DeleteByIdRequest) (*emptypb.Empty, error) {

	err := s.store.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "vertical not found")
		}
		return nil, fmt.Errorf("s.store.Delete failed: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *verticalServer) Get(ctx context.Context, req *pb.GetByIdRequest) (*pb.DmVertical, error) {

	item, err := s.store.SelectById(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "vertical not found")
		}
		return nil, fmt.Errorf("s.store.SelectById failed: %w", err)
	}

	return dmVerticalToPb(item), nil
}

func (s *verticalServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListVerticalsResponse, error) {

	// select items from db
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

func (s *verticalServer) Update(ctx context.Context, req *pb.UpdateVerticalRequest) (*pb.DmVertical, error) {

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
