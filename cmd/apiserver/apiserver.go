package main

import (
	"github.com/lucheng0127/wg/cmd/apiserver/app"
)

func main() {
	cmd := app.NewApiserverCmd()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
