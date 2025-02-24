package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/cron"
	"github.com/ptrvsrg/crack-hash/manager/internal/di"
	"github.com/ptrvsrg/crack-hash/manager/internal/logging"
	http2 "github.com/ptrvsrg/crack-hash/manager/internal/transport/http"
	"github.com/ptrvsrg/crack-hash/manager/internal/version"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	banner = func() string {
		format := "\n" +
			"   ___             _      _  _   _   ___ _  _         __  __                             \n" +
			"  / __|_ _ __ _ __| |_   | || | /_\\ / __| || |  ___  |  \\/  |__ _ _ _  __ _ __ _ ___ _ _ \n" +
			" | (__| '_/ _` (_-< ' \\  | __ |/ _ \\\\__ \\ __ | |___| | |\\/| / _` | ' \\/ _` / _` / -_) '_|\n" +
			"  \\___|_| \\__,_/__/_||_| |_||_/_/ \\_\\___/_||_|       |_|  |_\\__,_|_||_\\__,_\\__, \\___|_|  \n" +
			"                                                                           |___/         \n" +
			"Version: %s\n"
		return fmt.Sprintf(format, version.AppVersion)
	}()

	serverCmd = &cli.Command{
		Name:                  "server",
		Aliases:               []string{"s"},
		Usage:                 "Start the server",
		Action:                runServer,
		EnableShellCompletion: true,
	}
)

func runServer(ctx context.Context, _ *cli.Command) error {
	fmt.Println(banner)

	// Load config
	cfg := config.LoadOrDie()

	// Setup logger
	logging.Setup(cfg.Server.Env)

	// Setup DI
	c := di.NewContainer(cfg)
	defer func(c *di.Container) {
		if err := c.Close(); err != nil {
			log.Error().Err(err).Stack().Msg("failed to close container")
		}
	}(c)

	// Run server
	srv := http2.NewServer(c)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Stack().Msg("failed to start server")
		}
	}()
	defer func(ctx context.Context, srv *http.Server) {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		log.Info().Msg("shutting down server")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Stack().Msg("failed to shutdown server")
		}
	}(ctx, srv)

	log.Info().Msgf("server listens on port %d", cfg.Server.Port)

	// Start cron
	scheduler := cron.NewSchedulerOrDie()

	for _, e := range c.Executors {
		if err := e.RegisterJobs(scheduler); err != nil {
			log.Fatal().Err(err).Stack().Msg("failed to register cron jobs")
		}
	}

	scheduler.Start()
	defer func(scheduler gocron.Scheduler) {
		log.Info().Msg("stop cron tasks")
		if err := scheduler.Shutdown(); err != nil {
			log.Error().Err(err).Stack().Msg("failed to shutdown cron scheduler")
		}
	}(scheduler)

	log.Info().Msg("cron scheduler started")

	// Wait for signal
	<-quit

	return nil
}
