package provider

import (
	"context"
	"testing"
)

func TestProvider(t *testing.T) {
	p := Provider()
	if err := p.InternalValidate(); err != nil {
		t.Fatalf("Provider validation failed: %v", err)
	}
}

func TestResourceBucketCreate(t *testing.T) {
	p := Provider()
	resource := p.ResourcesMap["pranor_bucket"]
	d := resource.TestResourceData()
	d.Set("name", "my-test-bucket")

	diags := resourceBucketCreate(context.Background(), d, nil)
	if diags.HasError() {
		t.Fatalf("Bucket creation failed: %v", diags)
	}

	if d.Id() != "my-test-bucket" {
		t.Errorf("Expected ID my-test-bucket, got %s", d.Id())
	}
}

func TestResourceTopicCreate(t *testing.T) {
	p := Provider()
	resource := p.ResourcesMap["pranor_topic"]
	d := resource.TestResourceData()
	d.Set("name", "events-topic")

	diags := resourceTopicCreate(context.Background(), d, nil)
	if diags.HasError() {
		t.Fatalf("Topic creation failed: %v", diags)
	}

	if d.Id() != "topic-events-topic" {
		t.Errorf("Expected ID topic-events-topic, got %s", d.Id())
	}
}
