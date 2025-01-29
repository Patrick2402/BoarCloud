package main

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fatih/color"
)

type SqsQueues struct {
	QueueName string `json:"queueName"`
	QueueArn  string `json:"queueArn"`
	Encrypted bool   `json:"encrypted,omitempty"`
}

func serviceSQS(cfg AwsCfg, output string) {
	log.Println(color.RedString("SQS assessment!"))

	sqsClient := sqs.NewFromConfig(cfg.cfg)
	var queues []SqsQueues

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
		FormatTable(queues, []string{"Queue", "Arn", "Encrypted"})
	} else {
		FormatJSON(queues, "sqs")
	}
}

