package infra

import (
	"context"
	"errors"
	"goph-profile/internal/feature/thumbnail/domain/model"
	"io"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type EventReceiver struct {
	reader *kafka.Reader
	logger *zap.Logger
}

func NewEventReceiver(logger *zap.Logger, addr string, topic string) EventReceiver {
	return EventReceiver{
		reader: kafka.NewReader(
			kafka.ReaderConfig{
				Brokers:     []string{addr},
				GroupID:     "group",
				Topic:       topic,
				MinBytes:    1,
				MaxBytes:    10e6,
				StartOffset: kafka.FirstOffset,
			},
		),
		logger: logger,
	}
}

func (e EventReceiver) Observe(ctx context.Context, callback func(event model.Event)) {
	for {
		msg, err := e.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			e.logger.Error("failed to read message", zap.Error(err))
			continue
		}

		if len(msg.Headers) == 0 {
			e.logger.Error("expect header with action type", zap.String("key", string(msg.Key)))
			continue
		}

		if msg.Headers[0].Key != "action" {
			e.logger.Error("expect header", zap.String("key", string(msg.Headers[0].Key)))
			continue
		}

		switch string(msg.Headers[0].Value) {
		case "delete":
			avatarID, err := uuid.Parse(string(msg.Value))
			if err != nil {
				e.logger.Error("failed to parse avatar ID", zap.String("key", string(msg.Value)))
				continue
			}
			callback(model.DeleteEvent{AvatarID: avatarID})
		case "upload":
			avatarID, err := uuid.Parse(string(msg.Value))
			if err != nil {
				e.logger.Error("failed to parse avatar ID", zap.String("key", string(msg.Value)))
				continue
			}
			callback(model.UploadEvent{AvatarID: avatarID})
		default:
			e.logger.Error("unexpected header with action type", zap.String("key", string(msg.Headers[0].Key)))
		}
	}
}
