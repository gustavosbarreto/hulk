package main

import (
	"reflect"
	"testing"

	"fmt"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
)

func TestSubscribe(t *testing.T) {
	subscriber, _ := NewSubscriber()
	subscriber.Topics = []string{"topic"}

	broker := NewBroker()
	err := broker.Subscribe(subscriber)
	assert.NoError(t, err, "Failed to subscribe to topic")

	for _, topic := range subscriber.Topics {
		if s, ok := broker.Subscriptions[topic]; ok {
			assert.Equal(t, subscriber, s)
		} else {
			assert.Fail(t, fmt.Sprintf("Subscription list doesn't contains topic: %s", topic))
		}
	}
}

func TestUnsubscribe(t *testing.T) {
	topic := "topic"

	subscriber, _ := NewSubscriber()
	subscriber.Topics = []string{topic}

	broker := NewBroker()
	broker.Subscribe(subscriber)

	err := broker.Unsubscribe(subscriber, topic)
	assert.NoError(t, err, "Failed to unsubscribe from topic")
}

func TestPublish(t *testing.T) {
	topics := []string{"topic1", "topic2"}
	published := []string{}

	subscriber, _ := NewSubscriber()
	subscriber.Topics = topics

	monkey.PatchInstanceMethod(reflect.TypeOf(subscriber), "Receiver", func(_ *Subscriber, topic string) {
		published = append(published, topic)
	})

	broker := NewBroker()
	broker.Subscribe(subscriber)

	for _, topic := range topics {
		broker.Publish(topic)
	}

	assert.Equal(t, topics, published)
}
