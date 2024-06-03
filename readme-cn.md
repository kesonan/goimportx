# goimportx

[![Go](https://github.com/anqiansong/goimportx/actions/workflows/go.yml/badge.svg)](https://github.com/anqiansong/goimportx/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-gree.svg)](https://github.com/anqiansong/goimportx/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/anqiansong/goimpoetx.svg)](https://pkg.go.dev/github.com/anqiansong/goimportx)

中文｜[English](readme.md)

goimportx 是一个 go 包导入排序的 cli 工具，其继承了 goimports 的排序逻辑，又增强了一些 goimports 没有的特新。

## 特性

- 自动排序和分组 go 包导入。
- 支持自定义分组规则。
- 支持将结果写入文件。
- 仅支持 go module。
- 使用 go 官方排序算法.
- 当只有一个 import 时自动去除括号
- 自动移除空白行

| 特性                   | goimports | goimportx |
|----------------------|-----------|-----------|
| 自动排序和分组 go 包导入       | ✅         | ✅         |    
| 支持自定义分组规则            | ❌         | ✅         |        
| 支持将结果写入文件            | ✅         | ✅         |        
| 当只有一个 import 时自动去除括号 | ❌         | ✅         |        
| 自动移除空白行              | ❌         | ✅         |     
| 支持多文件、多目录            | ❌         | ✅         |        

## 安装

```bash
$ go install github.com/anqiansong/goimportx@latest
```

## Usage

```bash
$ goimportx --dir /path/to/dir --file /path/to/file.go
```

## Help

```bash
goimportx --help
sort and group go imports

Usage:
  goimportx [flags]

Examples:
goimportx --dir path/to/your/dir --file /path/to/file.go --group "system,local,third"

Flags:
  -d, --dir strings    file directory
  -f, --file string    file path
  -g, --group string   group rule, split by comma, only supports [system,local,third,others] (default "system,local,third")
  -h, --help           help for goimportx
  -v, --version        version for goimportx
  -w, --write          write result to (source) file instead of stdout
```

