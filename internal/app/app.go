package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	cliapp "github.com/xiaroo/migration-name-generator/internal/app/cli"
	"github.com/xiaroo/migration-name-generator/internal/config"
	"github.com/xiaroo/migration-name-generator/internal/infra/clislog"
	"github.com/xiaroo/migration-name-generator/internal/services/generator"
)

type App struct {
	cfg       config.Config
	generator *generator.Service
	cli       *cliapp.App
	printer   *clislog.Printer
}

func Run(ctx context.Context, in io.Reader, out io.Writer, args []string, printer *clislog.Printer) int {
	if in == nil {
		in = strings.NewReader("")
	}
	if out == nil {
		out = io.Discard
	}

	reader := bufio.NewReader(in)
	application, ok := New(reader, out, printer)
	if !ok {
		return 1
	}

	if len(args) > 0 {
		if err := application.Generate(ctx, strings.Join(args, " ")); err != nil {
			printer.CustomData("Generation error", err)
			return 1
		}
		return 0
	}

	if err := application.cli.Run(ctx); err != nil {
		printer.CustomData("CLI error", err)
		return 1
	}

	return 0
}

func New(reader *bufio.Reader, out io.Writer, printer *clislog.Printer) (*App, bool) {
	cfg, ok := loadConfigWithRecovery(reader, out, printer)
	if !ok {
		return nil, false
	}

	service := generator.New()
	application := &App{
		cfg:       cfg,
		generator: service,
		printer:   printer,
	}
	application.cli = cliapp.New(reader, out, printer, cfg, service)

	return application, true
}

func (app *App) Generate(ctx context.Context, title string) error {
	result, err := app.generator.Generate(ctx, app.cfg, title)
	if err != nil {
		return err
	}

	printMigration(app.cli.Output(), result)
	return nil
}

func loadConfigWithRecovery(reader *bufio.Reader, out io.Writer, printer *clislog.Printer) (config.Config, bool) {
	cfg, err := config.Load()
	if err == nil {
		return cfg, true
	}

	var validationErr *config.ValidationError
	if !errors.As(err, &validationErr) {
		printer.CustomData("Config error", err)
		return config.Config{}, false
	}

	if validationErr.ConfigPath == "" {
		if path, pathErr := config.Path(); pathErr == nil {
			validationErr.ConfigPath = path
		}
	}

	printer.CustomData("Config validation error", formatValidationError(validationErr))
	printer.CustomData("Config recovery", []clislog.Field{
		clislog.F("1", "regenerate config with default values"),
		clislog.F("2", "fix config manually and run migname again"),
	})

	for {
		fmt.Fprint(out, "Choose action [1/2]: ")

		answer, readErr := reader.ReadString('\n')
		if readErr != nil {
			printer.CustomData("Config error", readErr)
			return config.Config{}, false
		}

		switch strings.TrimSpace(answer) {
		case "1":
			cfg, resetErr := config.ResetDefault(validationErr.ConfigPath)
			if resetErr != nil {
				printer.CustomData("Config reset error", resetErr)
				return config.Config{}, false
			}

			printer.CustomData("Config regenerated", []string{
				fmt.Sprintf("file: %q", cfg.ConfigPath),
				"status: default values restored",
			})
			return cfg, true
		case "2":
			printer.CustomData("Manual fix", []string{
				fmt.Sprintf("file: %q", validationErr.ConfigPath),
				"next step: edit this YAML file and run migname again",
			})
			return config.Config{}, false
		default:
			fmt.Fprintln(out, "Please enter 1 or 2.")
		}
	}
}

func formatValidationError(validationErr *config.ValidationError) []string {
	if validationErr == nil {
		return []string{"config validation failed"}
	}

	lines := []string{fmt.Sprintf("file: %q", validationErr.ConfigPath)}

	for _, issue := range validationErr.Issues {
		lines = append(lines,
			fmt.Sprintf("field: %s", issue.Field),
			fmt.Sprintf("current value: %q", issue.Value),
			fmt.Sprintf("supported values: %s", strings.Join(issue.Allowed, ", ")),
		)
	}

	return lines
}

func printMigration(out io.Writer, result generator.Result) {
	fmt.Fprintf(out, "Up: %s\n", result.Up)
	fmt.Fprintf(out, "Down: %s\n", result.Down)
}
