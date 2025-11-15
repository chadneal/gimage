package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	logTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// CloudWatchClient wraps AWS CloudWatch operations
type CloudWatchClient struct {
	client     *cloudwatch.Client
	logsClient *cloudwatchlogs.Client
}

// NewCloudWatchClient creates a new CloudWatch client
func NewCloudWatchClient(cfg aws.Config) *CloudWatchClient {
	return &CloudWatchClient{
		client:     cloudwatch.NewFromConfig(cfg),
		logsClient: cloudwatchlogs.NewFromConfig(cfg),
	}
}

// GetLambdaMetrics retrieves metrics for a Lambda function
func (cwc *CloudWatchClient) GetLambdaMetrics(ctx context.Context, functionName string, startTime, endTime time.Time) (map[string]float64, error) {
	metrics := make(map[string]float64)

	// Get Invocations
	invocations, err := cwc.getMetricStatistics(ctx, "AWS/Lambda", "Invocations", functionName, "Sum", startTime, endTime)
	if err != nil {
		return nil, err
	}
	metrics["invocations"] = invocations

	// Get Errors
	errors, err := cwc.getMetricStatistics(ctx, "AWS/Lambda", "Errors", functionName, "Sum", startTime, endTime)
	if err != nil {
		return nil, err
	}
	metrics["errors"] = errors

	// Get Throttles
	throttles, err := cwc.getMetricStatistics(ctx, "AWS/Lambda", "Throttles", functionName, "Sum", startTime, endTime)
	if err != nil {
		return nil, err
	}
	metrics["throttles"] = throttles

	// Get Duration (average)
	duration, err := cwc.getMetricStatistics(ctx, "AWS/Lambda", "Duration", functionName, "Average", startTime, endTime)
	if err != nil {
		return nil, err
	}
	metrics["avg_duration"] = duration

	// Get ConcurrentExecutions
	concurrent, err := cwc.getMetricStatistics(ctx, "AWS/Lambda", "ConcurrentExecutions", functionName, "Maximum", startTime, endTime)
	if err != nil {
		return nil, err
	}
	metrics["concurrent_executions"] = concurrent

	return metrics, nil
}

// getMetricStatistics is a helper to get metric statistics
func (cwc *CloudWatchClient) getMetricStatistics(ctx context.Context, namespace, metricName, functionName, statistic string, startTime, endTime time.Time) (float64, error) {
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		Dimensions: []cwTypes.Dimension{
			{
				Name:  aws.String("FunctionName"),
				Value: aws.String(functionName),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int32(300), // 5 minutes
		Statistics: []cwTypes.Statistic{cwTypes.Statistic(statistic)},
	}

	result, err := cwc.client.GetMetricStatistics(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to get metric %s: %w", metricName, err)
	}

	if len(result.Datapoints) == 0 {
		return 0, nil
	}

	// Return the latest datapoint
	var value float64
	for _, dp := range result.Datapoints {
		switch statistic {
		case "Sum":
			if dp.Sum != nil {
				value += *dp.Sum
			}
		case "Average":
			if dp.Average != nil {
				value = *dp.Average
			}
		case "Maximum":
			if dp.Maximum != nil && *dp.Maximum > value {
				value = *dp.Maximum
			}
		}
	}

	return value, nil
}

// GetLogEvents retrieves log events from a log group
func (cwc *CloudWatchClient) GetLogEvents(ctx context.Context, logGroupName string, limit int32) ([]LogEvent, error) {
	// Get log streams
	streamsOutput, err := cwc.logsClient.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroupName),
		OrderBy:      logTypes.OrderByLastEventTime,
		Descending:   aws.Bool(true),
		Limit:        aws.Int32(5), // Get last 5 streams
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe log streams: %w", err)
	}

	var allEvents []LogEvent

	// Get events from each stream
	for _, stream := range streamsOutput.LogStreams {
		eventsOutput, err := cwc.logsClient.GetLogEvents(ctx, &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  aws.String(logGroupName),
			LogStreamName: stream.LogStreamName,
			Limit:         aws.Int32(limit),
			StartFromHead: aws.Bool(false), // Get most recent
		})
		if err != nil {
			continue // Skip streams with errors
		}

		for _, event := range eventsOutput.Events {
			if event.Message != nil && event.Timestamp != nil {
				allEvents = append(allEvents, LogEvent{
					Timestamp: time.UnixMilli(*event.Timestamp),
					Message:   *event.Message,
					Stream:    *stream.LogStreamName,
				})
			}
		}
	}

	return allEvents, nil
}

// FilterLogEvents filters log events by pattern
func (cwc *CloudWatchClient) FilterLogEvents(ctx context.Context, logGroupName, filterPattern string, startTime, endTime time.Time, limit int32) ([]LogEvent, error) {
	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		FilterPattern: aws.String(filterPattern),
		StartTime:     aws.Int64(startTime.UnixMilli()),
		EndTime:       aws.Int64(endTime.UnixMilli()),
		Limit:         aws.Int32(limit),
	}

	result, err := cwc.logsClient.FilterLogEvents(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to filter log events: %w", err)
	}

	var events []LogEvent
	for _, event := range result.Events {
		if event.Message != nil && event.Timestamp != nil {
			streamName := ""
			if event.LogStreamName != nil {
				streamName = *event.LogStreamName
			}
			events = append(events, LogEvent{
				Timestamp: time.UnixMilli(*event.Timestamp),
				Message:   *event.Message,
				Stream:    streamName,
			})
		}
	}

	return events, nil
}

// LogEvent represents a CloudWatch log event
type LogEvent struct {
	Timestamp time.Time
	Message   string
	Stream    string
}
