package infra

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type EventDispatcher struct {
	writer *kafka.Writer
}

func NewEventDispatcher(addr string, topic string) EventDispatcher {
	return EventDispatcher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(addr),
			Topic:        topic,
			RequiredAcks: kafka.RequireAll,
			Balancer:     &kafka.LeastBytes{},
		},
	}
}

func (e EventDispatcher) AvatarDeleted(ctx context.Context, avatarID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return e.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte("avatar-id"),
			Value: []byte(avatarID.String()),
			Headers: []kafka.Header{
				{
					Key:   "action",
					Value: []byte("delete"),
				},
			},
		},
	)
}

func (e EventDispatcher) AvatarUploaded(ctx context.Context, avatarID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return e.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte("avatar-id"),
			Value: []byte(avatarID.String()),
			Headers: []kafka.Header{
				{
					Key:   "action",
					Value: []byte("upload"),
				},
			},
		},
	)
}

func (e EventDispatcher) Close() {
	_ = e.writer.Close()
}
