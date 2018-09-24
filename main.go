package main

import (
	"github.com/rebuy-de/cost-exporter/cmd"
	"github.com/rebuy-de/rebuy-go-sdk/cmdutil"

	"github.com/sirupsen/logrus"
)

func main() {
	defer cmdutil.HandleExit()
	if err := cmd.NewRootCommand().Execute(); err != nil {
		logrus.Fatal(err)
	}
}
