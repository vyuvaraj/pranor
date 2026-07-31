package import (
	"fmt"
	"sort"
	"sync"
)

type PartitionAssignment struct {
	ConsumerID string   `json:"consumer_id"`
	Partitions []string `json:"partitions"`
}

type CooperativeRebalancer struct {
	mu           sync.Mutex
	assignments  map[string]string // partition -> consumerID
	groupMembers map[string]bool
}

func NewCooperativeRebalancer() *CooperativeRebalancer {
	return &CooperativeRebalancer{
		assignments:  make(map[string]string),
		groupMembers: make(map[string]bool),
	}
}

func (c *CooperativeRebalancer) JoinGroup(consumerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.groupMembers[consumerID] = true
}

func (c *CooperativeRebalancer) LeaveGroup(consumerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.groupMembers, consumerID)
	// Remove assignments for left consumer
	for partition, assignedConsumer := range c.assignments {
		if assignedConsumer == consumerID {
			delete(c.assignments, partition)
		}
	}
}

// Rebalance performs cooperative sticky partition rebalancing across active consumers.
func (c *CooperativeRebalancer) Rebalance(allPartitions []string) []PartitionAssignment {
	c.mu.Lock()
	defer c.mu.Unlock()

	var activeConsumers []string
	for consumerID := range c.groupMembers {
		activeConsumers = append(activeConsumers, consumerID)
	}
	sort.Strings(activeConsumers)

	if len(activeConsumers) == 0 {
		return nil
	}

	// 1. Keep existing valid assignments ("Sticky")
	unassignedPartitions := []string{}
	for _, partition := range allPartitions {
		currentConsumer, exists := c.assignments[partition]
		if !exists || !c.groupMembers[currentConsumer] {
			unassignedPartitions = append(unassignedPartitions, partition)
		}
	}

	// 2. Round-robin assign unassigned partitions
	for i, partition := range unassignedPartitions {
		assignedConsumer := activeConsumers[i%len(activeConsumers)]
		c.assignments[partition] = assignedConsumer
	}

	// 3. Build result group assignments
	consumerMap := make(map[string][]string)
	for partition, consumerID := range c.assignments {
		consumerMap[consumerID] = append(consumerMap[consumerID], partition)
	}

	var result []PartitionAssignment
	for _, consumerID := range activeConsumers {
		partitions := consumerMap[consumerID]
		sort.Strings(partitions)
		result = append(result, PartitionAssignment{
			ConsumerID: consumerID,
			Partitions: partitions,
		})
	}

	return result
}

func (c *CooperativeRebalancer) GetAssignment(partition string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	consumerID, exists := c.assignments[partition]
	if !exists {
		return "", fmt.Errorf("partition %s unassigned", partition)
	}
	return consumerID, nil
}
