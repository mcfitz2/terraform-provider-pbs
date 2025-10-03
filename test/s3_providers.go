package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/require"
)

// S3ProviderConfig holds configuration for testing with different S3 providers
type S3ProviderConfig struct {
	Name           string
	Endpoint       string
	Region         string
	AccessKey      string
	SecretKey      string
	BucketName     string
	EndpointID     string   // Consistent ID for PBS endpoint
	ProviderQuirks []string // Provider-specific quirks for compatibility
	S3Client       *s3.S3
	ForceDelete    bool // Some providers require force delete for non-empty buckets
}

// GetS3ProviderConfigs returns all configured S3 providers for testing
func GetS3ProviderConfigs(t *testing.T) []*S3ProviderConfig {
	var providers []*S3ProviderConfig

	// AWS S3
	if awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID"); awsAccessKey != "" {
		awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			awsRegion = "us-west-2" // default region
		}

		if awsSecretKey != "" {
			timestamp := time.Now().Unix()
			providers = append(providers, &S3ProviderConfig{
				Name:       "AWS",
				Endpoint:   fmt.Sprintf("s3.%s.amazonaws.com", awsRegion),
				Region:     awsRegion,
				AccessKey:  awsAccessKey,
				SecretKey:  awsSecretKey,
				BucketName: fmt.Sprintf("pbs-test-aws-%d", timestamp),
				EndpointID: fmt.Sprintf("pbs-test-aws-%d", timestamp),
			})
		}
	}

	// Backblaze B2
	if b2AccessKey := os.Getenv("B2_ACCESS_KEY_ID"); b2AccessKey != "" {
		b2SecretKey := os.Getenv("B2_SECRET_ACCESS_KEY")
		b2Region := os.Getenv("B2_REGION")
		if b2Region == "" {
			b2Region = "us-west-004" // default Backblaze region
		}

		if b2SecretKey != "" {
			timestamp := time.Now().Unix()
			// Backblaze B2 S3 API Compatibility Configuration
			// ================================================
			// B2's S3 API has limited compatibility with standard S3 implementations.
			// PBS supports the "skip-if-none-match-header" provider quirk which
			// prevents PBS from sending the If-None-Match header during chunk uploads.
			//
			// Required configuration:
			// - Path-style addressing: REQUIRED (virtual-hosted style not supported)
			// - Provider quirk: "skip-if-none-match-header" (prevents 501 errors)
			// - Endpoint format: s3.{region}.backblazeb2.com
			//
			// With this configuration, B2 works reliably with PBS for:
			// ✓ S3 endpoint creation and management
			// ✓ Datastore creation and verification
			// ✓ Chunk upload operations
			//
			// Reference: docs/BACKBLAZE_B2_REMEDIATION.md
			providers = append(providers, &S3ProviderConfig{
				Name:       "Backblaze",
				Endpoint:   fmt.Sprintf("s3.%s.backblazeb2.com", b2Region),
				Region:     b2Region,
				AccessKey:  b2AccessKey,
				SecretKey:  b2SecretKey,
				BucketName: fmt.Sprintf("pbs-test-b2-%d", timestamp),
				EndpointID: fmt.Sprintf("pbs-test-b2-%d", timestamp),
				ProviderQuirks: []string{
					"skip-if-none-match-header", // Required for B2 S3 API compatibility
				},
			})
		}
	}

	// Scaleway Object Storage
	if scwAccessKey := os.Getenv("SCALEWAY_ACCESS_KEY"); scwAccessKey != "" {
		scwSecretKey := os.Getenv("SCALEWAY_SECRET_KEY")
		scwRegion := os.Getenv("SCALEWAY_REGION")
		if scwRegion == "" {
			scwRegion = "fr-par" // default Scaleway region
		}

		if scwSecretKey != "" {
			timestamp := time.Now().Unix()
			providers = append(providers, &S3ProviderConfig{
				Name:       "Scaleway",
				Endpoint:   fmt.Sprintf("s3.%s.scw.cloud", scwRegion),
				Region:     scwRegion,
				AccessKey:  scwAccessKey,
				SecretKey:  scwSecretKey,
				BucketName: fmt.Sprintf("pbs-test-scw-%d", timestamp),
				EndpointID: fmt.Sprintf("pbs-test-scw-%d", timestamp),
			})
		}
	}

	return providers
}

// SetupS3Client configures the S3 client for the provider
func (p *S3ProviderConfig) SetupS3Client(t *testing.T) {
	creds := credentials.NewStaticCredentials(p.AccessKey, p.SecretKey, "")

	config := &aws.Config{
		Region:      aws.String(p.Region),
		Credentials: creds,
	}

	// Provider-specific configurations
	switch p.Name {
	case "AWS":
		// Use default AWS S3 endpoint (virtual-hosted-style)
		// AWS automatically handles endpoint selection
	case "Backblaze":
		// Backblaze requires custom endpoint and path-style addressing
		config.Endpoint = aws.String(fmt.Sprintf("https://%s", p.Endpoint))
		config.S3ForcePathStyle = aws.Bool(true)
	case "Scaleway":
		// Scaleway requires custom endpoint
		config.Endpoint = aws.String(fmt.Sprintf("https://%s", p.Endpoint))
		config.S3ForcePathStyle = aws.Bool(false) // Virtual-hosted style
	}

	sess, err := session.NewSession(config)
	require.NoError(t, err, "Failed to create AWS session for %s", p.Name)

	p.S3Client = s3.New(sess)
}

// CreateTestBucket creates a test bucket for the provider
func (p *S3ProviderConfig) CreateTestBucket(t *testing.T) {
	t.Logf("Creating test bucket %s on %s", p.BucketName, p.Name)

	// Create bucket with appropriate location constraint based on provider
	var locationConstraint *string
	switch p.Name {
	case "AWS":
		// For AWS, only set location constraint if not us-east-1
		if p.Region != "us-east-1" {
			locationConstraint = aws.String(p.Region)
		}
	case "Backblaze", "Scaleway":
		// For other providers, always set the region
		locationConstraint = aws.String(p.Region)
	}

	var createInput *s3.CreateBucketInput
	if locationConstraint != nil {
		createInput = &s3.CreateBucketInput{
			Bucket: aws.String(p.BucketName),
			CreateBucketConfiguration: &s3.CreateBucketConfiguration{
				LocationConstraint: locationConstraint,
			},
		}
	} else {
		createInput = &s3.CreateBucketInput{
			Bucket: aws.String(p.BucketName),
		}
	}

	_, err := p.S3Client.CreateBucket(createInput)
	require.NoError(t, err, "Failed to create bucket %s on %s", p.BucketName, p.Name)

	// Wait for bucket to be available
	t.Logf("Waiting for bucket %s to be available on %s", p.BucketName, p.Name)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = p.S3Client.WaitUntilBucketExistsWithContext(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(p.BucketName),
	})
	require.NoError(t, err, "Bucket %s not available on %s after 30 seconds", p.BucketName, p.Name)

	t.Logf("Successfully created bucket %s on %s", p.BucketName, p.Name)
}

// DeleteTestBucket deletes the test bucket
func (p *S3ProviderConfig) DeleteTestBucket(t *testing.T) {
	if p.S3Client == nil {
		return
	}

	t.Logf("Deleting test bucket %s on %s", p.BucketName, p.Name)

	// For Backblaze B2, we need to handle versioned objects
	// B2 keeps all versions of objects by default, so we need to delete all versions
	if p.Name == "Backblaze" {
		// List and delete all object versions (B2 specific)
		err := p.S3Client.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
			Bucket: aws.String(p.BucketName),
		}, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
			var objects []*s3.ObjectIdentifier

			// Delete all versions
			for _, version := range page.Versions {
				objects = append(objects, &s3.ObjectIdentifier{
					Key:       version.Key,
					VersionId: version.VersionId,
				})
			}

			// Delete all delete markers
			for _, marker := range page.DeleteMarkers {
				objects = append(objects, &s3.ObjectIdentifier{
					Key:       marker.Key,
					VersionId: marker.VersionId,
				})
			}

			if len(objects) > 0 {
				t.Logf("Deleting %d object versions/markers from B2 bucket %s", len(objects), p.BucketName)
				_, deleteErr := p.S3Client.DeleteObjects(&s3.DeleteObjectsInput{
					Bucket: aws.String(p.BucketName),
					Delete: &s3.Delete{Objects: objects, Quiet: aws.Bool(true)},
				})
				if deleteErr != nil {
					t.Logf("Warning: Failed to delete some object versions: %v", deleteErr)
				}
			}

			return true // Continue to next page
		})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucket" {
				// Bucket doesn't exist, nothing to delete
				return
			}
			t.Logf("Failed to list object versions in B2 bucket %s: %v", p.BucketName, err)
		}

		// Wait for deletions to propagate in B2
		time.Sleep(3 * time.Second)
	} else {
		// Standard S3 object deletion for non-B2 providers
		listOutput, err := p.S3Client.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(p.BucketName),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucket" {
				// Bucket doesn't exist, nothing to delete
				return
			}
			t.Logf("Failed to list objects in bucket %s on %s: %v", p.BucketName, p.Name, err)
			return
		}

		if len(listOutput.Contents) > 0 {
			// Delete all objects
			var objects []*s3.ObjectIdentifier
			for _, obj := range listOutput.Contents {
				objects = append(objects, &s3.ObjectIdentifier{
					Key: obj.Key,
				})
			}

			_, err = p.S3Client.DeleteObjects(&s3.DeleteObjectsInput{
				Bucket: aws.String(p.BucketName),
				Delete: &s3.Delete{
					Objects: objects,
				},
			})
			if err != nil {
				t.Logf("Failed to delete objects from bucket %s on %s: %v", p.BucketName, p.Name, err)
				return
			}

			// Wait for object deletions to propagate
			time.Sleep(2 * time.Second)
		}
	}

	// Try to delete the bucket with retries
	maxRetries := 5
	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, deleteErr := p.S3Client.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(p.BucketName),
		})

		if deleteErr == nil {
			t.Logf("Successfully deleted bucket %s on %s (attempt %d)", p.BucketName, p.Name, attempt)
			return
		}

		if awsErr, ok := deleteErr.(awserr.Error); ok {
			switch awsErr.Code() {
			case "BucketNotEmpty":
				if attempt < maxRetries {
					// Try force-deleting remaining objects (including versions for B2)
					t.Logf("Bucket %s on %s still not empty (attempt %d/%d), forcing object deletion...",
						p.BucketName, p.Name, attempt, maxRetries)

					if p.Name == "Backblaze" {
						// For B2, delete all versions and delete markers
						listErr := p.S3Client.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
							Bucket: aws.String(p.BucketName),
						}, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
							var objects []*s3.ObjectIdentifier
							for _, version := range page.Versions {
								objects = append(objects, &s3.ObjectIdentifier{
									Key:       version.Key,
									VersionId: version.VersionId,
								})
							}
							for _, marker := range page.DeleteMarkers {
								objects = append(objects, &s3.ObjectIdentifier{
									Key:       marker.Key,
									VersionId: marker.VersionId,
								})
							}
							if len(objects) > 0 {
								_, deleteErr := p.S3Client.DeleteObjects(&s3.DeleteObjectsInput{
									Bucket: aws.String(p.BucketName),
									Delete: &s3.Delete{Objects: objects, Quiet: aws.Bool(true)},
								})
								if deleteErr != nil {
									t.Logf("Warning: Failed to delete some versions: %v", deleteErr)
								}
							}
							return true
						})
						if listErr != nil {
							t.Logf("Warning: Failed to list versions for forced deletion: %v", listErr)
						}
					} else {
						// Standard object deletion for non-B2 providers
						listErr := p.S3Client.ListObjectsPages(&s3.ListObjectsInput{
							Bucket: aws.String(p.BucketName),
						}, func(page *s3.ListObjectsOutput, lastPage bool) bool {
							if len(page.Contents) > 0 {
								var objects []*s3.ObjectIdentifier
								for _, obj := range page.Contents {
									objects = append(objects, &s3.ObjectIdentifier{Key: obj.Key})
								}
								_, deleteErr := p.S3Client.DeleteObjects(&s3.DeleteObjectsInput{
									Bucket: aws.String(p.BucketName),
									Delete: &s3.Delete{Objects: objects},
								})
								if deleteErr != nil {
									t.Logf("Warning: Failed to delete some objects: %v", deleteErr)
								}
							}
							return true
						})
						if listErr != nil {
							t.Logf("Warning: Failed to list objects for forced deletion: %v", listErr)
						}
					}

					// Wait longer for eventual consistency
					time.Sleep(time.Duration(attempt*2) * time.Second)
					continue
				}
				// Final attempt failed
				t.Errorf("Failed to delete bucket %s on %s after %d attempts: still not empty",
					p.BucketName, p.Name, maxRetries)
				return
			case "NoSuchBucket":
				// Bucket already deleted or doesn't exist
				t.Logf("Bucket %s on %s already deleted", p.BucketName, p.Name)
				return
			}
		}

		if attempt < maxRetries {
			t.Logf("Failed to delete bucket %s on %s (attempt %d/%d): %v, retrying...",
				p.BucketName, p.Name, attempt, maxRetries, deleteErr)
			time.Sleep(time.Duration(attempt) * time.Second)
		} else {
			t.Errorf("Failed to delete bucket %s on %s after %d attempts: %v",
				p.BucketName, p.Name, maxRetries, deleteErr)
		}
	}
}

// TestS3Connectivity tests basic S3 operations to verify connectivity
func (p *S3ProviderConfig) TestS3Connectivity(t *testing.T) {
	t.Logf("Testing S3 connectivity for %s by uploading test object", p.Name)

	// Test object upload
	testKey := "test-connectivity"
	testContent := "PBS S3 connectivity test"

	_, err := p.S3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(p.BucketName),
		Key:    aws.String(testKey),
		Body:   strings.NewReader(testContent),
	})
	require.NoError(t, err, "Failed to upload test object to %s bucket %s", p.Name, p.BucketName)

	// Test object download
	result, err := p.S3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(p.BucketName),
		Key:    aws.String(testKey),
	})
	require.NoError(t, err, "Failed to download test object from %s bucket %s", p.Name, p.BucketName)
	result.Body.Close()

	// Clean up test object
	_, err = p.S3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(p.BucketName),
		Key:    aws.String(testKey),
	})
	require.NoError(t, err, "Failed to delete test object from %s bucket %s", p.Name, p.BucketName)

	t.Logf("S3 connectivity test successful for %s", p.Name)
}

// GetPBSEndpointConfig returns the S3 endpoint configuration for PBS
func (p *S3ProviderConfig) GetPBSEndpointConfig() map[string]string {
	config := map[string]string{
		"id":         p.EndpointID,
		"endpoint":   p.Endpoint,
		"region":     p.Region,
		"access_key": p.AccessKey,
		"secret_key": p.SecretKey,
	}

	// Add provider-specific configurations
	switch p.Name {
	case "Backblaze":
		// Backblaze requires path-style addressing
		config["path_style"] = "true"
		if len(p.ProviderQuirks) > 0 {
			// Format quirks as Terraform list
			config["provider_quirks"] = fmt.Sprintf(`["%s"]`, strings.Join(p.ProviderQuirks, `", "`))
		}
	case "Scaleway":
		// Scaleway also requires path-style addressing with PBS
		config["path_style"] = "true"
	case "AWS":
		// AWS also seems to require path-style addressing with PBS
		config["path_style"] = "true"
	default:
		// Default to path-style for PBS compatibility
		config["path_style"] = "true"
	}

	return config
}
