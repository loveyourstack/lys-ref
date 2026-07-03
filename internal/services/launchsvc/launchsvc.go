package launchsvc

import (
	"log"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchfb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchgads"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
)

type Service struct {
	PrepBatchSize int

	FbLaunchStore   dmlaunchfb.Store
	GAdsLaunchStore dmlaunchgads.Store

	CoStore   geocountry.Store
	VertStore dmvertical.Store

	Db     *pgxpool.Pool // for tx
	Logger *slog.Logger
}

func NewService(db *pgxpool.Pool, prepBatchSize int, logger *slog.Logger) (svc Service) {

	if prepBatchSize <= 0 {
		log.Fatal("prepBatchSize must be > 0")
	}

	svcShortname := "launch"

	return Service{
		PrepBatchSize: prepBatchSize,

		FbLaunchStore:   dmlaunchfb.Store{Db: db},
		GAdsLaunchStore: dmlaunchgads.Store{Db: db},

		CoStore:   geocountry.Store{Db: db},
		VertStore: dmvertical.Store{Db: db},

		Db:     db,
		Logger: logger.With("svc", svcShortname),
	}
}
