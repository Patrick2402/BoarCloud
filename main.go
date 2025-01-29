package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fatih/color"
)

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

	cfg, err := awsCfg(region)
	if err != nil {
		log.Fatalf("Cannot load AWS configuration: %v", err)
	}

	if inventory == "true" {
		log.Println(color.CyanString("Performing inventory scan..."))

		// Use appropriate formatter for inventory
		inventoryFormatter := checkOutputInventory(output)
		if inventoryFormatter == nil {
			log.Fatalf("Invalid output format for inventory: %s", output)
		}

		// Perform inventory scan
		performInventory(cfg, inventoryFormatter)
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
			serviceS3(cfg)
		}
	case "lambda":
		{
			log.Println(color.CyanString("Service assessment: Lambda functions"))
			serviceLambda(cfg, output)
		}
	case "sns":
		{
			log.Println(color.CyanString("Service assessment: SNS"))
			serviceSNS(cfg, output)
		}
	case "sqs":
		{
			log.Println(color.CyanString("Service assessment: SQS"))
			serviceSQS(cfg, output)
		}
	default:
		fmt.Print("Unsupported service. Use 's3' or 'lambda'.")
	}
}
