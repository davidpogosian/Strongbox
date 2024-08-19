package storage

import (
	"bytes"
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	bucketName = "strongbox-bucket"
	region     = "us-east-2"
)

// InitializeStorage loads the AWS configuration and creates an S3 client
func InitializeStorage() *s3.Client {
	// Load the AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an S3 client
	return s3.NewFromConfig(cfg)
}

// UploadFile uploads a file to the S3 bucket
func UploadFile(client *s3.Client, key string, fileBytes []byte) error {
	// Prepare the file upload input
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(fileBytes),
	}

	// Upload the file to S3
	_, err := client.PutObject(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}
