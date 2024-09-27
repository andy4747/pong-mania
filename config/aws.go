package config

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

func NewAWSConfig() *aws.Config {
	goenv := os.Getenv("GOENV")
	if goenv == "production" {
		region := os.Getenv("AWS_REGION")
		if region == "" {
			panic("AWS credentials or region not set in environment variables")
		}
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			panic(err)
		}
		return &cfg
	} else {
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		region := os.Getenv("AWS_REGION")

		if accessKey == "" || secretKey == "" || region == "" {
			panic("AWS credentials or region not set in environment variables")
		}
		// Create the config, explicitly using the credentials from environment variables
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			panic(err)
		}
		return &cfg
	}
}

func NewS3Client(cfg *aws.Config) *s3.Client {
	client := s3.NewFromConfig(*cfg)
	return client
}

func NewSESConfig(cfg *aws.Config) *ses.Client {
	client := ses.NewFromConfig(*cfg)
	return client
}

func NewSecretManager(cfg *aws.Config) *secretsmanager.Client {
	client := secretsmanager.NewFromConfig(*cfg)
	return client
}
