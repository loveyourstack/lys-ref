package lysp

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyspg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type iDeletable[idT lyspg.PrimaryKeyType] interface {
	Delete(ctx context.Context, id idT) error
}

func RpcDelete[idT lyspg.PrimaryKeyType](ctx context.Context, id idT, store iDeletable[idT]) (*emptypb.Empty, error) {

	err := store.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, fmt.Errorf("store.Delete failed: %w", err)
	}

	return &emptypb.Empty{}, nil
}
