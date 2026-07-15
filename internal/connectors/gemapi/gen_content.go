package gemapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/loveyourstack/lys/lysstring"
	"github.com/loveyourstack/lys/lystype"
	"google.golang.org/genai"
)

// GenerateImage generates an image from the prompt using the Gemini API and saves it to the configured path.
// Recommended model: gemini-3.1-flash-lite-image
func (c Client) GenerateImage(ctx context.Context, model, prompt string) (fName string, err error) {

	if model == "" {
		return "", fmt.Errorf("model is empty")
	}
	if prompt == "" {
		return "", fmt.Errorf("prompt is empty")
	}

	resp, err := c.genAiClient.Models.GenerateImages(ctx, model, prompt, &genai.GenerateImagesConfig{
		NumberOfImages: 1,
	})
	if err != nil {
		return "", fmt.Errorf("c.genAiClient.Models.GenerateImages failed: %w", err)
	}

	if len(resp.GeneratedImages) != 1 {
		return "", fmt.Errorf("expected 1 generated image, got %d", len(resp.GeneratedImages))
	}
	img := resp.GeneratedImages[0]

	ext := ""
	switch img.Image.MIMEType {
	case "image/gif":
		ext = "gif"
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/webp":
		ext = "webp"
	default:
		return "", fmt.Errorf("unsupported image MIME type: %s", img.Image.MIMEType)
	}

	fName = fmt.Sprintf("%s-%s.%s", time.Now().Format(lystype.DateFormat), lysstring.Rand(8), ext)
	err = os.WriteFile(fmt.Sprintf("%s/%s", c.generatedPath, fName), img.Image.ImageBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("os.WriteFile failed: %w", err)
	}

	return fName, nil
}

type MarketingCampaign struct {
	CallToAction string `json:"callToAction"`
	Body         string `json:"body"`
	Headline     string `json:"headline"`
}

// GenerateMarketingCampaign generates a marketing campaign for the given product using the Gemini API and returns it as a MarketingCampaign struct.
// Recommended model: gemini-3.1-flash-lite
func (c Client) GenerateMarketingCampaign(ctx context.Context, model, product string) (camp MarketingCampaign, err error) {

	if model == "" {
		return MarketingCampaign{}, fmt.Errorf("model is empty")
	}
	if product == "" {
		return MarketingCampaign{}, fmt.Errorf("product is empty")
	}

	prompt := fmt.Sprintf(
		"Generate a marketing campaign for the product %q, consisting of a JSON object with exactly these keys: headline, body, callToAction. "+
			"headline is a catchy phrase of up to 10 words. "+
			"body is a paragraph of up to 50 words selling the product to the reader. "+
			"callToAction will appear on the button, and is a short phrase of up to 3 words that encourages the user to take action.",
		product,
	)

	cfg := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"headline":     map[string]any{"type": "string"},
				"body":         map[string]any{"type": "string"},
				"callToAction": map[string]any{"type": "string"},
			},
			"required":             []string{"headline", "body", "callToAction"},
			"additionalProperties": false,
		},
	}

	resp, err := c.genAiClient.Models.GenerateContent(ctx, model, genai.Text(prompt), cfg)
	if err != nil {
		return camp, fmt.Errorf("c.genAiClient.Models.GenerateContent failed: %w", err)
	}

	raw := strings.TrimSpace(resp.Text())
	if raw == "" {
		return camp, fmt.Errorf("empty model response")
	}

	err = json.Unmarshal([]byte(raw), &camp)
	if err != nil {
		return camp, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	if camp.Headline == "" || camp.Body == "" || camp.CallToAction == "" {
		return MarketingCampaign{}, fmt.Errorf("campaign contains one or more empty fields")
	}

	return camp, nil
}

// GenerateText generates text from the prompt using the Gemini API and returns it as a string.
// Recommended model: gemini-3.1-flash-lite
func (c Client) GenerateText(ctx context.Context, model string, prompt string) (res string, err error) {

	if model == "" {
		return "", fmt.Errorf("model is empty")
	}
	if prompt == "" {
		return "", fmt.Errorf("prompt is empty")
	}

	resp, err := c.genAiClient.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("c.genAiClient.Models.GenerateContent failed: %w", err)
	}

	return resp.Text(), nil
}
