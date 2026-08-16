package provider

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

func TestS3ProviderMinMax(t *testing.T) {
	provider := S3Provider{name: "test"}
	objects := []*s3.Object{
		{Key: aws.String("other.v9.hcstate")},
		{Key: aws.String("test.v2.hcstate"), LastModified: aws.Time(time.Unix(2, 0))},
		{Key: aws.String("test.v5.hcstate"), LastModified: aws.Time(time.Unix(5, 0))},
		{Key: aws.String("test.v3.hcstate"), LastModified: aws.Time(time.Unix(3, 0))},
	}

	minimum, maximum, latest, err := provider.minMax(objects)
	if err != nil {
		t.Fatalf("minMax returned an error: %v", err)
	}
	if minimum != 2 {
		t.Fatalf("expected minimum revision 2, got %d", minimum)
	}
	if maximum != 5 {
		t.Fatalf("expected maximum revision 5, got %d", maximum)
	}
	if latest != objects[2] {
		t.Fatalf("expected latest object %p, got %p", objects[2], latest)
	}
}

func TestS3ProviderListsAllPages(t *testing.T) {
	client := &pagedS3Client{pages: [][]*s3.Object{
		{{Key: aws.String("test.v1.hcstate")}},
		{{Key: aws.String("test.v2.hcstate")}},
	}}
	provider := S3Provider{
		name:   "test",
		bucket: aws.String("bucket"),
		prefix: aws.String("prefix"),
		client: client,
	}

	objects, err := provider.listObjects()
	if err != nil {
		t.Fatalf("listObjects returned an error: %v", err)
	}
	if len(objects) != 2 {
		t.Fatalf("expected objects from two pages, got %d", len(objects))
	}
}

type pagedS3Client struct {
	s3iface.S3API
	pages [][]*s3.Object
}

func (c *pagedS3Client) ListObjectsV2Pages(_ *s3.ListObjectsV2Input, callback func(*s3.ListObjectsV2Output, bool) bool) error {
	for index, objects := range c.pages {
		if !callback(&s3.ListObjectsV2Output{Contents: objects}, index == len(c.pages)-1) {
			break
		}
	}
	return nil
}
