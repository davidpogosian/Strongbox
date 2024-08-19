package main

import (
	"context"
	"log"

	/*"fmt"*/
	"net/http"
	"time"

	"strongbox/platform/authenticator"
	"strongbox/platform/database"
	"strongbox/platform/router"

	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go-v2/aws"
	/*"github.com/aws/aws-sdk-go-v2/config"*/
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	bucketName = "strongbox-bucket"
	region     = "us-east-2"
)

func generatePresignedURL(client *s3.PresignClient, key string, expiration time.Duration, isUpload bool) (string, error) {
	if isUpload {
		// Generate pre-signed URL for uploading a file
		req, err := client.PresignPutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(expiration))

		if err != nil {
			return "", err
		}

		return req.URL, nil
	} else {
		// Generate pre-signed URL for downloading a file
		req, err := client.PresignGetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(expiration))

		if err != nil {
			return "", err
		}

		return req.URL, nil
	}
}

func main() {
	/*
	// Load the AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Create a Presign client
	presignClient := s3.NewPresignClient(s3Client)

	// Generate a presigned URL for uploading a file
	uploadKey := "parcel.txt"
	uploadURL, err := generatePresignedURL(presignClient, uploadKey, 15*time.Minute, true)
	if err != nil {
		log.Fatalf("failed to generate presigned URL: %v", err)
	}
	fmt.Printf("Presigned URL for upload: %s\n", uploadURL)

	// Generate a presigned URL for downloading a file
	downloadKey := "parcel.txt"
	downloadURL, err := generatePresignedURL(presignClient, downloadKey, 15*time.Minute, false)
	if err != nil {
		log.Fatalf("failed to generate presigned URL: %v", err)
	}
	fmt.Printf("Presigned URL for download: %s\n", downloadURL)
	 */



	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	database.DeleteDatabase("strongbox.db")
	db := database.InitializeDatabaseConnection("strongbox.db")
	defer database.TeardownDatabaseConnection(db)
	database.CreateAssetTable(db)

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	rtr := router.New(db, auth)

	log.Print("Server listening on http://localhost:3000/")
	if err := http.ListenAndServe("0.0.0.0:3000", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
