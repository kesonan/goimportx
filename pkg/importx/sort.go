package importx

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/anqiansong/goimportx/pkg/mapx"
	"golang.org/x/tools/go/ast/astutil"
)

// ImportSorter provides functionality to sort import paths in Go files.
type ImportSorter struct{}

// Sort sorts the import paths in the given list and returns the sorted list.
// It uses the go/ast and go/token packages to parse and sort the imports.
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

// Group groups the import paths by package type and sorts them according to the groupSort rule.
// It returns a 2D slice where each slice contains import paths of the same package type.
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

// NewImportSorter creates a new instance of ImportSorter with the given options.
func NewImportSorter() *ImportSorter {
	return &ImportSorter{}
}

// InitGroup initializes the groupSort rule with the given string.
// The string should be a comma-separated list of valid group names.
// If an invalid group name is encountered, it returns an error.
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
