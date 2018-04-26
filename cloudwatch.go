package main

import (
	_ "fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"sort"
	"time"
)

func createCloudwatchSession() *cloudwatch.CloudWatch {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return cloudwatch.New(sess)
}

func getCloudwatchMetric(resource *awsResource, cloudwatchInfo *cloudwatchInfo, metric metric) float64 {
	c := createCloudwatchSession()

	dimensions := []*cloudwatch.Dimension{
		&cloudwatch.Dimension{
			Name:  cloudwatchInfo.DimensionName,
			Value: resource.Id,
		},
	}

	for _, dim := range cloudwatchInfo.CustomDimension {
		dimension := &cloudwatch.Dimension{
			Name:  &dim.Key,
			Value: &dim.Value,
		}
		dimensions = append(dimensions, dimension)
	}

	period := int64(metric.Length)
	length := metric.Length
	endTime := time.Now()
	startTime := time.Now().Add(-time.Duration(length) * time.Minute)
	statistics := []*string{&metric.Statistics}

	resp, err := c.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		Dimensions: dimensions,
		Namespace:  cloudwatchInfo.Namespace,
		StartTime:  &startTime,
		EndTime:    &endTime,
		Period:     &period,
		MetricName: aws.String(metric.Name),
		Statistics: statistics,
	})

	if err != nil {
		panic(err)
	}

	points := sortDatapoints(resp.Datapoints, metric.Statistics)

	if len(points) == 0 {
		return float64(-1)
	} else {
		return float64(*points[0])
	}

}

func sortDatapoints(datapoints []*cloudwatch.Datapoint, statistic string) (points []*float64) {
	for _, point := range datapoints {
		if statistic == "Sum" {
			points = append(points, point.Sum)
		} else if statistic == "Average" {
			points = append(points, point.Average)
		} else if statistic == "Maximum" {
			points = append(points, point.Maximum)
		} else if statistic == "Minimum" {
			points = append(points, point.Minimum)
		}
	}

	sort.Slice(points, func(i, j int) bool { return *points[i] > *points[j] })

	return points
}
