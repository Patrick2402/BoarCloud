package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// ServiceFunction defines a function type for counting resources
type ServiceFunction func(cfg aws.Config, ctx context.Context) int

// InventoryResult stores the results of the inventory scan
type InventoryResult struct {
	Service string `json:"service"`
	Count   int    `json:"count"`
}

// listS3Buckets counts the number of S3 buckets
func listS3Buckets(cfg aws.Config, ctx context.Context) int {
	client := s3.NewFromConfig(cfg)

	output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Printf("Failed to list S3 buckets: %v", err)
		return 0
	}

	return len(output.Buckets)
}

// listLambdaFunctions counts the number of Lambda functions
func listLambdaFunctions(cfg aws.Config, ctx context.Context) int {
	client := lambda.NewFromConfig(cfg)

	var count int
	var marker *string

	for {
		output, err := client.ListFunctions(ctx, &lambda.ListFunctionsInput{
			Marker: marker,
		})
		if err != nil {
			log.Printf("Failed to list Lambda functions: %v", err)
			return count
		}

		count += len(output.Functions)

		if output.NextMarker == nil {
			break
		}
		marker = output.NextMarker
	}

	return count
}

func listSnsTopics(cfg aws.Config, ctx context.Context) int {
	snsClient := sns.NewFromConfig(cfg)

	var topics []types.Topic
	paginator := sns.NewListTopicsPaginator(snsClient, &sns.ListTopicsInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			log.Printf("Couldn't get topics. Here's why: %v\n", err)
			break
		} else {
			topics = append(topics, output.Topics...)
		}
	}

	return len(topics)

}

// performInventory performs an inventory scan and returns results
func performInventory(cfg aws.Config, ctx context.Context, formatter FormatterInventory) {
	// Map of supported services and their corresponding functions
	services := map[string]ServiceFunction{
		"s3":     listS3Buckets,
		"lambda": listLambdaFunctions,
		"sns": listSnsTopics,
	}

	var results []InventoryResult

	// Iterate over all services and count resources
	for serviceName, function := range services {
		log.Printf(color.CyanString("Inventory scanning service: %s", serviceName))
		count := function(cfg, ctx)
		results = append(results, InventoryResult{
			Service: serviceName,
			Count:   count,
		})
	}
	formatter.Format(results)
}

// Formatter interface for output formatting
type FormatterInventory interface {
	Format(results []InventoryResult)
}

// TableFormatter formats results in a table
type TableFormatterInventory struct{}

func (f *TableFormatterInventory) Format(results []InventoryResult) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Service", "Count"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, results := range results {
		table.Append([]string{
			results.Service,
			strconv.Itoa(results.Count),
		})
	}
	table.SetBorder(true)
	// table.SetRowLine(true)
	table.Render()
}

// JSONFormatter formats results as JSON
type JSONFormatterInventory struct{}

func (f *JSONFormatterInventory) Format(results []InventoryResult) {

	file, err := os.Create("inventory.json")
	if err != nil {
		log.Printf("Failed to create inventory.json: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		log.Printf("Failed to write to inventory.json: %v", err)
		return
	}

	log.Println(color.GreenString("Inventory saved to inventory.json"))
}
