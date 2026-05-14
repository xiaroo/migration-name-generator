package clislog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	cyan    = "\033[36m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	dim     = "\033[2m"
	maxLine = 66
)

type Printer struct {
	out     io.Writer
	appName string
	command string
	color   bool
	links   bool
	now     func() time.Time
}

type Field struct {
	Key   string
	Value string
}

type Option func(*Printer)

func New(out io.Writer, opts ...Option) *Printer {
	if out == nil {
		out = os.Stdout
	}

	printer := &Printer{
		out:     out,
		appName: "migname",
		command: "migname",
		color:   true,
		links:   true,
		now:     time.Now,
	}

	for _, opt := range opts {
		opt(printer)
	}

	return printer
}

func WithAppName(name string) Option {
	return func(p *Printer) {
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			p.appName = trimmed
		}
	}
}

func WithCommand(command string) Option {
	return func(p *Printer) {
		if trimmed := strings.TrimSpace(command); trimmed != "" {
			p.command = trimmed
		}
	}
}

func WithColor(enabled bool) Option {
	return func(p *Printer) {
		p.color = enabled
	}
}

func WithLinks(enabled bool) Option {
	return func(p *Printer) {
		p.links = enabled
	}
}

func WithClock(now func() time.Time) Option {
	return func(p *Printer) {
		if now != nil {
			p.now = now
		}
	}
}

func F(key string, value any) Field {
	return Field{
		Key:   strings.TrimSpace(key),
		Value: fmt.Sprint(value),
	}
}

func (p *Printer) Welcome() {
	p.box(
		strings.ToUpper(p.appName),
		[]string{
			"Migration name generator",
			fmt.Sprintf("Run command: %s", p.command),
			"Type migration title and press Enter",
			"Commands: /settings, /help, /exit",
		},
	)
}

func (p *Printer) Goodbye() {
	timestamp := p.now().Format("15:04:05")
	fmt.Fprintln(p.out)
	fmt.Fprintf(
		p.out,
		"%s%s%s %s%s%s\n",
		p.paint(green, "done"),
		p.paint(dim, " at "),
		p.paint(dim, timestamp),
		p.paint(dim, "Goodbye from"),
		p.paint(bold+cyan, " "+p.appName),
		p.paint(dim, "."),
	)
}

func (p *Printer) CustomData(title string, value any) {
	lines := p.formatValue(value)
	if len(lines) == 0 {
		lines = []string{"empty"}
	}

	p.section(title, lines)
}

func (p *Printer) section(title string, lines []string) {
	if strings.TrimSpace(title) == "" {
		title = "Data"
	}

	fmt.Fprintln(p.out)
	fmt.Fprintf(p.out, "%s\n", p.paint(bold+yellow, title))
	for _, line := range lines {
		fmt.Fprintf(p.out, "  %s %s\n", p.paint(cyan, "-"), p.linkifyPaths(line))
	}
}

func (p *Printer) box(title string, lines []string) {
	border := "+" + strings.Repeat("-", maxLine+2) + "+"
	fmt.Fprintln(p.out, p.paint(cyan, border))
	fmt.Fprintf(p.out, "%s %s %s\n", p.paint(cyan, "|"), p.paint(bold+green, pad(title, maxLine)), p.paint(cyan, "|"))
	fmt.Fprintln(p.out, p.paint(cyan, border))

	for _, line := range lines {
		fmt.Fprintf(p.out, "%s %s %s\n", p.paint(cyan, "|"), pad(line, maxLine), p.paint(cyan, "|"))
	}

	fmt.Fprintln(p.out, p.paint(cyan, border))
	fmt.Fprintln(p.out)
}

func (p *Printer) formatValue(value any) []string {
	switch data := value.(type) {
	case nil:
		return nil
	case error:
		return splitLines(data.Error())
	case string:
		return splitLines(data)
	case []string:
		return data
	case []Field:
		return formatFields(data)
	case map[string]string:
		return formatStringMap(data)
	case map[string]any:
		return formatAnyMap(data)
	default:
		encoded, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return []string{fmt.Sprint(data)}
		}
		return splitLines(string(encoded))
	}
}

func formatFields(fields []Field) []string {
	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.Key == "" {
			lines = append(lines, field.Value)
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s", field.Key, field.Value))
	}
	return lines
}

func formatStringMap(data map[string]string) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", key, data[key]))
	}
	return lines
}

func formatAnyMap(data map[string]any) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %v", key, data[key]))
	}
	return lines
}

func splitLines(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

func (p *Printer) paint(style string, value string) string {
	if !p.color || style == "" || value == "" {
		return value
	}
	return style + value + reset
}

func (p *Printer) linkifyPaths(line string) string {
	if !p.links || line == "" {
		return line
	}

	return linkifyQuotedPaths(line)
}

func linkifyQuotedPaths(line string) string {
	var builder strings.Builder
	builder.Grow(len(line))

	for index := 0; index < len(line); {
		if line[index] != '"' {
			builder.WriteByte(line[index])
			index++
			continue
		}

		end, ok := findQuotedStringEnd(line, index+1)
		if !ok {
			builder.WriteString(line[index:])
			break
		}

		quoted := line[index : end+1]
		value, err := strconv.Unquote(quoted)
		if err != nil || !filepath.IsAbs(value) {
			builder.WriteString(quoted)
			index = end + 1
			continue
		}

		builder.WriteByte('"')
		builder.WriteString(fileLink(value))
		builder.WriteByte('"')
		index = end + 1
	}

	return builder.String()
}

func findQuotedStringEnd(line string, start int) (int, bool) {
	escaped := false
	for index := start; index < len(line); index++ {
		switch {
		case escaped:
			escaped = false
		case line[index] == '\\':
			escaped = true
		case line[index] == '"':
			return index, true
		}
	}

	return 0, false
}

func fileLink(path string) string {
	uri := url.URL{
		Scheme: "file",
		Path:   path,
	}

	return "\033]8;;" + uri.String() + "\033\\" + path + "\033]8;;\033\\"
}

func pad(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}
