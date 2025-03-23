package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"github.com/ptrvsrg/crack-hash/commonlib/logging"
	"github.com/ptrvsrg/crack-hash/worker/internal/version"
)

var (
	rootCmd = &cli.Command{
		Name:                   os.Args[0],
		Version:                version.AppVersion,
		Authors:                []any{"ptrvsrg"},
		Copyright:              fmt.Sprintf("© %d ptrvsrg", time.Now().Year()),
		Usage:                  "the cli application for Crack-Hash worker",
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
	logging.Setup(true)
}

func main() {
	err := rootCmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal().Err(err).Stack().Msg("failed to run command")
	}
}
