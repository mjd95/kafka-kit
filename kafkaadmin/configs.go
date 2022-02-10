package kafkaadmin

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	topicResourceType, _  = kafka.ResourceTypeFromString("topic")
	brokerResourceType, _ = kafka.ResourceTypeFromString("broker")
)

// ResourceConfigs is a map of resource name to a map of configuration name
// and configuration.
// Example: map["my_topic"]map["retention.ms"] = 'retention.ms="4000000"'
type ResourceConfigs map[string]map[string]kafka.ConfigEntryResult

// AddConfig takes a resource name (ie a broker ID or topic name) and a
// kafka.ConfigEntryResult. It populates the kafka.ConfigEntryResult in the
// ResourceConfigs keyed by the provided resource name.
func (rc ResourceConfigs) AddConfig(name string, config kafka.ConfigEntryResult) error {
	if _, ok := rc[name]; !ok {
		rc[name] = make(map[string]kafka.ConfigEntryResult)
	}

	if config.Name == "" {
		return fmt.Errorf("empty ConfigEntryResult name")
	}

	rc[name][config.Name] = config

	return nil
}

// DynamicConfigMapForResources takes a kafka resource type (ie topic, broker) and
// list of names and returns a ResourceConfigs for all dynamic configurations
// discovered for each resource by name.
func (c Client) DynamicConfigMapForResources(ctx context.Context, kind string, names []string) (ResourceConfigs, error) {
	var configResources []kafka.ConfigResource

	var ckgType kafka.ResourceType
	switch kind {
	case "topic":
		ckgType = topicResourceType
	case "broker":
		ckgType = brokerResourceType
	default:
		return nil, fmt.Errorf("invalid resource type")
	}

	for _, n := range names {
		fmt.Println(topicResourceType)
		cr := kafka.ConfigResource{
			Type: ckgType,
			Name: n,
		}
		configResources = append(configResources, cr)
	}

	resourceConfigs, err := c.c.DescribeConfigs(ctx, configResources)
	if err != nil {
		return nil, err
	}

	var results = make(ResourceConfigs)
	for _, config := range resourceConfigs {
		for _, v := range config.Config {
			if v.Source == kafka.ConfigSourceDynamicTopic || v.Source == kafka.ConfigSourceDynamicBroker {
				results.AddConfig(config.Name, v)
			}
		}
	}

	return results, nil
}
