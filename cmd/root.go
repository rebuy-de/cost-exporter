package cmd

import (
	"context"

	"github.com/rebuy-de/cost-exporter/pkg/config"
	"github.com/rebuy-de/cost-exporter/pkg/prom"
	"github.com/rebuy-de/cost-exporter/pkg/retriever"
	"github.com/rebuy-de/rebuy-go-sdk/v3/pkg/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	app := new(App)

	return cmdutil.New(
		"cost-exporter", "AWS billing data Prometheus exporter.",
		app.Bind,
		cmdutil.WithLogToGraylog(),
		cmdutil.WithLogVerboseFlag(),
		cmdutil.WithVersionCommand(),
		cmdutil.WithVersionLog(logrus.DebugLevel),
		cmdutil.WithRun(app.Run),
	)
}

type App struct {
	config string
	port   string
}

func (app *App) Run(ctx context.Context, cmd *cobra.Command, args []string) {
	if app.config == "" {
		logrus.Fatal("Configuration file location not defined.")
	}

	config := config.Parse(app.config)

	APIRetriever := retriever.APIRetriever{
		Accounts:    config.Accounts,
		IntervalSec: config.Settings.CoresInterval,
	}
	APIRetriever.Run()

	costRetriever := retriever.CostRetriever{
		Accounts: config.Accounts,
		Cron:     config.Settings.CostCron,
	}
	costRetriever.Run()

	prom.Run(app.port)

	select {}
}

func (app *App) Bind(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringVarP(
		&app.config, "config", "c", "", `Path to configuration file.`)
	cmd.PersistentFlags().StringVarP(
		&app.port, "port", "p", "8080", `Port to bind to.`)

	return nil
}
