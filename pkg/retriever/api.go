package retriever

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	ceConfig "github.com/rebuy-de/cost-exporter/pkg/config"
	"github.com/rebuy-de/cost-exporter/pkg/prom"
)

type APIRetriever struct {
	Accounts    []ceConfig.Account
	IntervalSec int64
	Services    []Service
}

type Service struct {
	Region  string
	svc     *ec2.Client
	Account string
}

func (c *APIRetriever) Run(ctx context.Context) {
	c.initialize(ctx)
	c.getCores(ctx)
	c.getSpotInstances(ctx)
	go c.scheduleInterval(ctx)
}

func getAvailableRegions(ctx context.Context, credProvider credentials.StaticCredentialsProvider) []types.Region {
	conf, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credProvider),
		config.WithRegion("eu-west-1"),
	)
	if err != nil {
		fmt.Println(err)
	}

	svc := ec2.NewFromConfig(conf)
	regions, err := svc.DescribeRegions(ctx, &ec2.DescribeRegionsInput{AllRegions: aws.Bool(false)})
	if err != nil {
		fmt.Println(err)
	}
	return regions.Regions
}

func (c *APIRetriever) initialize(ctx context.Context) {
	for _, account := range c.Accounts {
		credProvider := credentials.NewStaticCredentialsProvider(account.ID, account.Secret, "")
		regions := getAvailableRegions(ctx, credProvider)

		for _, region := range regions {
			conf, err := config.LoadDefaultConfig(ctx,
				config.WithCredentialsProvider(credProvider),
				config.WithRegion(*region.RegionName),
			)
			if err != nil {
				fmt.Println(err)
			}

			svc := ec2.NewFromConfig(conf)

			c.Services = append(c.Services, Service{
				Region:  *region.RegionName,
				svc:     svc,
				Account: account.Name,
			})
		}
	}
}

func (c *APIRetriever) scheduleInterval(ctx context.Context) {
	for range time.Tick(time.Duration(c.IntervalSec) * time.Second) {
		c.getCores(ctx)
		c.getSpotInstances(ctx)
	}
}

func (c *APIRetriever) getCores(ctx context.Context) {
	for _, service := range c.Services {
		logrus.Infof("Getting cores for account '%s' and region '%s'", service.Account, service.Region)
		var totalCoreCount int32
		params := &ec2.DescribeInstancesInput{
			MaxResults: aws.Int32(100),
		}
		for {
			resp2, err := service.svc.DescribeInstances(ctx, params)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, reservation := range resp2.Reservations {
				for _, instance := range reservation.Instances {
					if instance.State.Name == types.InstanceStateNameRunning {
						totalCoreCount += *instance.CpuOptions.CoreCount * *instance.CpuOptions.ThreadsPerCore
					}
				}
			}

			if resp2.NextToken == nil {
				break
			}

			params.NextToken = resp2.NextToken
		}
		prom.C.SetTotalCoreCount(service.Account, service.Region, float64(totalCoreCount))
	}
}

func (c *APIRetriever) getSpotInstances(ctx context.Context) {
	spotRequestItems := []prometheus.Labels{}
	for _, service := range c.Services {
		logrus.Infof("Getting SpotInstances for account '%s' and region '%s'", service.Account, service.Region)
		resp, err := service.svc.DescribeSpotInstanceRequests(ctx, &ec2.DescribeSpotInstanceRequestsInput{})
		if err != nil {
			fmt.Println(err)
			continue
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
				"state":             string(request.State),
				"code":              *request.Status.Code,
				"instance_type":     string(request.LaunchSpecification.InstanceType),
				"instance_id":       instanceID,
				"availability_zone": launchedAvailabilityZone,
			}
			spotRequestItems = append(spotRequestItems, labels)
		}
	}
	prom.C.SetSpotRequest(spotRequestItems)
}
