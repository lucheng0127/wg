package main

import (
	"fmt"
	"os"

	"github.com/lucheng0127/wg/cmd/wgctl/app"
)

func main() {
	cmd := app.NewCliCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
