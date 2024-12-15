package main

import (
	"os"

	"ops_cli/cmd"
	"ops_cli/pkg/log"
)

func main() {
	log.InitLogger()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
