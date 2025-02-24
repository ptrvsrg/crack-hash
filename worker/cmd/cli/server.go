package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/di"
	"github.com/ptrvsrg/crack-hash/worker/internal/logging"
	http2 "github.com/ptrvsrg/crack-hash/worker/internal/transport/http"
	"github.com/ptrvsrg/crack-hash/worker/internal/version"
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
			"   ___             _      _  _   _   ___ _  _        __      __       _           \n" +
			"  / __|_ _ __ _ __| |_   | || | /_\\ / __| || |  ___  \\ \\    / /__ _ _| |_____ _ _ \n" +
			" | (__| '_/ _` (_-< ' \\  | __ |/ _ \\\\__ \\ __ | |___|  \\ \\/\\/ / _ \\ '_| / / -_) '_|\n" +
			"  \\___|_| \\__,_/__/_||_| |_||_/_/ \\_\\___/_||_|         \\_/\\_/\\___/_| |_\\_\\___|_|  \n" +
			"                                                                                  \n" +
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

	// Wait for signal
	<-quit

	return nil
}
