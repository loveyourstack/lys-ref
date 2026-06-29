package cmd

import (
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/connectors/aws/awsapi"
	"github.com/loveyourstack/connectors/aws/awssvc"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/services/procsvc"
	"github.com/loveyourstack/lys-ref/internal/services/syssvc"
	"github.com/loveyourstack/lys/lyslog"
)

// Application contains the fields common to all commands
type Application struct {
	Config   *myapp.Config
	Logger   *slog.Logger
	Db       *pgxpool.Pool // app-level connection for queries
	OwnerDb  *pgxpool.Pool // db owner connection for monitoring
	Validate *validator.Validate

	// clients
	AwsClient *awsapi.Client

	// services
	AwsSvc  awssvc.Service
	ProcSvc procsvc.Service
	SysSvc  syssvc.Service
}

// NewApplication returns an Application with default settings. Not all fields get initialized.
func NewApplication(conf *myapp.Config) (app *Application) {

	var opts *slog.HandlerOptions
	if conf.General.Debug {
		opts = &slog.HandlerOptions{Level: slog.LevelDebug}
	} else {
		opts = &slog.HandlerOptions{}
	}

	logger := slog.New(lyslog.NewSplitStreamHandler(os.Stdout, os.Stderr, opts))

	return &Application{
		Config:   conf,
		Logger:   logger,
		Validate: validator.New(validator.WithRequiredStructEnabled()),

		// services that can be initialized for all cmds, if any
		// ...
	}
}
