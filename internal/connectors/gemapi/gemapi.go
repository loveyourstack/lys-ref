package gemapi

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/genai"
)

type Client struct {
	//callStore gemapicall.Store

	genAiClient   *genai.Client
	generatedPath string
	logger        *slog.Logger
}

// NewClient creates a new Gemini API client.
func NewClient(ctx context.Context, db *pgxpool.Pool, generatedPath string, logger *slog.Logger) Client {

	if generatedPath == "" {
		log.Fatal("gemapi: generatedPath is empty")
	}

	genAiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		// APIKey: use env var GEMINI_API_KEY
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(fmt.Errorf("genai.NewClient failed: %w", err))
	}

	return Client{
		//callStore: gemapicall.Store{Db: db},

		genAiClient:   genAiClient,
		generatedPath: generatedPath,
		logger:        logger.With("api", "gem"),
	}
}
