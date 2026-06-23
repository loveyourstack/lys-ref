package syssvc

import "log/slog"

type Service struct {
	InfoLog  *slog.Logger
	ErrorLog *slog.Logger
}

func NewService(infoLog, errorLog *slog.Logger) (svc Service) {

	svcShortname := "sys"

	return Service{
		InfoLog:  infoLog.With("svc", svcShortname),
		ErrorLog: errorLog.With("svc", svcShortname),
	}
}
