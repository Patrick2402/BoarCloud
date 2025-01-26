package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/fatih/color"
)

// AwsInit holds AWS context and configuration
type AwsInit struct {
	ctx context.Context
	cfg aws.Config
}

// awsInit initializes the AWS configuration and context
func awsInit(region string) AwsInit {

	ctx := context.Background()
	var awsConfig AwsInit
	log.Println(color.BlueString("Loading default AWS account configuration..."))

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("Cannot load AWS configuration: %v", err)
	}
	log.Println(color.GreenString("Configuration loaded successfuly!"))

	awsConfig.cfg = cfg
	awsConfig.ctx = ctx

	return awsConfig
}

func checkOutput(output string) Formatter {
	if output == "table" {
		log.Println(color.CyanString("Output form: table"))
		formatter := &TableFormatter{}
		return formatter
	}
	if output == "json" {
		log.Println(color.CyanString("Output form: JSON"))
		formatter := &JSONFormatter{}
		return formatter
	}
	return nil
}

// need to change bc i do not have idea
func checkOutputInventory(output string) FormatterInventory {
	if output == "table" {
		log.Println(color.CyanString("Output form: table"))
		formatter := &TableFormatterInventory{}
		return formatter
	}
	if output == "json" {
		log.Println(color.CyanString("Output form: JSON"))
		formatter := &JSONFormatterInventory{}
		return formatter
	}
	return nil
}


func main() {
	var service string
	var region string
	var output string
	var inventory string
	banner()

	log.Println(color.BlueString("Analysing arguments..."))
	flag.StringVar(&service, "service", "--help", "Which AWS service should be assessed")
	flag.StringVar(&region, "region", "eu-central-1", "AWS region")
	flag.StringVar(&output, "output", "table", "Data format of the output")
	flag.StringVar(&inventory, "inventory", "false", "Basic inventory scan of resources in AWS account")

	flag.Parse()
	log.Println(color.GreenString("Arguments fine!"))

	awsConfig := awsInit(region)

	if inventory == "true" {
		log.Println(color.CyanString("Performing inventory scan..."))

		// Use appropriate formatter for inventory
		inventoryFormatter := checkOutputInventory(output)
		if inventoryFormatter == nil {
			log.Fatalf("Invalid output format for inventory: %s", output)
		}

		// Perform inventory scan
		performInventory(awsConfig.cfg, awsConfig.ctx, inventoryFormatter)
		return 
	}

	formatter := checkOutput(output)
	if formatter == nil {
		log.Fatalf("Invalid output format: %s", output)
	}

	switch service {
	case "s3":
		{
			log.Println(color.CyanString("Service assessment: S3 buckets"))
			serviceS3(awsConfig.cfg, awsConfig.ctx)
		}
	case "lambda":
		{
			log.Println(color.CyanString("Service assessment: Lambda functions"))
			serviceLambda(awsConfig.cfg, awsConfig.ctx, output)
		}
	case "sns":
		{
			log.Println(color.CyanString("Service assessment: SNS"))
			serviceSNS(awsConfig.cfg, awsConfig.ctx, output)
		}
	case "sqs":
		{
			log.Println(color.CyanString("Service assessment: SQS"))
			serviceSQS(awsConfig.cfg, awsConfig.ctx, output)
		}
	default:
		fmt.Print("Unsupported service. Use 's3' or 'lambda'.")
	}
}
