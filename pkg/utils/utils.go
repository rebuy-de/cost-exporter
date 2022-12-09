package utils

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

func GetIntervalForPastDay(daysAgo int) *types.DateInterval {
	now := time.Now()
	start := now.AddDate(0, 0, -daysAgo)
	end := now.AddDate(0, 0, (-daysAgo)+1)
	dateRange := types.DateInterval{
		Start: aws.String(start.Format("2006-01-02")),
		End:   aws.String(end.Format("2006-01-02")),
	}
	return &dateRange
}
