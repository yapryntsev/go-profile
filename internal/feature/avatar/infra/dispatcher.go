package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type EventDispatcher struct {
	writer *kafka.Writer
}

func NewEventDispatcher(addr string, topic string) EventDispatcher {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to connect to kafka: %w", err))
		return EventDispatcher{}
	}
	conn.Close()

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
