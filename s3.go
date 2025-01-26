package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// BucketBasics holds the S3 client for operations
type BucketBasics struct {
	S3Client *s3.Client
}

// ListBuckets lists all buckets in the S3 account
func (basics BucketBasics) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	output, err := basics.S3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Printf("Couldn't list buckets. Here's why: %v\n", err)
		return nil, err
	}
	return output.Buckets, nil
}

// serviceS3 handles the S3 service logic
func serviceS3(cfg aws.Config, ctx context.Context) {
	s3Client := s3.NewFromConfig(cfg)

	bucketBasics := BucketBasics{
		S3Client: s3Client,
	}

	buckets, err := bucketBasics.ListBuckets(ctx)
	if err != nil {
		log.Fatal("Fatal error: ", err)
	}

	fmt.Println("Bucket List:")
	for key, bucket := range buckets {
		fmt.Printf("%d: %s\n", key+1,
			aws.ToString(bucket.Name))
	}
}
