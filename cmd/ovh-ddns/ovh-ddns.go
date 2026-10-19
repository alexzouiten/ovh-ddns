package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexzouiten/ovh-ddns/internal/config"
	"github.com/alexzouiten/ovh-ddns/internal/fetcher"
	"github.com/alexzouiten/ovh-ddns/internal/ovh"
	"github.com/alexzouiten/ovh-ddns/internal/updater"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)
	defer slog.Debug("exiting")

	if cfg.DryRun {
		slog.Info("dry run mode active")
	}

	fetcher := fetcher.NewFetcher()
	ovh, err := ovh.New(cfg.Auth, cfg.DryRun)
	if err != nil {
		slog.Error("failed to create ovh client", "err", err)
		os.Exit(1)
	}
	updater := updater.New(fetcher, ovh, cfg.Domains)

	updateTicker := time.NewTicker(cfg.Update.Interval)
	defer updateTicker.Stop()

	shutdownCh := make(chan struct{})

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Debug("received signal to shut down")
		updateTicker.Stop()
		close(shutdownCh)
	}()

	doUpdate := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Update.Timeout)
		defer cancel()

		err = updater.Update(ctx)
		if err != nil {
			slog.Error(err.Error())
		}
	}
	doUpdate()

	for {
		select {
		case <-shutdownCh:
			return
		case <-updateTicker.C:
			doUpdate()
		}
	}
}
