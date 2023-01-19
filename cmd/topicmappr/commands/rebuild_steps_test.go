package commands

import (
	"testing"

	"github.com/DataDog/kafka-kit/v4/kafkazk"
	"github.com/DataDog/kafka-kit/v4/mapper"
)

func TestNotInReplicaSet(t *testing.T) {
	rs := []int{1001, 1002, 1003}

	if notInReplicaSet(1001, rs) != false {
		t.Errorf("Expected ID 1001 in replica set")
	}

	if notInReplicaSet(1010, rs) != true {
		t.Errorf("Expected true assertion for ID 1010 in replica set")
	}
}

func TestPhasedReassignment(t *testing.T) {
	zk := kafkazk.Stub{}
	pm1, _ := zk.GetPartitionMap("test_topic")
	pm2 := pm1.Copy()

	phased := phasedReassignment(pm1, pm2)

	// These maps should be equal; phasedReassignment will be a no-op since all of
	// the pm2 leaders == the pm1 leaders.
	if eq, _ := pm2.Equal(phased); !eq {
		t.Errorf("Unexpected PartitionMap inequality")
	}

	// Strip the pm2 leaders.
	for i := range pm2.Partitions {
		pm2.Partitions[i].Replicas = pm2.Partitions[i].Replicas[1:]
	}

	// We should expect the pm1 leaders now.
	phased = phasedReassignment(pm1, pm2)

	// Check for a non no-op.
	if eq, _ := pm2.Equal(phased); eq {
		t.Errorf("Unexpected PartitionMap equality")
	}

	// Validate each partition leader.
	for i := range phased.Partitions {
		phasedOutputLeader := phased.Partitions[i].Replicas[0]
		originalLeader := pm1.Partitions[i].Replicas[0]
		if phasedOutputLeader != originalLeader {
			t.Errorf("Expected leader ID %d in phased output map, got %d", phasedOutputLeader, originalLeader)
		}
	}
}

// previous evac_leadership_test
var topic = "testTopic"
var pMapIn = mapper.PartitionMap{
	Version: 1,
	Partitions: mapper.PartitionList{
		mapper.Partition{
			Topic:     topic,
			Partition: 0,
			Replicas: []int{
				10001,
				10002,
				10003},
		},
		mapper.Partition{
			Topic:     topic,
			Partition: 1,
			Replicas: []int{
				10002,
				10001,
				10003,
			},
		},
		mapper.Partition{
			Topic:     topic,
			Partition: 3,
			Replicas: []int{
				10003,
				10002,
				10001,
			},
		},
	},
}

func TestRemoveProblemBroker(t *testing.T) {
	problemBrokerId := 10001

	pMapOut := evacuateLeadership(pMapIn, []int{problemBrokerId}, []string{topic})

	for _, partition := range pMapOut.Partitions {
		if partition.Replicas[0] == problemBrokerId {
			t.Errorf("Expected Broker ID 10001 to be evacuated from leadership")
		}
	}
}

func TestEvacTwoProblemBrokers(t *testing.T) {
	problemBrokers := []int{10001, 10002}

	pMapOut := evacuateLeadership(pMapIn, problemBrokers, []string{topic})

	for _, partition := range pMapOut.Partitions {
		if partition.Replicas[0] == problemBrokers[0] || partition.Replicas[0] == problemBrokers[1] {
			t.Errorf("Expected Broker ID 10001 and 10002 to be evacuated from leadership")
		}
	}
}

func TestNoMatchingTopicToEvac(t *testing.T) {
	pMapOut := evacuateLeadership(pMapIn, []int{10001}, []string{"some other topic"})

	for i, partition := range pMapOut.Partitions {
		for j, broker := range partition.Replicas {
			if broker != pMapIn.Partitions[i].Replicas[j] {
				t.Errorf("Expected no changes in leadership because no matching topic was passed in.")
			}
		}
	}
}

// TODO: This test currently fails because the error case calls os.Exit(1). Better way to test, or better error handling.
//func TestEvacAllBrokersForPartitionFails(t *testing.T) {
//	problemBrokers := mapper.BrokerMap{}
//	problemBrokers[10001] = &mapper.Broker{}
//	problemBrokers[10002] = &mapper.Broker{}
//	problemBrokers[10003] = &mapper.Broker{}
//
//	evacuateLeadership(pMapIn, problemBrokers)
//
//	t.Errorf("evacuateLeadership should have errored out at this point.")
//}
