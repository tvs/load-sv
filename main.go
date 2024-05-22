package main

import (
	"os"

	_ "github.com/tvs/ultravisor/cmd"
	"github.com/tvs/ultravisor/cmd/root"
)

func main() {
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
