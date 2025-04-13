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

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	consumer2 "github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/consumer"
	commonconfig "github.com/ptrvsrg/crack-hash/commonlib/config"
	"github.com/ptrvsrg/crack-hash/commonlib/cron"
	"github.com/ptrvsrg/crack-hash/commonlib/http/server"
	"github.com/ptrvsrg/crack-hash/commonlib/logging"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/di"
	"github.com/ptrvsrg/crack-hash/manager/internal/job/hashcrack"
	"github.com/ptrvsrg/crack-hash/manager/internal/transport/http"
	"github.com/ptrvsrg/crack-hash/manager/internal/version"
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
	cfg := commonconfig.LoadOrDie[config.Config]()

	// Setup logger
	logging.Setup(cfg.Server.Env == config.EnvDev)

	// Setup DI
	c := di.NewContainer(ctx, cfg)
	defer c.Close(ctx)

	// Run server
	srv := startHTTPServer(ctx, cfg.Server, c)
	defer stopHTTPServer(ctx, srv)

	// Start cron
	scheduler := startCronScheduler(ctx, c)
	defer stopCronScheduler(ctx, scheduler)

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

func startCronScheduler(ctx context.Context, c *di.Container) *gocron.Scheduler {
	scheduler := cron.NewScheduler(
		ctx,
		hashcrack.RegisterDeleteExpiredTaskJob(c),
		hashcrack.RegisterFinishTimeoutTasksJob(c),
		hashcrack.RegisterExecutePendingTasksJob(c),
	)

	scheduler.StartAsync()

	log.Info().Msg("cron scheduler started")

	return scheduler
}

func stopCronScheduler(_ context.Context, scheduler *gocron.Scheduler) {
	log.Info().Msg("stopping cron scheduler")
	scheduler.Stop()
}

func startAMQPConsumer(ctx context.Context, c *di.Container) (*sync.WaitGroup, context.CancelFunc) {
	wg := &sync.WaitGroup{}
	consumerCtx, consumerCancel := context.WithCancel(ctx)

	for _, consumer := range c.Consumers {
		wg.Add(1)
		go func(consumer consumer2.Consumer, ctx context.Context) {
			consumer.Subscribe(ctx)
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
