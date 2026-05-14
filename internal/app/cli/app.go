package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/xiaroo/migration-name-generator/internal/config"
	"github.com/xiaroo/migration-name-generator/internal/domain/models"
	"github.com/xiaroo/migration-name-generator/internal/infra/clislog"
	"github.com/xiaroo/migration-name-generator/internal/services/generator"
)

type App struct {
	reader    *bufio.Reader
	out       io.Writer
	printer   *clislog.Printer
	cfg       config.Config
	generator *generator.Service
}

func New(
	reader *bufio.Reader,
	out io.Writer,
	printer *clislog.Printer,
	cfg config.Config,
	generatorService *generator.Service,
) *App {
	if reader == nil {
		reader = bufio.NewReader(strings.NewReader(""))
	}
	if out == nil {
		out = io.Discard
	}

	return &App{
		reader:    reader,
		out:       out,
		printer:   printer,
		cfg:       cfg,
		generator: generatorService,
	}
}

func (app *App) Output() io.Writer {
	return app.out
}

func (app *App) Run(ctx context.Context) error {
	app.printHelp()

	for {
		fmt.Fprint(app.out, "migname> ")

		input, err := app.reader.ReadString('\n')
		if err != nil && len(input) == 0 {
			fmt.Fprintln(app.out)
			return nil
		}

		trimmed := strings.TrimSpace(input)
		if trimmed == "" {
			if err != nil {
				return nil
			}
			continue
		}

		if strings.HasPrefix(trimmed, "/") {
			shouldExit, commandErr := app.handleCommand(ctx, trimmed)
			if commandErr != nil {
				app.printer.CustomData("Command error", commandErr)
			}
			if shouldExit {
				return nil
			}
		} else if commandErr := app.generate(ctx, trimmed); commandErr != nil {
			app.printer.CustomData("Generation error", commandErr)
		}

		if err != nil {
			return nil
		}
	}
}

func (app *App) handleCommand(ctx context.Context, command string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "/exit", "/quit", "/q":
		return true, nil
	case "/help", "/h":
		app.printHelp()
	case "/config":
		app.printConfig()
	case "/settings":
		return false, app.openSettings(ctx)
	case "/reset":
		cfg, err := config.ResetDefault(app.cfg.ConfigPath)
		if err != nil {
			return false, err
		}
		app.cfg = cfg
		app.printer.CustomData("Config reset", []string{
			fmt.Sprintf("file: %q", cfg.ConfigPath),
			"status: default values restored",
		})
	default:
		return false, fmt.Errorf("unknown command %q; type /help to see available commands", command)
	}

	return false, nil
}

func (app *App) generate(ctx context.Context, title string) error {
	result, err := app.generator.Generate(ctx, app.cfg, title)
	if err != nil {
		return err
	}

	fmt.Fprintf(app.out, "Up: %s\n", result.Up)
	fmt.Fprintf(app.out, "Down: %s\n", result.Down)
	return nil
}

func (app *App) printHelp() {
	app.printer.CustomData("Commands", []clislog.Field{
		clislog.F("/settings", "edit generation settings"),
		clislog.F("/config", "show active config"),
		clislog.F("/reset", "restore default config values"),
		clislog.F("/exit", "close migname"),
	})
}

func (app *App) printConfig() {
	app.printer.CustomData("Config", app.cfg)
}

func (app *App) openSettings(ctx context.Context) error {
	for {
		app.printer.CustomData("Settings", []string{
			fmt.Sprintf("1. time_format: %s", app.cfg.TimeFormat),
			fmt.Sprintf("2. name_format: %s", app.cfg.NameFormat),
			fmt.Sprintf("3. separator_format: %s", app.cfg.SeparatorFormat),
			fmt.Sprintf("4. up_suffix: %s", app.cfg.UpSuffix),
			fmt.Sprintf("5. down_suffix: %s", app.cfg.DownSuffix),
			fmt.Sprintf("6. extension: %s", app.cfg.Extension),
			"7. reset to defaults",
			fmt.Sprintf("8. config file: %q", app.cfg.ConfigPath),
			"0. back",
		})

		choice, err := app.ask("Choose setting [0-8]: ")
		if err != nil {
			return err
		}

		switch choice {
		case "0", "":
			return nil
		case "1":
			if err := app.updateTimeFormat(ctx); err != nil {
				return err
			}
		case "2":
			if err := app.updateNameFormat(ctx); err != nil {
				return err
			}
		case "3":
			if err := app.updateSeparatorFormat(ctx); err != nil {
				return err
			}
		case "4":
			if err := app.updateStringSetting("up_suffix", func(value string) {
				app.cfg.UpSuffix = value
			}); err != nil {
				return err
			}
		case "5":
			if err := app.updateStringSetting("down_suffix", func(value string) {
				app.cfg.DownSuffix = value
			}); err != nil {
				return err
			}
		case "6":
			if err := app.updateStringSetting("extension", func(value string) {
				app.cfg.Extension = value
			}); err != nil {
				return err
			}
		case "7":
			cfg, err := config.ResetDefault(app.cfg.ConfigPath)
			if err != nil {
				return err
			}
			app.cfg = cfg
			app.printer.CustomData("Settings saved", []string{"default values restored"})
		case "8":
			app.printer.CustomData("Config file", []string{fmt.Sprintf("file: %q", app.cfg.ConfigPath)})
		default:
			fmt.Fprintln(app.out, "Please choose a number from 0 to 8.")
		}
	}
}

func (app *App) updateTimeFormat(ctx context.Context) error {
	values := models.ValidTimeFormats()
	choice, ok, err := choose(app, "time_format", values, func(value models.TimeFormat) string {
		return value.String()
	})
	if err != nil || !ok {
		return err
	}

	app.cfg.TimeFormat = choice
	return app.saveSettings(ctx)
}

func (app *App) updateNameFormat(ctx context.Context) error {
	values := models.ValidNameFormats()
	choice, ok, err := choose(app, "name_format", values, func(value models.NameFormat) string {
		return value.String()
	})
	if err != nil || !ok {
		return err
	}

	app.cfg.NameFormat = choice
	return app.saveSettings(ctx)
}

func (app *App) updateSeparatorFormat(ctx context.Context) error {
	values := models.ValidSeparatorFormats()
	choice, ok, err := choose(app, "separator_format", values, func(value models.SeparatorFormat) string {
		return value.String()
	})
	if err != nil || !ok {
		return err
	}

	app.cfg.SeparatorFormat = choice
	return app.saveSettings(ctx)
}

func (app *App) updateStringSetting(fieldName string, apply func(string)) error {
	value, err := app.ask(fmt.Sprintf("Enter %s: ", fieldName))
	if err != nil {
		return err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s must not be empty", fieldName)
	}

	apply(value)
	return app.saveSettings(context.Background())
}

func (app *App) saveSettings(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := config.Save(app.cfg.ConfigPath, app.cfg); err != nil {
		return err
	}

	app.printer.CustomData("Settings saved", []string{fmt.Sprintf("file: %q", app.cfg.ConfigPath)})
	return nil
}

func choose[T any](app *App, title string, values []T, label func(T) string) (T, bool, error) {
	var zero T
	lines := make([]string, 0, len(values)+1)
	for index, value := range values {
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, label(value)))
	}
	lines = append(lines, "0. back")
	app.printer.CustomData(title, lines)

	answer, err := app.ask("Choose value: ")
	if err != nil {
		return zero, false, err
	}

	if answer == "0" || answer == "" {
		return zero, false, nil
	}

	var selected int
	if _, scanErr := fmt.Sscanf(answer, "%d", &selected); scanErr != nil {
		return zero, false, fmt.Errorf("invalid choice %q", answer)
	}

	if selected < 1 || selected > len(values) {
		return zero, false, fmt.Errorf("choice %d is out of range", selected)
	}

	return values[selected-1], true, nil
}

func (app *App) ask(prompt string) (string, error) {
	fmt.Fprint(app.out, prompt)
	value, err := app.reader.ReadString('\n')
	if err != nil && len(value) == 0 {
		return "", err
	}

	return strings.TrimSpace(value), nil
}
