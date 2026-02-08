package app

import "github.com/AuroraSec-Pivot/Aurora-Security-Intel-Hub/internal/observability"

type Options struct {
	Logger observability.Logger
}

type App struct {
	log observability.Logger
}

func New(opts Options) *App {
	log := opts.Logger
	if log == nil {
		log = observability.NewLogger(observability.LoggerConfig{Level: "info"})
	}
	return &App{log: log}
}
