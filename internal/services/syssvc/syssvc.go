package syssvc

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/connectors/aws/awsapi"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
)

type Service struct {
	AwsApiClient *awsapi.Client
	S3Bucket     string

	SysUserStore sysuser.Store

	Logger *slog.Logger
}

func NewService(awsApiClient *awsapi.Client, s3Bucket string, db *pgxpool.Pool, logger *slog.Logger) (svc Service) {

	svcShortname := "sys"

	return Service{
		AwsApiClient: awsApiClient,
		S3Bucket:     s3Bucket,

		SysUserStore: sysuser.Store{Db: db},

		Logger: logger.With("svc", svcShortname),
	}
}
