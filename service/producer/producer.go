package producer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

type Producer struct {
	broker *kafka.Producer
	topic  string
}

// New is used to create a Producer object
func New(broker *kafka.Producer, topic string) *Producer {
	return &Producer{
		broker,
		topic,
	}
}

// PublishMessage send a message to the Kafka users_topic
func (kafkaProducer *Producer) PublishMessage(message string) error {
	producerErr := kafkaProducer.broker.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kafkaProducer.topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)
	if producerErr != nil {
		log.Error("unable to enqueue message ", message)
		return producerErr
	}
	event := <-kafkaProducer.broker.Events()
	kafkaMessage := event.(*kafka.Message)
	if kafkaMessage.TopicPartition.Error != nil {
		log.Error("Delivery failed due to error ", kafkaMessage.TopicPartition.Error)
		return kafkaMessage.TopicPartition.Error
	} else {
		log.Warn("Delivered message to offset " + kafkaMessage.TopicPartition.Offset.String() + " in partition " + kafkaMessage.TopicPartition.String())
		return nil
	}
}
