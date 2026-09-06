package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type Event struct {
	ID         string         `json:"id" bson:"event_id"`
	Type       string         `json:"type" bson:"type"`
	OccurredAt time.Time      `json:"occurred_at" bson:"occurred_at"`
	Data       map[string]any `json:"data" bson:"data"`
}

type Publisher interface {
	Publish(context.Context, Event) error
}

type KafkaPublisher struct{ writer *kafka.Writer }

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{writer: &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}}
}

func (p *KafkaPublisher) Publish(ctx context.Context, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.ID),
		Value: payload,
		Time:  event.OccurredAt,
	})
}

func (p *KafkaPublisher) Close() error { return p.writer.Close() }
