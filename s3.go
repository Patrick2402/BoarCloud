package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func serviceS3(cfg AwsCfg) {
	s3Client := s3.NewFromConfig(cfg.cfg)

	buckets, err := s3Client.ListBuckets(cfg.ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal("Couldn't list buckets. Here's why: : ", err)
	}

	fmt.Println("Bucket List:")
	for key, bucket := range buckets.Buckets {
		fmt.Printf("%d: %s\n", key+1, aws.ToString(bucket.Name))
	}
}
