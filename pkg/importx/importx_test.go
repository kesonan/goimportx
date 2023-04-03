package importx

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestSort(t *testing.T) {
	data, err := Sort("/Users/sh00414ml/workspace/goimportx/main.go", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(os.Stdout, string(data))
}
