package importx

import (
	"io"
	"os"

	"github.com/anqiansong/goimportx/pkg/mapx"
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

func (i *ImportSorter) Sort(list []ImportPath) [][]ImportPath {
	var importPathGroup = make(map[Type][]ImportPath)
	for _, importPath := range list {
		tp := importPath.PackageType()
		importPathGroup[tp] = append(importPathGroup[tp], importPath)
	}

	return mapx.Sort[Type, []ImportPath](importPathGroup, func(i, j Type) bool {
		return pkgIndex[i] < pkgIndex[j]
	})
}
