package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

func serviceSQS(cfg AwsCfg, output string) {
	log.Println(color.RedString("SQS assessment!"))

	sqsClient := sqs.NewFromConfig(cfg.cfg)
	var queues []SqsQueues

	// Paginate through the results
	paginator := sqs.NewListQueuesPaginator(sqsClient, &sqs.ListQueuesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(cfg.ctx)
		if err != nil {
			log.Printf("Couldn't get queues. Here's why: %v\n", err)
			break
		}
		for _, queue := range page.QueueUrls {
			getQueueAttributesInput := sqs.GetQueueAttributesInput{
				QueueUrl: &queue,
				AttributeNames: []types.QueueAttributeName{
					types.QueueAttributeNameQueueArn,
					types.QueueAttributeNameKmsMasterKeyId,
				},
			}
			attributesOutput, err := sqsClient.GetQueueAttributes(cfg.ctx, &getQueueAttributesInput)

			isEncrypted := false
			queueArn := ""
			if err == nil {
				_, isEncrypted = attributesOutput.Attributes[string(types.QueueAttributeNameKmsMasterKeyId)]
				queueArn = attributesOutput.Attributes[string(types.QueueAttributeNameQueueArn)]
			}
			sqsName := strings.Split(queue, "/")

			queues = append(queues, SqsQueues{
				QueueName: sqsName[len(sqsName)-1],
				QueueArn:  queueArn,
				Encrypted: isEncrypted,
			})
		}
	}

	if output == "table" {
		formatter := &TableFormatterSqs{}
		formatter.Format(queues)
	} else {
		formatter := &JSONFormatterSqs{}
		formatter.Format(queues)
	}
}

type SqsQueues struct {
	QueueName string `json:"queueName"`
	QueueArn  string `json:"queueArn"`
	Encrypted bool   `json:"encrypted,omitempty"`
}

type FormatterSQS interface {
	Format(queues []SqsQueues)
}

type TableFormatterSqs struct{}
type JSONFormatterSqs struct{}

func (f *TableFormatterSqs) Format(queues []SqsQueues) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Queue", "Arn", "Encryprion"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, queue := range queues {
		table.Append([]string{
			queue.QueueName,
			queue.QueueArn,
			fmt.Sprintf("%t", queue.Encrypted),
		})
	}
	table.SetBorder(true)
	table.Render()
}

func (f *JSONFormatterSqs) Format(queues []SqsQueues) {
	err := serviceAssessmentToJSONFile("sqs", queues)

	if err != nil {
		log.Println(color.RedString("Cannot save assessment to the JSON file. "), err)
	}
}
