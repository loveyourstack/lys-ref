package cliapp

import (
	"github.com/loveyourstack/connectors/ecb/ecbapi"
	"github.com/loveyourstack/connectors/ecb/ecbsvc"
	"github.com/loveyourstack/connectors/maxmind/mmapi"
	"github.com/loveyourstack/connectors/maxmind/mmsvc"
	"github.com/loveyourstack/lys-ref/cmd"
)

// App is a refcli application. Defined in a separate package to avoid circular imports with rootcli and subcommands
type App struct {
	*cmd.Application

	EcbClient     ecbapi.Client
	MaxMindClient mmapi.Client

	EcbSvc     ecbsvc.Service
	MaxMindSvc mmsvc.Service
}
