package config

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func NewDynamoDBClient() *dynamodb.Client {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	if endpoint != "" {
		// local DynamoDB (e.g. LocalStack or dynamodb-local)
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(getEnv("AWS_REGION", "us-east-1")),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				getEnv("AWS_ACCESS_KEY_ID", "local"),
				getEnv("AWS_SECRET_ACCESS_KEY", "local"),
				"",
			)),
			config.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
				}),
			),
		)
		if err != nil {
			log.Fatalf("failed to load AWS config: %v", err)
		}
		return dynamodb.NewFromConfig(cfg)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(getEnv("AWS_REGION", "us-east-1")),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	return dynamodb.NewFromConfig(cfg)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
