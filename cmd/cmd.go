package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/anqiansong/goimportx/pkg/importx"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "goimportx",
	Short:   "sort and group go imports",
	Example: `goimportx --dir path/to/your/dir --file /path/to/file.go --group "system,local,third"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := importx.InitGroup(group); err != nil {
			return err
		}

		var list []string
		for _, dir := range dirs {
			disFiles, err := getGoFiles(dir)
			if err != nil {
				return err
			}
			list = append(list, disFiles...)
		}

		for _, file := range files {
			list = append(list, strings.FieldsFunc(file, func(r rune) bool {
				return r == ',' || r == '|'
			})...)
		}

		for _, file := range list {
			result, err := importx.Sort(file, nil)
			if err != nil {
				return err
			}

			if write {
				_ = os.WriteFile(file, result, 0644)
			} else {
				_, _ = fmt.Fprint(os.Stdout, string(result))
			}
		}

		return nil
	},
}

var files []string
var dirs []string
var group string
var write bool

func init() {
	rootCmd.Flags().StringSliceVarP(&files, "file", "f", nil, "file path")
	rootCmd.Flags().StringSliceVarP(&dirs, "dir", "d", nil, "file directory")
	rootCmd.Flags().StringVarP(&group, "group", "g", "system,local,third", "group rule, split by comma, only supports [system,local,third,others]")
	rootCmd.Flags().BoolVarP(&write, "write", "w", false, "write result to (source) file instead of stdout")
}

func Execute() {
	rootCmd.Version = fmt.Sprintf(
		"%s %s/%s", "v0.0.1",
		runtime.GOOS, runtime.GOARCH)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func getGoFiles(dir string) ([]string, error) {
	var files []string
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(abs, func(path string, info fs.FileInfo, err error) error {
		if path == abs {
			return nil
		}
		if info.IsDir() {
			subFiles, err := getGoFiles(path)
			if err == nil {
				files = append(files, subFiles...)
			}
		} else if filepath.Ext(path) == ".go" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
