package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Region string
	Bucket string
}

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client(cfg Config) (*S3Client, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("S3 region is required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	return &S3Client{
		client: s3.NewFromConfig(awsCfg),
		bucket: cfg.Bucket,
	}, nil
}

// StorageClassFor returns the S3 storage class string for a given variant name.
// "original" uses INTELLIGENT_TIERING to auto-tier old photos. Everything else is STANDARD.
func StorageClassFor(variant string) string {
	if variant == "original" {
		return string(types.StorageClassIntelligentTiering)
	}
	return string(types.StorageClassStandard)
}

// Upload uploads the file at localPath to S3 at key, using storageClass.
func (c *S3Client) Upload(ctx context.Context, localPath, key, storageClass string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", localPath, err)
	}
	defer f.Close()

	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(c.bucket),
		Key:          aws.String(key),
		Body:         f,
		StorageClass: types.StorageClass(storageClass),
	})
	if err != nil {
		return fmt.Errorf("S3 put %s: %w", key, err)
	}
	return nil
}
