package mapx

import (
	"strings"
	"testing"
)

func TestSort(t *testing.T) {
	var data = map[int]string{
		1: "a",
		3: "c",
		2: "b",
	}
	result := Sort[int, string](data, func(i, j int) bool {
		return i < j
	})
	if strings.Join(result, ",") != "a,b,c" {
		t.Errorf("Sort() failed, got %v, want %v", result, "[a,b,c]")
	}
}
