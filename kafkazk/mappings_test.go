package kafkazk

import (
	"sort"
	"testing"
)

func TestMappings(t *testing.T) {
	var topic string = "test_topic"
	pm, _ := PartitionMapFromString(testGetMapString4(topic))
	mappings := pm.Mappings()

	expected := map[int]partitionList{}

	expected[1001] = partitionList{
		Partition{Topic: topic, Partition: 2, Replicas: []int{1001, 1002}},
		Partition{Topic: topic, Partition: 4, Replicas: []int{1001, 1003}},
		Partition{Topic: topic, Partition: 5, Replicas: []int{1002, 1001}},
	}

	expected[1002] = partitionList{
		Partition{Topic: topic, Partition: 2, Replicas: []int{1001, 1002}},
		Partition{Topic: topic, Partition: 3, Replicas: []int{1003, 1002}},
		Partition{Topic: topic, Partition: 5, Replicas: []int{1002, 1001}},
	}

	expected[1003] = partitionList{
		Partition{Topic: topic, Partition: 0, Replicas: []int{1004, 1003}},
		Partition{Topic: topic, Partition: 1, Replicas: []int{1003, 1004}},
		Partition{Topic: topic, Partition: 3, Replicas: []int{1003, 1002}},
		Partition{Topic: topic, Partition: 4, Replicas: []int{1001, 1003}},
	}

	expected[1004] = partitionList{
		Partition{Topic: topic, Partition: 0, Replicas: []int{1004, 1003}},
		Partition{Topic: topic, Partition: 1, Replicas: []int{1003, 1004}},
	}

	for id := range expected {
		sort.Sort(mappings[id]["test_topic"])
		sort.Sort(expected[id])

		if len(mappings[id]["test_topic"]) != len(expected[id]) {
			t.Errorf("Broker %d: expected partitionList len of %d, got %d",
				id, len(expected[id]), len(mappings[id]["test_topic"]))
			continue
		}

		for i, p := range mappings[id]["test_topic"] {
			if !p.Equal(expected[id][i]) {
				t.Errorf("Broker %d partitionList[%d]: Expected %+v, got %+v",
					id, i, expected[id][i], p)
			}
		}
	}
}

func TestLArgestPartitions(t *testing.T) {
	var topic string = "test_topic"
	pm, _ := PartitionMapFromString(testGetMapString4(topic))
	zk := &Mock{}
	pmm, _ := zk.GetAllPartitionMeta()
	mappings := pm.Mappings()

	l, _ := mappings.LargestPartitions(1003, 3, pmm)

	expected := []int{3, 4, 1}

	if len(l) != 3 {
		t.Errorf("Expected result length 3, got %d", len(l))
	}

	for i, p := range l {
		if p.Partition != expected[i] {
			t.Errorf("Expected partition # %d, got %d", expected[i], p.Partition)
		}
	}
}
