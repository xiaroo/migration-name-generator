package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/xiaroo/migration-name-generator/internal/domain/models"
)

const (
	appDirName    = "migname"
	configName    = "config.yaml"
	fileMode      = 0o644
	directoryMode = 0o755
)

type Config struct {
	TimeFormat      models.TimeFormat      `yaml:"time_format" default:"YYYYMMDDHHMMSS"`
	NameFormat      models.NameFormat      `yaml:"name_format" default:"snake_case"`
	SeparatorFormat models.SeparatorFormat `yaml:"separator_format" default:"underscore"`
	UpSuffix        string                 `yaml:"up_suffix" default:"up"`
	DownSuffix      string                 `yaml:"down_suffix" default:"down"`
	Extension       string                 `yaml:"extension" default:"sql"`
	ConfigPath      string                 `yaml:"config_path" default:""`
}

type ValidationIssue struct {
	Field   string
	Value   string
	Allowed []string
}

type ValidationError struct {
	ConfigPath string
	Issues     []ValidationIssue
}

func (err *ValidationError) Error() string {
	if err == nil || len(err.Issues) == 0 {
		return "config validation failed"
	}

	if len(err.Issues) == 1 {
		issue := err.Issues[0]
		return fmt.Sprintf("%s has invalid value %q", issue.Field, issue.Value)
	}

	return fmt.Sprintf("config has %d invalid values", len(err.Issues))
}

func Default() Config {
	cfg := Config{}
	value := reflect.ValueOf(&cfg).Elem()
	typ := value.Type()

	for i := range value.NumField() {
		field := value.Field(i)
		if field.Kind() != reflect.String || !field.CanSet() {
			continue
		}

		if fallback := typ.Field(i).Tag.Get("default"); fallback != "" {
			field.SetString(fallback)
		}
	}

	return cfg
}

func Path() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(userConfigDir, appDirName, configName), nil
}

func Load() (Config, error) {
	configPath, err := Path()
	if err != nil {
		return Config{}, err
	}

	return LoadFromPath(configPath)
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}

	return cfg
}

func LoadFromPath(configPath string) (Config, error) {
	if strings.TrimSpace(configPath) == "" {
		return Config{}, fmt.Errorf("config path is empty")
	}

	configPath = filepath.Clean(configPath)

	if err := os.MkdirAll(filepath.Dir(configPath), directoryMode); err != nil {
		return Config{}, fmt.Errorf("create config directory: %w", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		if !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("stat config file: %w", err)
		}

		cfg := Default()
		cfg.ConfigPath = configPath
		if err := cfg.Validate(); err != nil {
			return Config{}, err
		}
		if err := Save(configPath, cfg); err != nil {
			return Config{}, err
		}

		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	cfg := Default()
	cfg.ConfigPath = configPath
	if err := decode(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config file: %w", err)
	}
	cfg.ConfigPath = configPath
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config file: %w", err)
	}

	return cfg, nil
}

func Save(configPath string, cfg Config) error {
	if strings.TrimSpace(configPath) == "" {
		return fmt.Errorf("config path is empty")
	}

	configPath = filepath.Clean(configPath)
	cfg.ConfigPath = configPath

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), directoryMode); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := encode(cfg)
	if err != nil {
		return err
	}

	tmpPath := configPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, fileMode); err != nil {
		return fmt.Errorf("write temp config file: %w", err)
	}

	if err := os.Rename(tmpPath, configPath); err != nil {
		return fmt.Errorf("replace config file: %w", err)
	}

	return nil
}

func ResetDefault(configPath string) (Config, error) {
	if strings.TrimSpace(configPath) == "" {
		return Config{}, fmt.Errorf("config path is empty")
	}

	configPath = filepath.Clean(configPath)
	cfg := Default()
	cfg.ConfigPath = configPath

	if err := Save(configPath, cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	issues := make([]ValidationIssue, 0, 3)

	if !cfg.TimeFormat.IsValid() {
		issues = append(issues, ValidationIssue{
			Field:   "time_format",
			Value:   cfg.TimeFormat.String(),
			Allowed: timeFormatsAsStrings(models.ValidTimeFormats()),
		})
	}

	if !cfg.NameFormat.IsValid() {
		issues = append(issues, ValidationIssue{
			Field:   "name_format",
			Value:   cfg.NameFormat.String(),
			Allowed: nameFormatsAsStrings(models.ValidNameFormats()),
		})
	}

	if !cfg.SeparatorFormat.IsValid() {
		issues = append(issues, ValidationIssue{
			Field:   "separator_format",
			Value:   cfg.SeparatorFormat.String(),
			Allowed: separatorFormatsAsStrings(models.ValidSeparatorFormats()),
		})
	}

	if len(issues) > 0 {
		return &ValidationError{
			ConfigPath: cfg.ConfigPath,
			Issues:     issues,
		}
	}

	return nil
}

func encode(cfg Config) ([]byte, error) {
	value := reflect.ValueOf(cfg)
	typ := value.Type()

	var builder strings.Builder
	builder.WriteString("# migname configuration\n")

	for i := range value.NumField() {
		fieldInfo := typ.Field(i)
		key := yamlKey(fieldInfo)
		if key == "" {
			continue
		}

		field := value.Field(i)
		if field.Kind() != reflect.String {
			return nil, fmt.Errorf("unsupported config field %s", fieldInfo.Name)
		}

		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(strconv.Quote(field.String()))
		builder.WriteByte('\n')
	}

	return []byte(builder.String()), nil
}

func decode(data []byte, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	fields := configFields(cfg)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return fmt.Errorf("line %d: expected key: value", lineNumber)
		}

		key = strings.TrimSpace(key)
		field, exists := fields[key]
		if !exists {
			continue
		}

		parsed, err := parseScalar(value)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNumber, err)
		}

		field.SetString(parsed)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func configFields(cfg *Config) map[string]reflect.Value {
	value := reflect.ValueOf(cfg).Elem()
	typ := value.Type()
	fields := make(map[string]reflect.Value, value.NumField())

	for i := range value.NumField() {
		field := value.Field(i)
		if field.Kind() != reflect.String || !field.CanSet() {
			continue
		}

		if key := yamlKey(typ.Field(i)); key != "" {
			fields[key] = field
		}
	}

	return fields
}

func yamlKey(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag == "-" {
		return ""
	}

	if key, _, ok := strings.Cut(tag, ","); ok {
		return key
	}

	if tag != "" {
		return tag
	}

	return strings.ToLower(field.Name)
}

func parseScalar(value string) (string, error) {
	trimmed := strings.TrimSpace(stripInlineComment(value))
	if trimmed == "" {
		return "", nil
	}

	if unquoted, err := strconv.Unquote(trimmed); err == nil {
		return unquoted, nil
	}

	if strings.HasPrefix(trimmed, `"`) || strings.HasPrefix(trimmed, `'`) {
		return "", fmt.Errorf("invalid quoted value %q", trimmed)
	}

	return trimmed, nil
}

func stripInlineComment(value string) string {
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for i, char := range value {
		switch {
		case escaped:
			escaped = false
		case char == '\\' && inDoubleQuote:
			escaped = true
		case char == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote
		case char == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote
		case char == '#' && !inSingleQuote && !inDoubleQuote:
			return value[:i]
		}
	}

	return value
}

func timeFormatsAsStrings(values []models.TimeFormat) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result
}

func nameFormatsAsStrings(values []models.NameFormat) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result
}

func separatorFormatsAsStrings(values []models.SeparatorFormat) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result
}
