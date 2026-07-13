package cliapp

import (
	"github.com/loveyourstack/lys-ref/cmd"
)

// App is a refcli application. Defined in a separate package to avoid circular imports with rootcli and subcommands
type App struct {
	*cmd.Application
}
