package retriever

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rebuy-de/cost-exporter/pkg/config"
	"github.com/rebuy-de/cost-exporter/pkg/prom"
	"github.com/sirupsen/logrus"
)

type CoreRetriever struct {
	Accounts    []config.Account
	IntervalSec int64
	Services    []Service
}

type Service struct {
	Region  string
	svc     *ec2.EC2
	Account string
}

func (c *CoreRetriever) Run() {
	c.initialize()
	c.getCores()
	go c.scheduleInterval()
}

func (c *CoreRetriever) initialize() {
	for _, account := range c.Accounts {
		regions := endpoints.AwsPartition().Services()[endpoints.Ec2ServiceID].Regions()
		for regionName := range regions {
			opts := session.Options{
				Config: aws.Config{
					Credentials: credentials.NewStaticCredentials(
						account.ID,
						account.Secret,
						"",
					)}}
			sess := session.Must(session.NewSessionWithOptions(opts))
			svc := ec2.New(sess, aws.NewConfig().WithRegion(regionName))

			c.Services = append(c.Services, Service{
				Region:  regionName,
				svc:     svc,
				Account: account.Name,
			})
		}
	}
}

func (c *CoreRetriever) scheduleInterval() {
	for range time.Tick(time.Duration(c.IntervalSec) * time.Second) {
		c.getCores()
	}
}

func (c *CoreRetriever) getCores() {
	for _, service := range c.Services {
		logrus.Infof("Getting cores for account '%s' and region '%s'", service.Account, service.Region)
		var totalCoreCount int64
		params := &ec2.DescribeInstancesInput{}
		resp2, err := service.svc.DescribeInstances(params)
		if err != nil {
			fmt.Println(err)
		}
		for _, reservation := range resp2.Reservations {
			for _, instance := range reservation.Instances {
				if *instance.State.Name == "running" {
					totalCoreCount = totalCoreCount + *instance.CpuOptions.CoreCount
				}
			}
		}
		prom.C.SetTotalCoreCount(service.Account, service.Region, float64(totalCoreCount))
	}
}
