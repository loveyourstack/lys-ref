package awsbedrockapi

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	awsCreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/connectors/aws/awsapi"
	"github.com/loveyourstack/connectors/aws/stores/awsapicall"
)

type Client struct {
	callStore awsapicall.Store
	conf      awsapi.Conf

	bedrockClient   *bedrock.Client
	bedrockRtClient *bedrockruntime.Client
	generatedPath   string

	logger *slog.Logger
}

// NewClient creates a new AWS API client.
func NewClient(conf awsapi.Conf, generatedPath string, db *pgxpool.Pool, logger *slog.Logger) *Client {

	if conf.AccessKeyId == "" {
		log.Fatal("awsapi client: conf.AccessKeyId is required")
	}
	if conf.Region == "" {
		log.Fatal("awsapi client: conf.Region is required")
	}
	if conf.SecretAccessKey == "" {
		log.Fatal("awsapi client: conf.SecretAccessKey is required")
	}

	apiShortname := "awsbedrock"

	// return pointer: client is modified by lazy initialization
	return &Client{
		conf:      conf,
		callStore: awsapicall.Store{Db: db},

		bedrockClient:   nil, // lazily initialized in makeBedrockClient
		bedrockRtClient: nil, // lazily initialized in makeBedrockRtClient

		generatedPath: generatedPath,
		logger:        logger.With("api", apiShortname),
	}
}

// connect return AWS config using the provided credentials and region.
func (c *Client) connect(ctx context.Context) (cfg aws.Config, err error) {

	staticProvider := awsCreds.NewStaticCredentialsProvider(c.conf.AccessKeyId, c.conf.SecretAccessKey, "")
	cfg, err = awsCfg.LoadDefaultConfig(ctx, awsCfg.WithRegion(c.conf.Region), awsCfg.WithCredentialsProvider(staticProvider))
	if err != nil {
		return aws.Config{}, fmt.Errorf("awsCfg.LoadDefaultConfig failed: %w", err)
	}

	return cfg, nil
}

// makeBedrockClient lazily initializes the Bedrock client and reuses it for subsequent calls
func (c *Client) makeBedrockClient(ctx context.Context) (err error) {

	if c.bedrockClient != nil {
		return nil
	}

	cfg, err := c.connect(ctx)
	if err != nil {
		return fmt.Errorf("c.connect failed: %w", err)
	}

	c.bedrockClient = bedrock.NewFromConfig(cfg)
	return nil
}

// makeBedrockRtClient lazily initializes the Bedrock Runtime client and reuses it for subsequent calls
func (c *Client) makeBedrockRtClient(ctx context.Context) (err error) {

	if c.bedrockRtClient != nil {
		return nil
	}

	cfg, err := c.connect(ctx)
	if err != nil {
		return fmt.Errorf("c.connect failed: %w", err)
	}

	c.bedrockRtClient = bedrockruntime.NewFromConfig(cfg)
	return nil
}
