# goimportx

goimportx is a tool to help you manage your go imports.

## Features

- Automatically sort and group your go imports.
- Support custom group rules.
- Support write result to file.
- Only support go module.
- Use golang official sort algorithm.
- Automatically remove parentheses when there is only one import.

## Installation

```bash
$ go install github.com/anqiansong/goimportx@latest
```

## Usage

```bash
$ goimportx --file /path/to/file.go
```

## Help

```bash
goimportx --help
sort and group go imports

Usage:
  goimportx [flags]

Examples:
goimportx --file /path/to/file.go --group "system,local,third"

Flags:
  -f, --file string    file path
  -g, --group string   group rule, split by comma, only supports [system,local,third,others] (default "system,local,third")
  -h, --help           help for goimportx
  -v, --version        version for goimportx
  -w, --write          write result to (source) file instead of stdout
```

