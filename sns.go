package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/fatih/color"

	// "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/olekukonko/tablewriter"
)

type SnsTopics struct {
	TopicName string `json:"topicName"`
	TopicArn  string `json:"topicArn"`
	Encrypted bool   `json:"encrypted,omitempty"`
	SubscriptionsConfirmed int `json:"SubscriptionsConfirmed"`
}

type FormatterSNS interface {
	Format(topics []SnsTopics)
}

type JSONFormatterSns struct {}


func (f *JSONFormatterSns) Format(topics []SnsTopics) {
	file, err := os.Create("sns.json")
	if err != nil {
		log.Printf("Failed to create sns.json: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(topics); err != nil {
		log.Printf("Failed to write to sns.json: %v", err)
		return
	}

	log.Println(color.GreenString("Inventory saved to inventory.json"))
}


type TableFormatterSns struct{}

func (f *TableFormatterSns) Format(topics []SnsTopics) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Topic", "Arn", "Encrypted", "Subscriptions"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, topic := range topics {
		table.Append([]string{
			topic.TopicName,
			topic.TopicArn,
			fmt.Sprintf("%t", topic.Encrypted),
			fmt.Sprintf("%d", topic.SubscriptionsConfirmed),
		})
	}
	table.SetBorder(true)
	table.Render()
}

func serviceSNS(cfg aws.Config, ctx context.Context, output string) {
	// Create SNS service client
	snsClient := sns.NewFromConfig(cfg)

	input := &sns.ListTopicsInput{}
	var topics []SnsTopics

	// Paginate through the results
	paginator := sns.NewListTopicsPaginator(snsClient, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to list topics: %v", err)
		}

		for _, topic := range page.Topics {
			topicName := strings.Split(*topic.TopicArn, ":")
			getTopicAttributesInput := &sns.GetTopicAttributesInput{
				TopicArn: topic.TopicArn,
			}
			topicAttributes, err := snsClient.GetTopicAttributes(ctx, getTopicAttributesInput)
			if err != nil {
				log.Printf("failed to get topic attributes for %s: %v", *topic.TopicArn, err)
				continue
			}

			encrypted := topicAttributes.Attributes["KmsMasterKeyId"] != ""
			subscriptionsConfirmed, _ := strconv.Atoi(topicAttributes.Attributes["SubscriptionsConfirmed"])
		
			topics = append(topics, SnsTopics{
				TopicName: topicName[len(topicName)-1],
				TopicArn:  *topic.TopicArn,
				Encrypted: encrypted,
				SubscriptionsConfirmed: subscriptionsConfirmed,
			})
		}
	}

	if output == "table" {
		formatter := &TableFormatterSns{}
		formatter.Format(topics)
	} else {
		formatter := &JSONFormatterSns{}
		formatter.Format(topics)
	}
}
