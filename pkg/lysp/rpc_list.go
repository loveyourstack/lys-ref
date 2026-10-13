package lysp

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysset"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// in lys: use iGetable
type iRpcListable[T any] interface {
	GetPlan() lysmeta.Plan
	Select(ctx context.Context, params lyspg.SelectParams) (items []T, unpagedCount lyspg.TotalCount, err error)
}

// iRpcListRequest is a request that can be used by RpcList.
// The type will typically be a protobuf message generated from a .proto file.
type iRpcListRequest interface {
	GetFields() []string
	GetSorts() []string
	GetPage() int32
	GetPerPage() int32
	GetFilters() map[string]string
}

// RpcList parses the request, selects items from the store, converts them using toPb and assembles the response using build.
func RpcList[modelT, pbModelT, pbRespT any](
	ctx context.Context,
	store iRpcListable[modelT],
	req iRpcListRequest,
	getOptions lys.GetOptions,
	toPb func(modelT) *pbModelT,
	build func(data []*pbModelT, meta lys.GetMetadata) *pbRespT,
) (*pbRespT, error) {

	// parse request and select items from db
	items, unpagedCount, err := rpcSelect(ctx, store, req, getOptions)
	if err != nil {
		return nil, err // already a gRPC status error where appropriate
	}

	// convert db items to protobuf models
	data := make([]*pbModelT, 0, len(items))
	for _, item := range items {
		data = append(data, toPb(item))
	}

	// when using fields param, Postman still shows fields with zero values, but using protojson.Marshal
	// shows the actual response. Use this to check that fields are omitted correctly.
	//b, _ := protojson.Marshal(data)
	//fmt.Println(string(b))

	// build the response using the provided build function and metadata
	return build(data, lys.GetMetadata{
		Count:                 len(items),
		TotalCount:            unpagedCount.Value,
		TotalCountIsEstimated: unpagedCount.IsEstimated,
	}), nil
}

func rpcSelect[T any](ctx context.Context, store iRpcListable[T], req iRpcListRequest, getOptions lys.GetOptions) (items []T, unpagedCount lyspg.TotalCount, err error) {

	dbNames := lysset.FromSlice(store.GetPlan().DbNames())
	jsonKeyDbNameMap := store.GetPlan().JsonKeyDbNameMap()

	// parse fields, if set
	fields, err := lys.ExtractFields(getOptions.FieldsParamName, strings.Join(req.GetFields(), ","), dbNames, jsonKeyDbNameMap)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, lyspg.TotalCount{}, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, lyspg.TotalCount{}, fmt.Errorf("lys.ExtractFields failed: %w", err)
	}

	// parse sorts, if set
	sorts, err := lys.ExtractSorts(getOptions.SortParamName, strings.Join(req.GetSorts(), ","), jsonKeyDbNameMap)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, lyspg.TotalCount{}, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, lyspg.TotalCount{}, fmt.Errorf("lys.ExtractSorts failed: %w", err)
	}

	// apply default paging if not set
	pageVal := req.GetPage()
	if pageVal == 0 {
		pageVal = 1
	}
	perPageVal := req.GetPerPage()
	if perPageVal == 0 {
		perPageVal = int32(getOptions.DefaultPerPage)
	}

	// parse paging
	page, perPage, err := lys.ExtractPaging(getOptions.PageParamName, fmt.Sprintf("%d", pageVal), getOptions.PerPageParamName,
		fmt.Sprintf("%d", perPageVal), getOptions.DefaultPerPage, getOptions.MaxPerPage)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, lyspg.TotalCount{}, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, lyspg.TotalCount{}, fmt.Errorf("lys.ExtractPaging failed: %w", err)
	}

	// get offset from paging params (starts at 0, not 1)
	offset := perPage * (page - 1)

	// parse filters, if set

	// convert req map to url.Values for so lys library func can be used
	values := url.Values{}
	for k, v := range req.GetFilters() {
		values.Set(k, v)
	}

	conds, err := lys.ExtractFilters(values, jsonKeyDbNameMap, nil, nil, getOptions)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, lyspg.TotalCount{}, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, lyspg.TotalCount{}, fmt.Errorf("lys.ExtractFilters failed: %w", err)
	}

	// select items from db
	items, unpagedCount, err = store.Select(ctx, lyspg.SelectParams{
		Fields:          fields,
		Conditions:      conds,
		Sorts:           sorts,
		Limit:           perPage,
		Offset:          offset,
		GetUnpagedCount: true,
	})
	if err != nil {
		return nil, lyspg.TotalCount{}, fmt.Errorf("store.Select failed: %w", err)
	}

	return items, unpagedCount, nil
}
