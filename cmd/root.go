package cmd

import (
	"github.com/rebuy-de/cost-exporter/pkg/config"
	"github.com/rebuy-de/cost-exporter/pkg/prom"
	"github.com/rebuy-de/cost-exporter/pkg/retriever"
	"github.com/rebuy-de/rebuy-go-sdk/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type App struct {
	config string
	port   string
}

func (app *App) Run(cmd *cobra.Command, args []string) {
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

func (app *App) Bind(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(
		&app.config, "config", "c", "", `Path to configuration file.`)
	cmd.PersistentFlags().StringVarP(
		&app.port, "port", "p", "8080", `Port to bind to.`)
}

func NewRootCommand() *cobra.Command {
	cmd := cmdutil.NewRootCommand(new(App))
	cmd.Short = "AWS billing data Prometheus exporter."
	cmd.Use = "cost-exporter"
	return cmd
}
