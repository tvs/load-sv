package main

import (
	"os"

	_ "github.com/tvs/ultravisor/cmd"
	"github.com/tvs/ultravisor/cmd/root"
)

func main() {
	os.Exit(root.Execute())
}
