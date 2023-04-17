package importx

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/anqiansong/goimportx/pkg/collection"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/astutil"
)

const (
	gomodFile       = "go.mod"
	groupNameSystem = "system"
	groupNameLocal  = "local"
	groupNameThird  = "third"
	groupNameOthers = "others"

	goExt = ".go"
)

var (
	validGroupRule = map[string]struct{}{
		groupNameSystem: {},
		groupNameLocal:  {},
		groupNameThird:  {},
		groupNameOthers: {},
	}

	groupSort = map[string]int{
		groupNameSystem: 0,
		groupNameLocal:  1,
		groupNameThird:  2,
		groupNameOthers: 3,
	}
)

// Sorter is the interface that wraps the basic Group and Sort method.
type Sorter interface {
	Group(list []ImportPath) [][]ImportPath
	Sort(list []ImportPath) []ImportPath
}

type commentGroup struct {
	doc, comment *ast.CommentGroup
}

type commentGroups []*ast.CommentGroup

// ImportPath represents an import path with its name, value, use status, module path, and comment group.
type ImportPath struct {
	name         string
	value        string
	use          bool
	modulePath   string
	commentGroup *commentGroup
}

func (cg commentGroups) in(comment *ast.CommentGroup) bool {
	for _, v := range cg {
		if v == nil {
			continue
		}
		if comment.Pos() >= v.Pos() && comment.End() <= v.End() {
			return true
		}
	}

	return false
}

// PackageType returns the package type of the import path.
func (ip ImportPath) PackageType() string {
	// Inspired by https://cs.opensource.google/go/x/tools/+/master:go/ast/astutil/imports.go;l=196
	if strings.Contains(ip.value, ".") {
		return groupNameThird
	}

	if len(ip.modulePath) > 0 && strings.HasPrefix(ip.value, ip.modulePath) {
		return groupNameLocal
	}

	return groupNameSystem
}

// Sort reads the go.mod file in the same directory as the given file to determine the module path.
// The sorted imports will be written back to the file.
// If sorter implements io.Writer, the formatted file will be written to it instead of being written back to the file.
// If sorter is nil, use the default.
func Sort(filename string, sorter Sorter) ([]byte, error) {
	if sorter == nil {
		sorter = &ImportSorter{}
	}

	abs, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	filename = abs
	ext := filepath.Ext(filename)
	if ext != goExt {
		return nil, fmt.Errorf("expected go files, invalid file extension: %s", ext)
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("expected go files, invalid file path: %s", filename)
	}

	moduleFilename := getGoModFile(filename)
	if len(moduleFilename) == 0 {
		return nil, fmt.Errorf("can not find go.mod file")
	}

	data, err := os.ReadFile(moduleFilename)
	if err != nil {
		return nil, err
	}

	modFile, err := modfile.Parse(moduleFilename, data, nil)
	if err != nil {
		return nil, err
	}

	if modFile.Module == nil {
		return nil, fmt.Errorf("invalid go.mod file: %s", moduleFilename)
	}

	modulePath := modFile.Module.Mod.Path

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	importSet := collection.NewArraySet[ImportPath]()
	importComment := make(map[string]*commentGroup)
	var commentGroups commentGroups
	importSpecIterator(f, modulePath, func(decl *ast.GenDecl, spec ast.Spec, path ImportPath) {
		importSet.Add(path)
		importComment[fmt.Sprintf("%s %s", path.name, path.value)] = path.commentGroup
		commentGroups = append(commentGroups, path.commentGroup.doc, path.commentGroup.comment)
	})

	var specs []ast.Spec
	var importUnGroupList = importSet.List()
	var groupedImports = sorter.Group(importUnGroupList)
	for idx, v := range groupedImports {
		sortedImports := sorter.Sort(v)
		for _, v := range sortedImports {
			key := fmt.Sprintf("%s %s", v.name, v.value)
			cg := importComment[key]
			var doc, comment string
			if cg != nil {
				doc = getCommentGroupString(cg.doc)
				comment = getCommentGroupString(cg.comment)
			}

			if len(doc) > 0 {
				specs = append(specs, &ast.ImportSpec{
					Path: &ast.BasicLit{Value: fmt.Sprintf("%s%s", "", doc), Kind: token.STRING},
				})
			}

			var spec = ast.ImportSpec{
				Path: &ast.BasicLit{Value: fmt.Sprintf(`"%s"%s`, v.value, comment), Kind: token.STRING},
			}
			if len(v.name) > 0 {
				spec.Name = ast.NewIdent(v.name)
			}

			specs = append(specs, &spec)
		}
		if idx < len(groupedImports)-1 {
			specs = append(specs, &ast.ImportSpec{
				Path: &ast.BasicLit{Value: "", Kind: token.STRING},
			})
		}
	}

	rewriteImport(f, specs)
	deletedOriginImportCommentGroup(f, commentGroups)
	var buffer = bytes.NewBuffer(nil)
	_ = printer.Fprint(buffer, fset, f)

	result, err := format.Source(buffer.Bytes())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func deletedOriginImportCommentGroup(f *ast.File, originCommentGroup commentGroups) {
	var comments []*ast.CommentGroup
	for _, d := range f.Comments {
		if d == nil {
			continue
		}

		if !originCommentGroup.in(d) {
			comments = append(comments, d)
		}
	}
	f.Comments = comments
}

func getCommentGroupString(commentGroup *ast.CommentGroup) string {
	if commentGroup == nil {
		return ""
	}

	var list []string
	for _, v := range commentGroup.List {
		list = append(list, v.Text)
	}

	return " " + strings.Join(list, " ")
}

func getGoModFile(file string) string {
	var lastFile = filepath.Clean(file)
	for {
		dir := filepath.Dir(lastFile)
		if lastFile == dir {
			return ""
		}

		expectedGoModFile := filepath.Join(dir, gomodFile)
		if _, err := os.Stat(expectedGoModFile); err == nil {
			return expectedGoModFile
		}

		lastFile = dir
	}
}

func rewriteImport(f *ast.File, specs []ast.Spec) {
	var written bool
	var decls []ast.Decl
	for _, d := range f.Decls {
		decl, ok := d.(*ast.GenDecl)
		if ok && !written && isImportDecl(decl) {
			decl.Specs = specs
			if len(specs) == 1 {
				decl.Lparen = 0
				decl.Rparen = 0
			}
			written = true
			decls = append(decls, decl)
		} else {
			decls = append(decls, d)
		}
	}

	f.Decls = decls
}

func isImportDecl(decl *ast.GenDecl) bool {
	for _, spec := range decl.Specs {
		_, ok := spec.(*ast.ImportSpec)
		if !ok {
			return false
		}
	}
	return true
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
				commentGroup: &commentGroup{
					doc:     imp.Doc,
					comment: imp.Comment,
				},
			}
			iterator(decl, spec, importPath)
		}
	}
}

func trimQuote(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return r == '"' || r == '`'
	})
}
