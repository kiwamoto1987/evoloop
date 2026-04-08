package main

import (
	"os"

	"github.com/kiwamoto1987/evoloop/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
