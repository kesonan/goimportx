package importx

import (
	"os"
	"strings"
	"testing"

	"github.com/anqiansong/goimportx/pkg/mapx"
)

func TestSort(t *testing.T) {
	_ = Sort("/Users/sh00414ml/workspace/demo/main.go", NewImportSorter())
}

type testSorter struct {
}

func (t testSorter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (t testSorter) Sort(list []ImportPath) [][]ImportPath {
	var importPathGroup = make(map[string][]ImportPath)
	for _, importPath := range list {
		if !strings.Contains(importPath.value, "/") {
			importPathGroup[""] = append(importPathGroup[""], importPath)
			continue
		}

		list := strings.Split(importPath.value, "/")
		var key = importPath.value
		if len(key) > 0 {
			key = list[0]
		}
		importPathGroup[key] = append(importPathGroup[key], importPath)
	}

	importGroupList := mapx.Sort[string, []ImportPath](importPathGroup, func(i, j string) bool {
		return i < j
	})

	var result []string
	for idx, importGroup := range importGroupList {
		for _, importPath := range importGroup {
			result = append(result, importPath.value)
		}

		if idx < len(importGroupList)-1 {
			result = append(result, "")
		}
	}

	//fmt.Println(strings.Join(result, "\n"))
	return importGroupList
}
