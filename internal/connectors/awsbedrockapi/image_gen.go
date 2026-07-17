package awsbedrockapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/loveyourstack/lys/lysstring"
	"github.com/loveyourstack/lys/lystype"
)

func (c *Client) GenerateMarketingImage(ctx context.Context, product string) (fName string, err error) {

	if product == "" {
		return "", fmt.Errorf("product is required")
	}

	// make bedrock runtime client if needed
	err = c.makeBedrockRtClient(ctx)
	if err != nil {
		return "", fmt.Errorf("c.makeBedrockRtClient failed: %w", err)
	}

	extension := "png"
	reqBody := map[string]any{
		"prompt":        fmt.Sprintf("Generate a marketing image for the product: %q", product),
		"aspect_ratio":  "21:9",
		"output_format": extension,
		"style_preset":  "photographic",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("json.Marshal failed: %w", err)
	}

	resp, err := c.bedrockRtClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("stability.stable-image-core-v1:1"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        bodyBytes,
	})
	if err != nil {
		return "", fmt.Errorf("invoke model: %w", err)
	}

	type stableImageCoreResponse struct {
		Images []string `json:"images"`
	}

	var out stableImageCoreResponse
	if err := json.Unmarshal(resp.Body, &out); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if len(out.Images) != 1 {
		return "", fmt.Errorf("expected 1 image, got %d", len(out.Images))
	}

	imgBytes, err := base64.StdEncoding.DecodeString(out.Images[0])
	if err != nil {
		return "", fmt.Errorf("base64.StdEncoding.DecodeString failed: %w", err)
	}

	fName = fmt.Sprintf("%s-%s.%s", time.Now().Format(lystype.DateFormat), lysstring.Rand(8), extension)
	if err := os.WriteFile(fmt.Sprintf("%s/%s", c.generatedPath, fName), imgBytes, 0644); err != nil {
		return "", fmt.Errorf("os.WriteFile failed: %w", err)
	}

	return fName, nil
}
