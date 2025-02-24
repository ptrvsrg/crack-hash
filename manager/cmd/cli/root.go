package main

import (
	"context"
	"fmt"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/logging"
	"github.com/ptrvsrg/crack-hash/manager/internal/version"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"os"
	"time"
)

var (
	rootCmd = &cli.Command{
		Name:                   os.Args[0],
		Version:                version.AppVersion,
		Authors:                []any{"ptrvsrg"},
		Copyright:              fmt.Sprintf("© %d ptrvsrg", time.Now().Year()),
		Usage:                  "The cli application for Crack-Hash manager",
		UseShortOptionHandling: true,
		EnableShellCompletion:  true,
		Commands: []*cli.Command{
			serverCmd,
			healthcheckCmd,
			versionCmd,
		},
	}
)

func init() {
	logging.Setup(config.EnvDev)
}

func main() {
	err := rootCmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal().Err(err).Stack().Msg("failed to run command")
	}
}
