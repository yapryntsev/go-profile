package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type EventDispatcher struct {
	writer *kafka.Writer
	tracer trace.Tracer
}

type kafkaHeaderCarrier []kafka.Header

func (c kafkaHeaderCarrier) Get(key string) string {
	for _, h := range c {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *kafkaHeaderCarrier) Set(key string, value string) {
	for i, h := range *c {
		if h.Key == key {
			(*c)[i].Value = []byte(value)
			return
		}
	}
	*c = append(*c, kafka.Header{Key: key, Value: []byte(value)})
}

func (c kafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(c))
	for i, h := range c {
		keys[i] = h.Key
	}
	return keys
}

func NewEventDispatcher(tracer trace.Tracer, addr string, topic string) EventDispatcher {
	return EventDispatcher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(addr),
			Topic:        topic,
			RequiredAcks: kafka.RequireAll,
			Balancer:     &kafka.LeastBytes{},
		},
		tracer: tracer,
	}
}

func (e EventDispatcher) AvatarDeleted(ctx context.Context, avatarID uuid.UUID) error {
	ctx, span := e.tracer.Start(
		ctx,
		fmt.Sprintf("%s.v1 publish", e.writer.Topic),
		trace.WithAttributes(
			attribute.String("avatar id", avatarID.String()),
			attribute.String("action", "deleted"),
		),
	)
	defer span.End()

	headers := kafkaHeaderCarrier{{Key: "action", Value: []byte("delete")}}
	otel.GetTextMapPropagator().Inject(ctx, &headers)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return e.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:     []byte("avatar-id"),
			Value:   []byte(avatarID.String()),
			Headers: []kafka.Header(headers),
		},
	)
}

func (e EventDispatcher) AvatarUploaded(ctx context.Context, avatarID uuid.UUID) error {
	ctx, span := e.tracer.Start(
		ctx,
		fmt.Sprintf("%s.v1 publish", e.writer.Topic),
		trace.WithAttributes(
			attribute.String("avatar id", avatarID.String()),
			attribute.String("action", "upload"),
		),
	)
	defer span.End()

	headers := kafkaHeaderCarrier{{Key: "action", Value: []byte("upload")}}
	otel.GetTextMapPropagator().Inject(ctx, &headers)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return e.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:     []byte("avatar-id"),
			Value:   []byte(avatarID.String()),
			Headers: []kafka.Header(headers),
		},
	)
}

func (e EventDispatcher) Close() {
	_ = e.writer.Close()
}
