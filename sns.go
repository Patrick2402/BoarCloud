package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

type SnsTopics struct {
	TopicName              string `json:"topicName"`
	TopicArn               string `json:"topicArn"`
	Encrypted              bool   `json:"encrypted,omitempty"`
	SubscriptionsConfirmed int    `json:"SubscriptionsConfirmed"`
}

type FormatterSNS interface {
	Format(topics []SnsTopics)
}

type JSONFormatterSns struct{}

func (f *JSONFormatterSns) Format(topics []SnsTopics) {
	err := serviceAssessmentToJSONFile("sns", topics)

	if err != nil {
		log.Println(color.RedString("Cannot save assessment to the JSON file. "), err)
	}
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

func serviceSNS(cfg AwsCfg, output string) {
	// Create SNS service client
	snsClient := sns.NewFromConfig(cfg.cfg)
	var topics []SnsTopics

	// Paginate through the results
	paginator := sns.NewListTopicsPaginator(snsClient, &sns.ListTopicsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(cfg.ctx)
		if err != nil {
			log.Fatalf("failed to list topics: %v", err)
		}

		for _, topic := range page.Topics {
			topicName := strings.Split(*topic.TopicArn, ":")
			getTopicAttributesInput := &sns.GetTopicAttributesInput{
				TopicArn: topic.TopicArn,
			}
			topicAttributes, err := snsClient.GetTopicAttributes(cfg.ctx, getTopicAttributesInput)
			if err != nil {
				log.Printf("failed to get topic attributes for %s: %v", *topic.TopicArn, err)
				continue
			}

			encrypted := topicAttributes.Attributes["KmsMasterKeyId"] != ""
			subscriptionsConfirmed, _ := strconv.Atoi(topicAttributes.Attributes["SubscriptionsConfirmed"])

			topics = append(topics, SnsTopics{
				TopicName:              topicName[len(topicName)-1],
				TopicArn:               *topic.TopicArn,
				Encrypted:              encrypted,
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
