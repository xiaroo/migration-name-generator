package generator

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/xiaroo/migration-name-generator/internal/config"
	"github.com/xiaroo/migration-name-generator/internal/domain/models"
)

type Service struct {
	now func() time.Time
}

type Result struct {
	Up   string
	Down string
}

type Option func(*Service)

func New(opts ...Option) *Service {
	service := &Service{
		now: time.Now,
	}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func WithClock(now func() time.Time) Option {
	return func(service *Service) {
		if now != nil {
			service.now = now
		}
	}
}

func (service *Service) Generate(ctx context.Context, cfg config.Config, title string) (Result, error) {
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
	}

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return Result{}, fmt.Errorf("migration title must not be empty")
	}

	if err := cfg.Validate(); err != nil {
		return Result{}, err
	}

	version, err := formatTime(service.now(), cfg.TimeFormat)
	if err != nil {
		return Result{}, err
	}

	name, err := FormatName(trimmedTitle, cfg.NameFormat)
	if err != nil {
		return Result{}, err
	}

	separator, err := separatorValue(cfg.SeparatorFormat)
	if err != nil {
		return Result{}, err
	}

	baseName := version + separator + name

	return Result{
		Up:   baseName + fileTail(cfg.UpSuffix, cfg.Extension),
		Down: baseName + fileTail(cfg.DownSuffix, cfg.Extension),
	}, nil
}

func formatTime(value time.Time, format models.TimeFormat) (string, error) {
	layouts := map[models.TimeFormat]string{
		models.TimeFormatYYYYMMDDHHMMSS:               "20060102150405",
		models.TimeFormatYYYYMMDD_HHMMSS:              "20060102_150405",
		models.TimeFormatYYYY_MM_DD_HH_MM_SS:          "2006_01_02_15_04_05",
		models.TimeFormatYYYYDashMMDashDD_HHMMSS:      "2006-01-02 150405",
		models.TimeFormatYYYYDashMMDashDD_HH_MM_SS:    "2006-01-02 15:04:05",
		models.TimeFormatYYYYDashMMDashDDTHHMMSS:      "2006-01-02T150405",
		models.TimeFormatYYYYDashMMDashDDTHH_MM_SS:    "2006-01-02T15:04:05",
		models.TimeFormatYYYYSlashMMSlashDD_HH_MM_SS:  "2006/01/02 15:04:05",
		models.TimeFormatYYYYDotMMDotDD_HH_MM_SS:      "2006.01.02 15:04:05",
		models.TimeFormatDDMMYYYYHHMMSS:               "02012006150405",
		models.TimeFormatDD_MM_YYYY_HH_MM_SS:          "02_01_2006_15_04_05",
		models.TimeFormatDDDashMMDashYYYY_HH_MM_SS:    "02-01-2006 15:04:05",
		models.TimeFormatDDSplashMMSlashYYYY_HH_MM_SS: "02/01/2006 15:04:05",
		models.TimeFormatDDDotMMDotYYYY_HH_MM_SS:      "02.01.2006 15:04:05",
		models.TimeFormatMMDDYYYYHHMMSS:               "01022006150405",
		models.TimeFormatMM_DD_YYYY_HH_MM_SS:          "01_02_2006_15_04_05",
		models.TimeFormatMMDashDDDashYYYY_HH_MM_SS:    "01-02-2006 15:04:05",
		models.TimeFormatMMSlashDDSlashYYYY_HH_MM_SS:  "01/02/2006 15:04:05",
		models.TimeFormatFilenameSafe:                 "2006-01-02_15-04-05",
		models.TimeFormatISO8601:                      "2006-01-02T15:04:05",
		models.TimeFormatRFC3339:                      time.RFC3339,
	}

	layout, ok := layouts[format]
	if !ok {
		return "", fmt.Errorf("unsupported time format %q", format)
	}

	return value.Format(layout), nil
}

func FormatName(title string, format models.NameFormat) (string, error) {
	words := splitWords(title)
	if len(words) == 0 {
		return "", fmt.Errorf("migration title must contain letters or digits")
	}

	switch format {
	case models.NameFormatCamelCase:
		return words[0] + titleWords(words[1:], ""), nil
	case models.NameFormatPascalCase:
		return titleWords(words, ""), nil
	case models.NameFormatSnakeCase:
		return strings.Join(words, "_"), nil
	case models.NameFormatScreamingSnakeCase:
		return strings.ToUpper(strings.Join(words, "_")), nil
	case models.NameFormatKebabCase:
		return strings.Join(words, "-"), nil
	case models.NameFormatScreamingKebabCase:
		return strings.ToUpper(strings.Join(words, "-")), nil
	case models.NameFormatTrainCase:
		return titleWords(words, "-"), nil
	case models.NameFormatDotCase:
		return strings.Join(words, "."), nil
	case models.NameFormatPathCase:
		return strings.Join(words, "/"), nil
	case models.NameFormatFlatCase:
		return strings.Join(words, ""), nil
	case models.NameFormatUpperFlatCase:
		return strings.ToUpper(strings.Join(words, "")), nil
	case models.NameFormatTitleCase:
		return titleWords(words, " "), nil
	case models.NameFormatSentenceCase:
		return titleWord(words[0]) + joinWithPrefix(words[1:], " "), nil
	case models.NameFormatLowerCase:
		return strings.Join(words, " "), nil
	case models.NameFormatUpperCase:
		return strings.ToUpper(strings.Join(words, " ")), nil
	default:
		return "", fmt.Errorf("unsupported name format %q", format)
	}
}

func splitWords(value string) []string {
	words := make([]string, 0)
	current := make([]rune, 0)
	var previous rune

	flush := func() {
		if len(current) == 0 {
			return
		}
		words = append(words, strings.ToLower(string(current)))
		current = current[:0]
	}

	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			flush()
			previous = 0
			continue
		}

		if len(current) > 0 && startsNewWord(previous, char) {
			flush()
		}

		current = append(current, char)
		previous = char
	}

	flush()

	return words
}

func startsNewWord(previous rune, current rune) bool {
	if previous == 0 {
		return false
	}

	if unicode.IsLower(previous) && unicode.IsUpper(current) {
		return true
	}

	if unicode.IsLetter(previous) && unicode.IsDigit(current) {
		return true
	}

	return unicode.IsDigit(previous) && unicode.IsLetter(current)
}

func titleWords(words []string, separator string) string {
	result := make([]string, 0, len(words))
	for _, word := range words {
		result = append(result, titleWord(word))
	}
	return strings.Join(result, separator)
}

func titleWord(word string) string {
	if word == "" {
		return word
	}

	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func joinWithPrefix(words []string, separator string) string {
	if len(words) == 0 {
		return ""
	}

	return separator + strings.Join(words, separator)
}

func separatorValue(format models.SeparatorFormat) (string, error) {
	switch format {
	case models.SeparatorFormatUnderscore:
		return "_", nil
	case models.SeparatorFormatDash:
		return "-", nil
	case models.SeparatorFormatSlash:
		return "/", nil
	case models.SeparatorFormatDot:
		return ".", nil
	default:
		return "", fmt.Errorf("unsupported separator format %q", format)
	}
}

func fileTail(suffix string, extension string) string {
	return dotPrefix(suffix) + dotPrefix(extension)
}

func dotPrefix(value string) string {
	if value == "" {
		return ""
	}

	if strings.HasPrefix(value, ".") {
		return value
	}

	return "." + value
}
