package retriever

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rebuy-de/cost-exporter/pkg/config"
	"github.com/rebuy-de/cost-exporter/pkg/prom"
	"github.com/sirupsen/logrus"
)

type APIRetriever struct {
	Accounts    []config.Account
	IntervalSec int64
	Services    []Service
}

type Service struct {
	Region  string
	svc     *ec2.EC2
	Account string
}

func (c *APIRetriever) Run() {
	c.initialize()
	c.getCores()
	c.getSpotInstances()
	go c.scheduleInterval()
}

func (c *APIRetriever) initialize() {
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

func (c *APIRetriever) scheduleInterval() {
	for range time.Tick(time.Duration(c.IntervalSec) * time.Second) {
		c.getCores()
		c.getSpotInstances()
	}
}

func (c *APIRetriever) getCores() {
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
					totalCoreCount += *instance.CpuOptions.CoreCount * *instance.CpuOptions.ThreadsPerCore
				}
			}
		}
		prom.C.SetTotalCoreCount(service.Account, service.Region, float64(totalCoreCount))
	}
}

func (c *APIRetriever) getSpotInstances() {
	spotRequestItems := []prometheus.Labels{}
	for _, service := range c.Services {
		logrus.Infof("Getting SpotInstances for account '%s' and region '%s'", service.Account, service.Region)
		resp, err := service.svc.DescribeSpotInstanceRequests(&ec2.DescribeSpotInstanceRequestsInput{})
		if err != nil {
			fmt.Println(err)
		}

		for _, request := range resp.SpotInstanceRequests {
			var instanceID string
			if request.InstanceId == nil {
				instanceID = ""
			} else {
				instanceID = *request.InstanceId
			}

			var launchedAvailabilityZone string
			if request.LaunchedAvailabilityZone == nil {
				launchedAvailabilityZone = ""
			} else {
				launchedAvailabilityZone = *request.LaunchedAvailabilityZone
			}

			labels := prometheus.Labels{
				"account":           service.Account,
				"region":            service.Region,
				"state":             *request.State,
				"code":              *request.Status.Code,
				"instance_type":     *request.LaunchSpecification.InstanceType,
				"instance_id":       instanceID,
				"availability_zone": launchedAvailabilityZone,
			}
			spotRequestItems = append(spotRequestItems, labels)
		}
	}
	prom.C.SetSpotRequest(spotRequestItems)
}
