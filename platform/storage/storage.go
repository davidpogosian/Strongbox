package storage

import (
	"io"
	"bytes"
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithy "github.com/aws/smithy-go"
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

// DeleteFile deletes a file from the S3 bucket
func DeleteFile(client *s3.Client, key string) error {
	// Prepare the file deletion input
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	// Delete the file from S3
	_, err := client.DeleteObject(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}

// GetFile retrieves a file from the S3 bucket
func GetFile(client *s3.Client, key string) ([]byte, error) {
	// Prepare the file retrieval input
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	// Retrieve the file from S3
	resp, err := client.GetObject(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the file content into a byte slice
	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func ObjectExists(s3Client *s3.Client, key string) (bool, error) {
    // Create the HeadObjectInput
    headObjectInput := &s3.HeadObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(key),
    }

    // Call HeadObject to check if the object exists
    _, err := s3Client.HeadObject(context.TODO(), headObjectInput)
    if err != nil {
        var apiErr smithy.APIError
        if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
            return false, nil // Object does not exist
        }
        return false, err // Some other error occurred
    }

    return true, nil // Object exists
}
