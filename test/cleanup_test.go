package test

import (
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

// TestCleanupB2Buckets removes all leftover pbs-test-* buckets from Backblaze B2
// This is useful for cleaning up after failed test runs or manual testing
func TestCleanupB2Buckets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping B2 cleanup test in short mode")
	}

	// Get B2 credentials
	b2AccessKey := os.Getenv("B2_ACCESS_KEY_ID")
	b2SecretKey := os.Getenv("B2_SECRET_ACCESS_KEY")
	b2Region := os.Getenv("B2_REGION")

	if b2AccessKey == "" || b2SecretKey == "" {
		t.Skip("B2 credentials not configured (set B2_ACCESS_KEY_ID and B2_SECRET_ACCESS_KEY)")
	}

	if b2Region == "" {
		b2Region = "us-east-005" // Updated default Backblaze region to match tests
	}

	t.Logf("🧹 Starting B2 bucket cleanup...")
	t.Logf("   Region: %s", b2Region)

	// Setup S3 client for B2 using same config as working tests
	creds := credentials.NewStaticCredentials(b2AccessKey, b2SecretKey, "")
	config := &aws.Config{
		Region:           aws.String(b2Region),
		Credentials:      creds,
		Endpoint:         aws.String(fmt.Sprintf("https://s3.%s.backblazeb2.com", b2Region)),
		S3ForcePathStyle: aws.Bool(true),
		MaxRetries:       aws.Int(3),
	}

	sess, err := session.NewSession(config)
	require.NoError(t, err, "Failed to create B2 session")

	s3Client := s3.New(sess)

	// List all buckets
	t.Log("📋 Listing all B2 buckets...")
	listBucketsOutput, err := s3Client.ListBuckets(&s3.ListBucketsInput{})
	require.NoError(t, err, "Failed to list B2 buckets")

	// Find buckets to delete
	var bucketsToDelete []*s3.Bucket
	for _, bucket := range listBucketsOutput.Buckets {
		if strings.HasPrefix(*bucket.Name, "pbs-test-") {
			bucketsToDelete = append(bucketsToDelete, bucket)
		}
	}

	if len(bucketsToDelete) == 0 {
		t.Log("✅ No leftover pbs-test-* buckets found - all clean!")
		return
	}

	t.Logf("🗑️  Found %d leftover buckets to delete:", len(bucketsToDelete))
	for _, bucket := range bucketsToDelete {
		age := time.Since(*bucket.CreationDate)
		t.Logf("   - %s (created %s ago)", *bucket.Name, age.Round(time.Minute))
	}

	// Delete each bucket
	deletedCount := 0
	failedCount := 0

	for _, bucket := range bucketsToDelete {
		bucketName := *bucket.Name
		t.Logf("\n🧹 Cleaning bucket: %s", bucketName)

		if deleteB2Bucket(t, s3Client, bucketName) {
			deletedCount++
			t.Logf("   ✅ Successfully deleted: %s", bucketName)
		} else {
			failedCount++
			t.Logf("   ❌ Failed to delete: %s", bucketName)
		}
	}

	// Final summary
	separator := strings.Repeat("=", 60)
	t.Logf("\n%s", separator)
	t.Logf("🎯 CLEANUP SUMMARY")
	t.Logf("%s", separator)
	t.Logf("   Total buckets found:     %d", len(bucketsToDelete))
	t.Logf("   Successfully deleted:    %d", deletedCount)
	t.Logf("   Failed to delete:        %d", failedCount)
	t.Logf("%s", separator)

	if failedCount > 0 {
		t.Logf("⚠️  Some buckets could not be deleted. They may need manual cleanup.")
	} else {
		t.Log("✅ All leftover buckets successfully cleaned up!")
	}
}

// deleteB2Bucket deletes a single B2 bucket including all versions and delete markers
func deleteB2Bucket(t *testing.T, s3Client *s3.S3, bucketName string) bool {
	// Step 1: Delete all object versions and delete markers
	t.Logf("   📦 Deleting all object versions from: %s", bucketName)

	totalVersions := 0
	var deleteErrors []error

	err := s3Client.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	}, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		// Delete versions one at a time to avoid B2's Content-MD5 requirement
		for _, version := range page.Versions {
			totalVersions++
			_, deleteErr := s3Client.DeleteObject(&s3.DeleteObjectInput{
				Bucket:    aws.String(bucketName),
				Key:       version.Key,
				VersionId: version.VersionId,
			})
			if deleteErr != nil {
				deleteErrors = append(deleteErrors, fmt.Errorf("version %s/%s: %w", *version.Key, *version.VersionId, deleteErr))
			}
		}

		// Delete markers one at a time
		for _, marker := range page.DeleteMarkers {
			totalVersions++
			_, deleteErr := s3Client.DeleteObject(&s3.DeleteObjectInput{
				Bucket:    aws.String(bucketName),
				Key:       marker.Key,
				VersionId: marker.VersionId,
			})
			if deleteErr != nil {
				deleteErrors = append(deleteErrors, fmt.Errorf("marker %s/%s: %w", *marker.Key, *marker.VersionId, deleteErr))
			}
		}

		if totalVersions > 0 && totalVersions%10 == 0 {
			t.Logf("      Deleted %d objects so far...", totalVersions)
		}

		return true // Continue to next page
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucket" {
			t.Logf("   ℹ️  Bucket already deleted: %s", bucketName)
			return true
		}
		t.Logf("   ❌ Failed to list object versions: %v", err)
		return false
	}

	if len(deleteErrors) > 0 {
		t.Logf("   ⚠️  Encountered %d errors while deleting objects (showing first 3):", len(deleteErrors))
		for i, err := range deleteErrors {
			if i >= 3 {
				break
			}
			t.Logf("      - %v", err)
		}
	}

	if totalVersions > 0 {
		t.Logf("   ✓ Deleted %d object versions/markers", totalVersions)
		// Wait for B2 eventual consistency
		time.Sleep(3 * time.Second)
	} else {
		t.Logf("   ✓ No objects to delete (empty bucket)")
	}

	// Step 2: Verify bucket is empty
	listOutput, err := s3Client.ListObjectVersions(&s3.ListObjectVersionsInput{
		Bucket:  aws.String(bucketName),
		MaxKeys: aws.Int64(1),
	})

	if err != nil {
		t.Logf("   ⚠️  Warning: Could not verify bucket is empty: %v", err)
	} else if len(listOutput.Versions) > 0 || len(listOutput.DeleteMarkers) > 0 {
		remaining := len(listOutput.Versions) + len(listOutput.DeleteMarkers)
		t.Logf("   ⚠️  Warning: Bucket still has %d+ objects/markers remaining", remaining)
	}

	// Step 3: Delete the bucket with retries
	t.Logf("   🗑️  Deleting bucket: %s", bucketName)

	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err = s3Client.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})

		if err == nil {
			return true
		}

		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case "NoSuchBucket":
				// Already deleted
				return true
			case "BucketNotEmpty":
				if attempt < maxRetries {
					t.Logf("      Bucket still not empty, waiting (attempt %d/%d)...", attempt, maxRetries)
					time.Sleep(time.Duration(attempt*2) * time.Second)
					continue
				}
				t.Logf("   ❌ Bucket still not empty after %d attempts", maxRetries)
				return false
			default:
				t.Logf("   ❌ Delete failed: %s - %s", awsErr.Code(), awsErr.Message())
				return false
			}
		}

		if attempt < maxRetries {
			t.Logf("      Retrying delete (attempt %d/%d)...", attempt, maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	t.Logf("   ❌ Failed to delete bucket after %d attempts: %v", maxRetries, err)
	return false
}

// TestCleanupAllProviderBuckets removes leftover test buckets from all configured S3 providers
func TestCleanupAllProviderBuckets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping all-provider cleanup test in short mode")
	}

	providers := GetS3ProviderConfigs(t)
	if len(providers) == 0 {
		t.Skip("No S3 providers configured for cleanup")
	}

	t.Logf("🧹 Starting cleanup for %d S3 provider(s)...\n", len(providers))

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			cleanupProviderBuckets(t, provider)
		})
	}

	t.Log("\n✅ Cleanup complete for all providers!")
}

// cleanupProviderBuckets removes all pbs-test-* buckets for a specific provider
func cleanupProviderBuckets(t *testing.T, provider *S3ProviderConfig) {
	t.Logf("🧹 Cleaning up %s buckets...", provider.Name)

	// Setup S3 client
	provider.SetupS3Client(t)

	// List all buckets
	listBucketsOutput, err := provider.S3Client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		t.Logf("   ❌ Failed to list buckets: %v", err)
		return
	}

	// Find buckets to delete based on provider prefix
	var prefix string
	switch provider.Name {
	case "AWS":
		prefix = "pbs-test-aws-"
	case "Backblaze":
		prefix = "pbs-test-b2-"
	case "Scaleway":
		prefix = "pbs-test-scw-"
	default:
		prefix = "pbs-test-"
	}

	var bucketsToDelete []string
	for _, bucket := range listBucketsOutput.Buckets {
		if strings.HasPrefix(*bucket.Name, prefix) {
			bucketsToDelete = append(bucketsToDelete, *bucket.Name)
		}
	}

	if len(bucketsToDelete) == 0 {
		t.Logf("   ✅ No leftover buckets found for %s", provider.Name)
		return
	}

	t.Logf("   Found %d bucket(s) to delete", len(bucketsToDelete))

	deletedCount := 0
	for _, bucketName := range bucketsToDelete {
		t.Logf("   🗑️  Deleting: %s", bucketName)

		// Use provider's delete method (handles versioning for B2)
		provider.BucketName = bucketName
		provider.DeleteTestBucket(t)
		deletedCount++
	}

	t.Logf("   ✅ Deleted %d bucket(s) from %s", deletedCount, provider.Name)
}
