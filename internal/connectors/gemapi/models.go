package gemapi

import "context"

func (c Client) ListModels(ctx context.Context) (models []string, err error) {
	resp, err := c.genAiClient.Models.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, m := range resp.Items {
		models = append(models, m.Name)
	}

	return models, nil
}
