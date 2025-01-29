package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	// "github.com/fatih/color"
)

type SnsTopics struct {
	TopicName              string `json:"topicName"`
	TopicArn               string `json:"topicArn"`
	Encrypted              bool   `json:"encrypted,omitempty"`
	SubscriptionsConfirmed int    `json:"SubscriptionsConfirmed"`
}

func serviceSNS(cfg AwsCfg, output string) {
	snsClient := sns.NewFromConfig(cfg.cfg)
	var topics []SnsTopics

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
		FormatTable(topics, []string{"Topic", "Arn", "Encrypted", "Subscriptions"})
	} else {
		FormatJSON(topics, "sns")
	}
}
