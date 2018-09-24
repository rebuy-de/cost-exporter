package utils

import (
	"time"

	"github.com/aws/aws-sdk-go/service/costexplorer"
)

func GetIntervalForPastDay(daysAgo int) *costexplorer.DateInterval {
	now := time.Now()
	start := now.AddDate(0, 0, -daysAgo)
	end := now.AddDate(0, 0, (-daysAgo)+1)
	dateRange := costexplorer.DateInterval{}
	dateRange.SetStart(start.Format("2006-01-02"))
	dateRange.SetEnd(end.Format("2006-01-02"))
	return &dateRange
}
