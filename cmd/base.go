package cmd

import (
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/services/procsvc"
	"github.com/loveyourstack/lys-ref/internal/services/syssvc"
	"github.com/loveyourstack/lys-ref/pkg/aws/awsapi"
	"github.com/loveyourstack/lys-ref/pkg/aws/awssvc"
)

// Application contains the fields common to all commands
type Application struct {
	Config   *myapp.Config
	InfoLog  *slog.Logger
	ErrorLog *slog.Logger
	Db       *pgxpool.Pool // app-level connection for queries
	OwnerDb  *pgxpool.Pool // db owner connection for monitoring
	Validate *validator.Validate

	// clients
	AwsClient awsapi.Client

	// services
	AwsSvc  awssvc.Service
	ProcSvc procsvc.Service
	SysSvc  syssvc.Service
}

// NewApplication returns an Application with default settings. Not all fields get initialized.
func NewApplication(conf *myapp.Config) (app *Application) {

	// declare and configure logs
	var infoLog, errorLog *slog.Logger
	if conf.General.Debug {
		infoLog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		errorLog = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		infoLog = slog.New(slog.NewTextHandler(os.Stdout, nil))
		errorLog = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	return &Application{
		Config:   conf,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
		Validate: validator.New(validator.WithRequiredStructEnabled()),

		// services that can be initialized for all cmds, if any
		// ...
	}
}
