package retriever

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	ceConfig "github.com/rebuy-de/cost-exporter/pkg/config"
	"github.com/rebuy-de/cost-exporter/pkg/prom"
	"github.com/rebuy-de/cost-exporter/pkg/utils"
)

type CostRetriever struct {
	Accounts []ceConfig.Account
	Cron     string
	Services map[string]*costexplorer.Client
}

func (c *CostRetriever) Run(ctx context.Context) {
	c.initialize(ctx)
	c.getCosts()
	c.scheduleCron()
}

func (c *CostRetriever) initialize(ctx context.Context) {
	c.Services = make(map[string]*costexplorer.Client)
	for _, account := range c.Accounts {
		credProvider := credentials.NewStaticCredentialsProvider(account.ID, account.Secret, "")
		conf, err := config.LoadDefaultConfig(ctx,
			config.WithCredentialsProvider(credProvider),
			config.WithRegion("eu-west-1"),
		)
		if err != nil {
			fmt.Println(err)
		}

		svc := costexplorer.NewFromConfig(conf)

		c.Services[account.Name] = svc
	}
}

func (c *CostRetriever) scheduleCron() {
	cron := cron.New()
	cron.AddFunc(c.Cron, c.getCosts)
	cron.Start()
}

func (c *CostRetriever) getCosts() {
	ctx := context.Background()
	for _, account := range c.Accounts {
		c.getReservationCoverage(ctx, account)
		c.getReservationUtilization(ctx, account)
		c.getCostsByService(ctx, account)
	}
}

func (c *CostRetriever) getCostsByService(ctx context.Context, account ceConfig.Account) {
	logrus.Infof("Getting costs for account '%s'", account.Name)
	svc := c.Services[account.Name]
	respCost, err := svc.GetCostAndUsage(ctx, (&costexplorer.GetCostAndUsageInput{
		Metrics: []string{"BlendedCost"},
		// Getting the cost from 2 days ago is a workaround because getting data
		// for yesterday yielded unstable numbers:
		TimePeriod:  utils.GetIntervalForPastDay(2),
		Granularity: types.GranularityDaily,
		GroupBy: []types.GroupDefinition{{
			Key:  aws.String("SERVICE"),
			Type: types.GroupDefinitionTypeDimension,
		}},
	}))
	if err != nil {
		logrus.Fatal(err)
	}

	for _, cost := range respCost.ResultsByTime[0].Groups {
		amount, err := strconv.ParseFloat(*cost.Metrics["BlendedCost"].Amount, 64)
		if err != nil {
			logrus.Fatal(err)
		}

		// ignore taxes because they are not attributed to billing in real-time
		if cost.Keys[0] == "Tax" {
			continue
		}

		prom.C.SetCosts(account.Name, cost.Keys[0], amount)
	}
}

func (c *CostRetriever) getReservationCoverage(ctx context.Context, account ceConfig.Account) {
	logrus.Infof("Getting reservation coverage for account '%s'", account.Name)
	svc := c.Services[account.Name]

	respReservation, err := svc.GetReservationCoverage(ctx, &costexplorer.GetReservationCoverageInput{
		Granularity: types.GranularityDaily,
		// Unfortunately there is no newer data then 3 days ago.
		TimePeriod: utils.GetIntervalForPastDay(3),
	})
	if err != nil {
		logrus.Fatal(err)
	}

	coveragePercent, err := strconv.ParseFloat(*respReservation.Total.CoverageHours.CoverageHoursPercentage, 64)
	if err != nil {
		logrus.Fatal(err)
	}

	prom.C.SetReservationCoverage(account.Name, coveragePercent)
}

func (c *CostRetriever) getReservationUtilization(ctx context.Context, account ceConfig.Account) {
	logrus.Infof("Getting reservation utilization for account '%s'", account.Name)
	svc := c.Services[account.Name]

	respReservation, err := svc.GetReservationUtilization(ctx, &costexplorer.GetReservationUtilizationInput{
		Granularity: types.GranularityDaily,
		// Unfortunately there is no newer data then 3 days ago.
		TimePeriod: utils.GetIntervalForPastDay(3),
	})
	if err != nil {
		logrus.Fatal(err)
	}

	utilizationPercent, err := strconv.ParseFloat(*respReservation.Total.UtilizationPercentage, 64)
	if err != nil {
		logrus.Fatal(err)
	}

	prom.C.SetReservationUtilization(account.Name, utilizationPercent)
}
