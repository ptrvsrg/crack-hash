package main

import (
	"context"
	"errors"
	"fmt"
	syshttp "net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	consumer2 "github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/consumer"
	commonconfig "github.com/ptrvsrg/crack-hash/commonlib/config"
	"github.com/ptrvsrg/crack-hash/commonlib/http/server"
	"github.com/ptrvsrg/crack-hash/commonlib/logging"
	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/di"
	"github.com/ptrvsrg/crack-hash/worker/internal/transport/http"
	"github.com/ptrvsrg/crack-hash/worker/internal/version"
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
	cfg := commonconfig.LoadOrDie[config.Config]()

	// Setup logger
	logging.Setup(cfg.Server.Env == config.EnvDev)

	// Setup DI
	c := di.NewContainer(ctx, cfg)
	defer c.Close(ctx)

	// Run server
	srv := startHTTPServer(ctx, cfg.Server, c)
	defer stopHTTPServer(ctx, srv)

	// Start consumers
	wg, consumerCancel := startAMQPConsumer(ctx, c)
	defer stopAMQPConsumer(ctx, wg, consumerCancel)

	// Wait for signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	return nil
}

func startHTTPServer(_ context.Context, cfg config.ServerConfig, c *di.Container) *syshttp.Server {
	srv := server.NewHTTP2(cfg.Port, http.SetupRouter(c))

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, syshttp.ErrServerClosed) {
			log.Fatal().Err(err).Stack().Msg("failed to start server")
		}
	}()

	log.Info().Msgf("server listens on port %d", cfg.Port)

	return srv
}

func stopHTTPServer(ctx context.Context, srv *syshttp.Server) {
	log.Info().Msg("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Stack().Msg("failed to shutdown server")
	}
}

func startAMQPConsumer(ctx context.Context, c *di.Container) (*sync.WaitGroup, context.CancelFunc) {
	wg := &sync.WaitGroup{}
	consumerCtx, consumerCancel := context.WithCancel(ctx)

	for _, consumer := range c.Consumers {
		wg.Add(1)
		go func(consumer consumer2.Consumer, ctx context.Context) {
			if err := consumer.Subscribe(ctx); err != nil {
				log.Error().Err(err).Msg("failed to subscribe")
			}
			wg.Done()
		}(consumer, consumerCtx)
	}

	log.Info().Msg("AMQP consumers started")

	return wg, consumerCancel
}

func stopAMQPConsumer(_ context.Context, wg *sync.WaitGroup, cancel context.CancelFunc) {
	log.Info().Msg("stopping AMQP consumers")
	cancel()
	wg.Wait()
}
