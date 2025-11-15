package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client wraps AWS S3 operations
type S3Client struct {
	client *s3.Client
}

// NewS3Client creates a new S3 client
func NewS3Client(cfg aws.Config) *S3Client {
	return &S3Client{
		client: s3.NewFromConfig(cfg),
	}
}

// CreateBucket creates a new S3 bucket
func (sc *S3Client) CreateBucket(ctx context.Context, bucketName, region string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	// For regions other than us-east-1, specify location constraint
	if region != "us-east-1" {
		input.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(region),
		}
	}

	_, err := sc.client.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	return nil
}

// PutBucketCORS configures CORS for the bucket
func (sc *S3Client) PutBucketCORS(ctx context.Context, bucketName string) error {
	corsRules := []s3Types.CORSRule{
		{
			AllowedHeaders: []string{"*"},
			AllowedMethods: []string{"GET", "PUT", "POST", "DELETE", "HEAD"},
			AllowedOrigins: []string{"*"},
			ExposeHeaders:  []string{"ETag"},
			MaxAgeSeconds:  aws.Int32(3000),
		},
	}

	_, err := sc.client.PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: aws.String(bucketName),
		CORSConfiguration: &s3Types.CORSConfiguration{
			CORSRules: corsRules,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to configure CORS: %w", err)
	}

	return nil
}

// PutBucketLifecycle configures lifecycle policy for the bucket
func (sc *S3Client) PutBucketLifecycle(ctx context.Context, bucketName string, expirationDays int32) error {
	lifecycleRules := []s3Types.LifecycleRule{
		{
			ID:     aws.String("expire-old-images"),
			Status: s3Types.ExpirationStatusEnabled,
			Expiration: &s3Types.LifecycleExpiration{
				Days: aws.Int32(expirationDays),
			},
			Filter: &s3Types.LifecycleRuleFilter{
				Prefix: aws.String("images/"),
			},
		},
	}

	_, err := sc.client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
		LifecycleConfiguration: &s3Types.BucketLifecycleConfiguration{
			Rules: lifecycleRules,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to configure lifecycle policy: %w", err)
	}

	return nil
}

// BlockPublicAccess blocks all public access to the bucket
func (sc *S3Client) BlockPublicAccess(ctx context.Context, bucketName string) error {
	_, err := sc.client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
		PublicAccessBlockConfiguration: &s3Types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to block public access: %w", err)
	}

	return nil
}

// BucketExists checks if a bucket exists
func (sc *S3Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	_, err := sc.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	return true, nil
}

// DeleteBucket deletes a bucket (must be empty)
func (sc *S3Client) DeleteBucket(ctx context.Context, bucketName string) error {
	_, err := sc.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	return nil
}

// EmptyBucket deletes all objects in a bucket
func (sc *S3Client) EmptyBucket(ctx context.Context, bucketName string) error {
	// List all objects
	paginator := s3.NewListObjectsV2Paginator(sc.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	var objectsToDelete []s3Types.ObjectIdentifier

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			objectsToDelete = append(objectsToDelete, s3Types.ObjectIdentifier{
				Key: obj.Key,
			})
		}
	}

	// Delete objects in batches of 1000
	if len(objectsToDelete) > 0 {
		for i := 0; i < len(objectsToDelete); i += 1000 {
			end := i + 1000
			if end > len(objectsToDelete) {
				end = len(objectsToDelete)
			}

			_, err := sc.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &s3Types.Delete{
					Objects: objectsToDelete[i:end],
					Quiet:   aws.Bool(true),
				},
			})
			if err != nil {
				return fmt.Errorf("failed to delete objects: %w", err)
			}
		}
	}

	return nil
}
