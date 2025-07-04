package main

import (
	"os"

	"github.com/graingo/maltose/cmd/maltose/cli"
	"github.com/graingo/maltose/cmd/maltose/utils"
)

func main() {
	if err := cli.Execute(); err != nil {
		utils.PrintError(err.Error(), nil)
		os.Exit(1)
	}
}
