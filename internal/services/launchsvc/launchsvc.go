package launchsvc

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchfb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchgads"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
)

type Service struct {
	FbLaunchStore   dmlaunchfb.Store
	GAdsLaunchStore dmlaunchgads.Store

	CoStore   geocountry.Store
	VertStore dmvertical.Store

	Db     *pgxpool.Pool // for tx
	Logger *slog.Logger
}

func NewService(db *pgxpool.Pool, logger *slog.Logger) (svc Service) {

	svcShortname := "launch"

	return Service{
		FbLaunchStore:   dmlaunchfb.Store{Db: db},
		GAdsLaunchStore: dmlaunchgads.Store{Db: db},

		CoStore:   geocountry.Store{Db: db},
		VertStore: dmvertical.Store{Db: db},

		Db:     db,
		Logger: logger.With("svc", svcShortname),
	}
}
