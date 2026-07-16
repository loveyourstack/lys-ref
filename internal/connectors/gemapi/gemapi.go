package gemapi

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/stores/gemini/gemapicall"
	"google.golang.org/genai"
)

type Conf struct {
	ApiKey string
}

type Client struct {
	callStore gemapicall.Store

	genAiClient   *genai.Client
	generatedPath string
	logger        *slog.Logger
}

// NewClient creates a new Gemini API client.
func NewClient(ctx context.Context, conf Conf, generatedPath string, db *pgxpool.Pool, logger *slog.Logger) Client {

	if conf.ApiKey == "" {
		log.Fatal("gemapi: conf.ApiKey is empty")
	}
	if generatedPath == "" {
		log.Fatal("gemapi: generatedPath is empty")
	}

	genAiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  conf.ApiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(fmt.Errorf("genai.NewClient failed: %w", err))
	}

	return Client{
		callStore: gemapicall.Store{Db: db},

		genAiClient:   genAiClient,
		generatedPath: generatedPath,
		logger:        logger.With("api", "gem"),
	}
}
