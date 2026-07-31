package broker

import (
	"testing"
)

func TestCooperativeStickyRebalancing(t *testing.T) {
	rebalancer := NewCooperativeRebalancer()

	rebalancer.JoinGroup("consumer_1")
	rebalancer.JoinGroup("consumer_2")

	partitions := []string{"topic_p0", "topic_p1", "topic_p2", "topic_p3"}
	assignments := rebalancer.Rebalance(partitions)

	if len(assignments) != 2 {
		t.Fatalf("Expected 2 assigned consumers, got %d", len(assignments))
	}

	// Verify consumer_1 has assignments
	c1, err := rebalancer.GetAssignment("topic_p0")
	if err != nil || c1 == "" {
		t.Errorf("Expected assignment for topic_p0")
	}

	// Scale up: Join consumer_3
	rebalancer.JoinGroup("consumer_3")
	newAssignments := rebalancer.Rebalance(partitions)

	if len(newAssignments) != 3 {
		t.Fatalf("Expected 3 assigned consumers after scale-up, got %d", len(newAssignments))
	}

	// Verify sticky assignment: topic_p0 remains with consumer_1
	c1After, _ := rebalancer.GetAssignment("topic_p0")
	if c1After != c1 {
		t.Errorf("Sticky rebalance failed: expected topic_p0 to remain with %s, got %s", c1, c1After)
	}
}
