package launchsvc

import "log/slog"

type Service struct {
	Logger *slog.Logger
}

func NewService(logger *slog.Logger) (svc Service) {

	svcShortname := "launch"

	return Service{
		Logger: logger.With("svc", svcShortname),
	}
}
