package retriever

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/rebuy-de/cost-exporter/pkg/config"
	"github.com/rebuy-de/cost-exporter/pkg/prom"
	"github.com/rebuy-de/cost-exporter/pkg/utils"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

type CostRetriever struct {
	Accounts []config.Account
	Cron     string
	Services map[string]*costexplorer.CostExplorer
}

func (c *CostRetriever) Run() {
	c.initialize()
	c.getCosts()
	c.scheduleCron()
}

func (c *CostRetriever) initialize() {
	c.Services = make(map[string]*costexplorer.CostExplorer)
	for _, account := range c.Accounts {
		opts := session.Options{
			Config: aws.Config{
				Credentials: credentials.NewStaticCredentials(
					account.ID,
					account.Secret,
					"",
				)}}
		sess := session.Must(session.NewSessionWithOptions(opts))
		svc := costexplorer.New(sess)

		c.Services[account.Name] = svc
	}
}

func (c *CostRetriever) scheduleCron() {
	cron := cron.New()
	cron.AddFunc(c.Cron, c.getCosts)
	cron.Start()
}

func (c *CostRetriever) getCosts() {
	for _, account := range c.Accounts {
		c.getReservationCoverage(account)
		c.getReservationUtilization(account)
		c.getCostsByService(account)
	}
}

func (c *CostRetriever) getCostsByService(account config.Account) {
	logrus.Infof("Getting costs for account '%s'", account.Name)
	svc := c.Services[account.Name]
	respCost, err := svc.GetCostAndUsage((&costexplorer.GetCostAndUsageInput{
		Metrics: []*string{aws.String("BlendedCost")},
		// Getting the cost from 2 days ago is a workaround because getting data
		// for yesterday yielded unstable numbers:
		TimePeriod:  utils.GetIntervalForPastDay(2),
		Granularity: aws.String("DAILY"),
		GroupBy: []*costexplorer.GroupDefinition{
			&costexplorer.GroupDefinition{
				Key:  aws.String("SERVICE"),
				Type: aws.String("DIMENSION"),
			},
		},
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
		if *cost.Keys[0] == "Tax" {
			continue
		}

		prom.C.SetCosts(account.Name, *cost.Keys[0], amount)
	}
}

func (c *CostRetriever) getReservationCoverage(account config.Account) {
	logrus.Infof("Getting reservation coverage for account '%s'", account.Name)
	svc := c.Services[account.Name]

	respReservation, err := svc.GetReservationCoverage(&costexplorer.GetReservationCoverageInput{
		Granularity: aws.String("DAILY"),
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

func (c *CostRetriever) getReservationUtilization(account config.Account) {
	logrus.Infof("Getting reservation utilization for account '%s'", account.Name)
	svc := c.Services[account.Name]

	respReservation, err := svc.GetReservationUtilization(&costexplorer.GetReservationUtilizationInput{
		Granularity: aws.String("DAILY"),
		// Unfortunately there is no newer data then 3 days ago.
		TimePeriod: utils.GetIntervalForPastDay(3),
	})
	if err != nil {
		logrus.Fatal(err)
	}

	coveragePercent, err := strconv.ParseFloat(*respReservation.Total.UtilizationPercentage, 64)
	if err != nil {
		logrus.Fatal(err)
	}

	prom.C.SetReservationCoverage(account.Name, coveragePercent)
}
