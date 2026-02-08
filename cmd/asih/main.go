package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/AuroraSec-Pivot/Aurora-Security-Intel-Hub/internal/app"
	"github.com/AuroraSec-Pivot/Aurora-Security-Intel-Hub/internal/observability"
	"github.com/AuroraSec-Pivot/Aurora-Security-Intel-Hub/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println(version.String())
		return
	case "run":
		runCmd(os.Args[2:])
		return
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func runCmd(args []string) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		configPath = fs.String("config", "configs/config.yaml", "config file path")
		mode       = fs.String("mode", "once", "run mode: once|daemon")
		timeout    = fs.Duration("timeout", 2*time.Minute, "overall run timeout (once mode)")
	)
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	logger := observability.NewLogger(observability.LoggerConfig{
		Level: "info",
	})
	logger.Info("starting asih", "mode", *mode, "config", *configPath, "version", version.String())

	a := app.New(app.Options{
		Logger: logger,
	})

	ctx := context.Background()
	if *mode == "once" {
		cctx, cancel := context.WithTimeout(ctx, *timeout)
		defer cancel()

		if err := a.RunOnce(cctx, *configPath); err != nil {
			logger.Error("run once failed", "err", err)
			os.Exit(1)
		}
		logger.Info("run once completed")
		return
	}

	if *mode == "daemon" {
		if err := a.RunDaemon(ctx, *configPath); err != nil {
			logger.Error("run daemon failed", "err", err)
			os.Exit(1)
		}
		return
	}

	logger.Error("invalid mode", "mode", *mode)
	os.Exit(2)
}

func usage() {
	fmt.Println(`Aurora Security Intel Hub (ASIH)

Usage:
  asih <command> [flags]

Commands:
  run       Run the pipeline (once or daemon)
  version   Print version info
  help      Show this help

Examples:
  asih version
  asih run --mode once --config configs/config.yaml
  asih run --mode daemon --config configs/config.yaml
`)
}
