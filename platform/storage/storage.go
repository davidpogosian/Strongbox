package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"
)

const (
	bucketName = "strongbox-bucket"
	region     = "us-east-2" // Only used in init func?
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

func ListObjects(s3Client *s3.Client, prefix string) ([]types.Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	}

	result, err := s3Client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return result.Contents, nil
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
func DestroyObject(client *s3.Client, key string) error {
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

func GetObject(s3Client *s3.Client, key string) (io.ReadCloser, error) {
	// Define the GetObject input parameters
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName), // Replace with your bucket name
		Key:    aws.String(key),
	}

	// Retrieve the object from S3
	result, err := s3Client.GetObject(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	// Return the object's body (data) and no error
	return result.Body, nil
}

func GeneratePresignedURL(s3Client *s3.Client, key string, expiration time.Duration) (string, error) {
	// Create a presign client
	psClient := s3.NewPresignClient(s3Client)

	// Define the input parameters for the presigned URL request
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	// Generate the presigned URL
	req, err := psClient.PresignGetObject(context.TODO(), input, func(po *s3.PresignOptions) {
		po.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}
