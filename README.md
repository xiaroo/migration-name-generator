# migname

CLI for generating paired migration file names.

## Run locally

From the project root:

```bash
go run ./cmd/migname
```

One-shot mode:

```bash
go run ./cmd/migname Create user table
```

Output:

```text
Up: 20260514163541_create_user_table.up.sql
Down: 20260514163541_create_user_table.down.sql
```

## Install

```bash
go install ./cmd/migname
```

Then run:

```bash
migname
migname Create user table
```

If the command is not found, add Go binaries to `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Commands

Inside interactive mode:

```text
/settings  edit generation settings
/config    show active config
/reset     restore default config values
/help      show commands
/exit      close migname
```

## Config

`migname` stores config as YAML in the user config directory:

```text
~/Library/Application Support/migname/config.yaml
```

Default config:

```yaml
time_format: "YYYYMMDDHHMMSS"
name_format: "snake_case"
separator_format: "underscore"
up_suffix: "up"
down_suffix: "down"
extension: "sql"
config_path: "/Users/<user>/Library/Application Support/migname/config.yaml"
```
