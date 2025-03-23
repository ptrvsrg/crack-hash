package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/ptrvsrg/crack-hash/manager/internal/version"
)

var (
	versionCmd = &cli.Command{
		Name:                  "version",
		Aliases:               []string{"v"},
		Usage:                 "Print the Version",
		Action:                printVersion,
		EnableShellCompletion: true,
	}
)

func printVersion(context.Context, *cli.Command) error {
	fmt.Printf("Application: %s\nRuntime: %s %s\n", version.AppVersion, version.GoVersion, version.Platform)
	return nil
}
