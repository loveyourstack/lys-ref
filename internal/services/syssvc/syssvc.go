package syssvc

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
)

type Service struct {
	SysUserStore sysuser.Store

	Logger *slog.Logger
}

func NewService(db *pgxpool.Pool, logger *slog.Logger) (svc Service) {

	svcShortname := "sys"

	return Service{
		SysUserStore: sysuser.Store{Db: db},

		Logger: logger.With("svc", svcShortname),
	}
}
