package importx

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"os"
	"strings"

	"github.com/anqiansong/goimportx/pkg/mapx"
	"golang.org/x/tools/go/ast/astutil"
)

type SorterOption func(s *ImportSorter)

func WithWriter(writer io.Writer) SorterOption {
	return func(s *ImportSorter) {
		s.writer = writer
	}
}

type ImportSorter struct {
	writer io.Writer
}

func (i *ImportSorter) Sort(list []ImportPath) []ImportPath {
	fset := token.NewFileSet()
	file := &ast.File{}
	for _, v := range list {
		astutil.AddNamedImport(fset, file, v.name, v.value)
	}

	ast.SortImports(fset, file)

	var result []ImportPath
	for _, v := range file.Imports {
		var name string
		if v.Name != nil {
			name = v.Name.String()
		}

		value := trimQuote(v.Path.Value)
		result = append(result, ImportPath{
			name:  name,
			value: value,
		})
	}

	return result
}

func (i *ImportSorter) Write(p []byte) (n int, err error) {
	if i.writer != nil {
		return i.writer.Write(p)
	}
	return 0, nil
}

func NewImportSorter(opts ...SorterOption) *ImportSorter {
	instance := &ImportSorter{
		writer: os.Stdout,
	}

	for _, o := range opts {
		o(instance)
	}

	return instance
}

func (i *ImportSorter) Group(list []ImportPath) [][]ImportPath {
	var importPathGroup = make(map[string][]ImportPath)
	for _, importPath := range list {
		group := importPath.PackageType()
		if _, ok := groupSort[group]; ok {
			importPathGroup[group] = append(importPathGroup[group], importPath)
		} else {
			importPathGroup[groupNameOthers] = append(importPathGroup[groupNameOthers], importPath)
		}
	}

	return mapx.Sort[string, []ImportPath](importPathGroup, func(i, j string) bool {
		return groupSort[i] < groupSort[j]
	})
}

func InitGroup(s string) error {
	if len(s) == 0 {
		return nil
	}
	list := strings.FieldsFunc(s, func(r rune) bool {
		return r == ','
	})

	if len(list) == 0 {
		return nil
	}

	groupSort = map[string]int{}
	for idx, v := range list {
		_, ok := validGroupRule[v]
		if !ok {
			return fmt.Errorf("invalid group name: %s", v)
		}

		groupSort[v] = idx
	}

	var containsOthers bool
	for k := range groupSort {
		if k == groupNameOthers {
			containsOthers = true
			break
		}
	}
	if !containsOthers {
		groupSort[groupNameOthers] = len(list)
	}

	return nil
}
