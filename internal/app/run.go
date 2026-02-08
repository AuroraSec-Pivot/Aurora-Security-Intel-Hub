package app

import (
	"context"
	"fmt"

	"github.com/AuroraSec-Pivot/Aurora-Security-Intel-Hub/internal/config"
)

// RunOnce loads config and executes one pipeline cycle.
// MVP 阶段：这里只做加载+校验，后续接 pipeline.RunOnce。
func (a *App) RunOnce(ctx context.Context, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	a.log.Info("config loaded",
		"mode", cfg.Pipeline.Mode,
		"sources", len(cfg.Sources),
		"db", cfg.Archive.Path,
	)

	// TODO: wire dependencies:
	// - archive/sqlite (init + schema)
	// - provider/rss
	// - notifier/wecom
	// - pipeline.RunOnce(ctx, cfg, ...)
	_ = ctx

	a.log.Info("RunOnce placeholder (pipeline not wired yet)")
	return nil
}

// RunDaemon loads config and runs in daemon mode.
// MVP 阶段：先做占位，后续接 scheduler (interval + jitter)。
func (a *App) RunDaemon(ctx context.Context, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if cfg.Pipeline.Mode != "daemon" {
		a.log.Info("overriding pipeline.mode to daemon (cli flag)", "was", cfg.Pipeline.Mode)
	}

	a.log.Info("RunDaemon placeholder (scheduler not wired yet)",
		"interval", cfg.Pipeline.Interval,
		"jitter", cfg.Pipeline.Jitter,
	)

	_ = ctx
	return fmt.Errorf("daemon mode not implemented yet")
}
