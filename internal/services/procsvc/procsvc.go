package procsvc

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/loveyourstack/lys-ref/internal/myapp"
)

type Service struct {
	Conf myapp.Process

	InfoLog  *slog.Logger
	ErrorLog *slog.Logger
}

func NewService(conf myapp.Process, infoLog, errorLog *slog.Logger) (svc Service) {

	if conf.CliCmdPrefix == "" {
		log.Fatalf("procsvc: conf.CliCmdPrefix is required")
	}

	svcShortname := "proc"

	return Service{
		Conf: conf,

		InfoLog:  infoLog.With("svc", svcShortname),
		ErrorLog: errorLog.With("svc", svcShortname),
	}
}

func (svc Service) prefixAndSplitCmd(cmd string) (splitCmd []string, err error) {
	pointCmd := strings.TrimSpace(fmt.Sprintf("%s%s", svc.Conf.CliCmdPrefix, cmd))

	name, argLine, found := strings.Cut(pointCmd, " ")
	if !found {
		return []string{name}, nil
	}

	args := strings.Fields(argLine)

	minArgs := 2
	if len(args) < minArgs {
		return []string{name}, fmt.Errorf("command: %s has %d args, expected at least %d", pointCmd, len(args), minArgs)
	}

	return append([]string{name}, args...), nil
}
