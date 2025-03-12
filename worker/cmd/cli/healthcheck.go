package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"gopkg.in/resty.v1"
)

var (
	healthcheckCmd = &cli.Command{
		Name:                  "healthcheck",
		Aliases:               []string{"H"},
		Usage:                 "Healthcheck",
		Action:                healthcheck,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "host",
				Aliases:     []string{"H"},
				Usage:       "Server hostname",
				HideDefault: false,
				Required:    false,
				Local:       true,
				Value:       "0.0.0.0:8080",
			},
		},
	}
	errHealthcheckFailed = fmt.Errorf("healthcheck failed")
)

func healthcheck(_ context.Context, command *cli.Command) error {
	host := command.String("host")

	resp, err := resty.R().Get(fmt.Sprintf("http://%s/api/worker/health/readiness", host))
	if err != nil {
		return fmt.Errorf("%w: %w", errHealthcheckFailed, err)
	}

	if resp.IsError() {
		return fmt.Errorf("%w: %s", errHealthcheckFailed, resp.String())
	}

	fmt.Println("OK")

	return nil
}
