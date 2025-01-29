package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

type ServiceFunction func(AwsCfg) int

type InventoryResult struct {
	Service string `json:"service"`
	Count   int    `json:"count"`
}

type AwsCfg struct {
	cfg aws.Config
	ctx context.Context
}

// awsInit initializes the AWS configuration and context
func awsCfg(region string) (init AwsCfg, err error) {
	ctx := context.Background()
	log.Println(color.BlueString("Loading default AWS account configuration..."))

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return AwsCfg{}, err
	}

	log.Println(color.GreenString("Configuration loaded successfuly!"))

	return AwsCfg{ctx: ctx, cfg: cfg}, nil
}

func listS3Buckets(cfg AwsCfg) int {
	client := s3.NewFromConfig(cfg.cfg)

	output, err := client.ListBuckets(cfg.ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Printf("Failed to list S3 buckets: %v", err)
		return 0
	}

	return len(output.Buckets)
}

func listLambdaFunctions(cfg AwsCfg) int {
	client := lambda.NewFromConfig(cfg.cfg)

	var count int
	var marker *string

	for {
		output, err := client.ListFunctions(cfg.ctx, &lambda.ListFunctionsInput{
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

func listSnsTopics(cfg AwsCfg) int {
	snsClient := sns.NewFromConfig(cfg.cfg)

	var topics []types.Topic
	paginator := sns.NewListTopicsPaginator(snsClient, &sns.ListTopicsInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(cfg.ctx)
		if err != nil {
			log.Printf("Couldn't get topics. Here's why: %v\n", err)
			break
		} else {
			topics = append(topics, output.Topics...)
		}
	}

	return len(topics)

}

func listSqsQueues(cfg AwsCfg) int {
	sqsClient := sqs.NewFromConfig(cfg.cfg)

	// Paginate through the results
	paginator := sqs.NewListQueuesPaginator(sqsClient, &sqs.ListQueuesInput{})
	var queues []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(cfg.ctx)
		if err != nil {
			log.Printf("Couldn't get queues. Here's why: %v\n", err)
			break
		} else {
			queues = append(queues, page.QueueUrls...)
		}
	}
	return len(queues)

}

func performInventory(cfg AwsCfg, formatter func([]InventoryResult)) {
	var results []InventoryResult
	makeInventory := func(serviceName string, f ServiceFunction) {
		log.Printf(color.CyanString("Inventory scanning service: %s"), serviceName)
		results = append(results, InventoryResult{
			Service: serviceName,
			Count:   f(cfg),
		})
	}

	makeInventory("s3", listS3Buckets)
	makeInventory("lambda", listLambdaFunctions)
	makeInventory("sns", listSnsTopics)
	makeInventory("sqs", listSqsQueues)

	formatter(results)
}

type FormatterInventory interface {
	Format(results []InventoryResult)
}

type TableFormatterInventory struct{}

func formatTable(results []InventoryResult) {
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

func formatJson(results []InventoryResult) {
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

func checkOutputInventory(output string) func([]InventoryResult) {
	if output == "table" {
		log.Println(color.CyanString("Output form: table"))
		return formatTable
	}
	if output == "json" {
		log.Println(color.CyanString("Output form: JSON"))
		return formatJson
	}
	log.Println(color.CyanString("Default output: table"))
	return formatTable
}
