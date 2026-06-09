package awsapi

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	awsCreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type Conf struct {
	AccessKeyId     string // IAM user access key ID
	Region          string
	SecretAccessKey string
}

type Client struct {
	// callStore awsapicall.Store
	conf     Conf
	errorLog *slog.Logger
	infoLog  *slog.Logger
}

func NewClient(conf Conf, infoLog, errorLog *slog.Logger) (client Client) {

	if conf.AccessKeyId == "" {
		log.Fatal("conf.AccessKeyId is required")
	}
	if conf.Region == "" {
		log.Fatal("conf.Region is required")
	}
	if conf.SecretAccessKey == "" {
		log.Fatal("conf.SecretAccessKey is required")
	}

	apiShortname := "aws"

	return Client{
		conf:     conf,
		infoLog:  infoLog.With("api", apiShortname),
		errorLog: errorLog.With("api", apiShortname),
	}
}

func (c Client) Connect(ctx context.Context) (cfg aws.Config, err error) {

	staticProvider := awsCreds.NewStaticCredentialsProvider(c.conf.AccessKeyId, c.conf.SecretAccessKey, "")
	cfg, err = awsCfg.LoadDefaultConfig(ctx, awsCfg.WithRegion(c.conf.Region), awsCfg.WithCredentialsProvider(staticProvider))
	if err != nil {
		return aws.Config{}, fmt.Errorf("awsCfg.LoadDefaultConfig failed: %w", err)
	}

	return cfg, nil
}

func (c Client) MakeEc2Client(ctx context.Context) (*ec2.Client, error) {
	cfg, err := c.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("c.Connect failed: %w", err)
	}

	return ec2.NewFromConfig(cfg), nil
}
