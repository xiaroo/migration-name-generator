package main

import (
	"context"
	"os"

	"github.com/xiaroo/migration-name-generator/internal/app"
	"github.com/xiaroo/migration-name-generator/internal/infra/clislog"
)

func main() {
	printer := clislog.New(
		os.Stdout,
		clislog.WithAppName("migname"),
		clislog.WithCommand("migname"),
		clislog.WithLinks(true),
	)

	printer.Welcome()
	exitCode := app.Run(context.Background(), os.Stdin, os.Stdout, os.Args[1:], printer)
	printer.Goodbye()

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
