package lysp

import (
	"context"
	"fmt"
)

type iRpcCreateable[inputT any, modelT any] interface {
	InsertSelect(ctx context.Context, input inputT) (modelT, error)
}

func RpcCreate[inputT any, modelT any, pbInputT any, pbModelT any](
	ctx context.Context,
	pbInput pbInputT,
	store iRpcCreateable[inputT, modelT],
	pbToInput func(pbInputT) inputT,
	toPb func(modelT) *pbModelT,
) (*pbModelT, error) {

	input := pbToInput(pbInput)

	item, err := store.InsertSelect(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("store.InsertSelect failed: %w", err)
	}

	return toPb(item), nil
}
