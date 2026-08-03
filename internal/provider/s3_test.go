package provider

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
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
