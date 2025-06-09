package main

import (
	"fmt"
	"os"

	"github.com/graingo/maltose/cmd/maltose/cli"
	_ "github.com/graingo/maltose/cmd/maltose/i18n"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
