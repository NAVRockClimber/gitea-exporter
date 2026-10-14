package main

import (
	"context"
	"flag"
	"gitea-exporter/internal/application"
	"gitea-exporter/internal/infrastructure/config"
	"gitea-exporter/internal/infrastructure/gitea"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "err", err)
		os.Exit(1)
	}

	var targets []*application.Target
	for name, server := range cfg.Servers {
		client := gitea.New(server.URL, server.Token)
		service := application.NewService(server.ExcludeOrgs)
		targets = append(targets, application.NewTarget(name, client, service))
	}

	runner := application.NewRunner(cfg.ScrapeInterval, targets...)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner.Run(ctx)
}
