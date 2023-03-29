package importx

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/anqiansong/goimportx/pkg/collection"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/astutil"
)

const (
	SystemPkg Type = 1 << iota
	LocalPkg
	ThirdPkg
)

var pkgIndex = map[Type]int{
	SystemPkg: 1,
	LocalPkg:  2,
	ThirdPkg:  3,
}

type Type int

type Sorter interface {
	Sort(list []ImportPath) [][]ImportPath
}

type ImportPath struct {
	name       string
	value      string
	use        bool
	modulePath string
}

func (ip ImportPath) PackageType() Type {
	// Inspired by https://cs.opensource.google/go/x/tools/+/master:go/ast/astutil/imports.go;l=196
	if strings.Contains(ip.value, ".") {
		return ThirdPkg
	}

	if len(ip.modulePath) > 0 && strings.HasPrefix(ip.value, ip.modulePath) {
		return LocalPkg
	}

	return SystemPkg
}

func Sort(filename string, sorter Sorter) error {
	_, err := os.Stat(filename)
	if err != nil {
		return err
	}

	var modulePath string
	if err := walkDir(filename, func(path string, d fs.DirEntry, err error) error {
		if len(modulePath) > 0 {
			return filepath.SkipAll
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".mod" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			f, err := modfile.Parse(path, data, nil)
			if err != nil {
				return err
			}
			if f.Module == nil {
				return nil
			}
			modulePath = f.Module.Mod.Path
		}

		return nil
	}); err != nil {
		return err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return err
	}

	importSet := collection.NewArraySet[ImportPath]()
	importSpecIterator(f, modulePath, func(decl *ast.GenDecl, spec ast.Spec, path ImportPath) {
		importSet.Add(path)
	})

	var importUnGroupList = importSet.List()
	deleteNamedImport(fset, f, importUnGroupList)

	var importPathList = [][]ImportPath{importUnGroupList}
	if sorter != nil {
		importPathList = sorter.Sort(importUnGroupList)
	}

	rewriteImport(fset, f, importPathList)
	ast.SortImports(fset, f)
	importGroupBounds := getImportGroupBounds(importPathList)
	addGroupGap(f, modulePath, importGroupBounds)
	if sorter != nil {
		if writer, ok := sorter.(io.Writer); ok {
			_ = format.Node(writer, fset, f)
		}
	}

	return nil
}

func walkDir(file string, fn fs.WalkDirFunc) error {
	var lastFile = file
	for {
		dir := filepath.Dir(lastFile)
		if lastFile == dir {
			return nil
		}

		fileSystem := os.DirFS(dir)
		err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return fn(path, d, err)
			}
			return fn(filepath.Join(dir, path), d, err)
		})
		if err == fs.SkipAll {
			return nil
		}
		if err != nil {
			return err
		}

		lastFile = dir
	}
}
func deleteNamedImport(fset *token.FileSet, f *ast.File, importPathList []ImportPath) {
	for _, v := range importPathList {
		astutil.DeleteNamedImport(fset, f, v.name, v.value)
	}
}

func rewriteImport(fset *token.FileSet, f *ast.File, importPathGroup [][]ImportPath) {
	for _, group := range importPathGroup {
		for _, importPath := range group {
			if !importPath.use {
				continue
			}
			astutil.AddNamedImport(fset, f, importPath.name, importPath.value)
		}
	}

	ast.SortImports(fset, f)
}

func getImportGroupBounds(importPathGroup [][]ImportPath) map[string]struct{} {
	var importGroupBounds = make(map[string]struct{})
	for _, group := range importPathGroup {
		for i := len(group) - 1; i >= 0; i-- {
			importPath := group[i]
			if !importPath.use {
				continue
			}

			importGroupBounds[importPath.value] = struct{}{}
			break
		}
	}
	return importGroupBounds
}

func importSpecIterator(f *ast.File, modulePath string, iterator func(decl *ast.GenDecl, spec ast.Spec, path ImportPath)) {
	for _, d := range f.Decls {
		decl, ok := d.(*ast.GenDecl)
		if !ok || decl.Tok != token.IMPORT {
			continue
		}

		var newSpecs []ast.Spec
		for _, spec := range decl.Specs {
			newSpecs = append(newSpecs, spec)
			imp, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}

			var name string
			if imp.Name != nil {
				name = imp.Name.String()
			}

			value := trimQuote(imp.Path.Value)

			importPath := ImportPath{
				name:       name,
				value:      value,
				use:        astutil.UsesImport(f, value),
				modulePath: modulePath,
			}
			iterator(decl, spec, importPath)
		}
	}
}

func addGroupGap(f *ast.File, modulePath string, importGroupBounds map[string]struct{}) {
	importSpecIterator(f, modulePath, func(decl *ast.GenDecl, spec ast.Spec, path ImportPath) {
		astutil.Apply(decl, nil, func(cursor *astutil.Cursor) bool {
			if cursor.Node() == nil {
				return true
			}
			if cursor.Name() != "Path" {
				return true
			}

			_, ok := cursor.Parent().(*ast.ImportSpec)
			if !ok {
				return true
			}

			basicLit, ok := cursor.Node().(*ast.BasicLit)
			if !ok {
				return true
			}

			value := trimQuote(basicLit.Value)
			if _, ok := importGroupBounds[value]; ok {
				basicLit.Value = basicLit.Value + "\n"
				cursor.Replace(basicLit)
				return true
			}

			return true
		})
	})
}

func trimQuote(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return r == '"' || r == '`'
	})
}
