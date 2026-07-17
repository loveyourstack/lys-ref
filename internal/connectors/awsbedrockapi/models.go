package awsbedrockapi

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
)

type Model struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) ListImageModels(ctx context.Context) (models []Model, err error) {

	// make bedrock client if needed
	err = c.makeBedrockClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("c.makeBedrockClient failed: %w", err)
	}

	modelResp, err := c.bedrockClient.ListFoundationModels(ctx, &bedrock.ListFoundationModelsInput{
		ByOutputModality: "IMAGE",
	})
	if err != nil {
		return nil, fmt.Errorf("c.bedrockClient.ListFoundationModels failed: %w", err)
	}
	if len(modelResp.ModelSummaries) == 0 {
		return nil, fmt.Errorf("no foundation models found")
	}

	for _, model := range modelResp.ModelSummaries {
		models = append(models, Model{
			Id:   *model.ModelId,
			Name: *model.ModelName,
		})
	}

	return models, nil
}
