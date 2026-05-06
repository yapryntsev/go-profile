package infra

import (
	"context"
	"errors"
	"fmt"
	"goph-profile/internal/feature/thumbnail/domain/model"
	"io"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type EventReceiver struct {
	reader *kafka.Reader
	logger *zap.Logger
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

func NewEventReceiver(tracer trace.Tracer, logger *zap.Logger, addr string, topic string) EventReceiver {
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
		tracer: tracer,
	}
}

func (e EventReceiver) Observe(ctx context.Context, callback func(event model.Event)) {
	for {
		msg, err := e.reader.ReadMessage(ctx)

		carrier := kafkaHeaderCarrier(msg.Headers)
		msgCtx := otel.GetTextMapPropagator().Extract(ctx, &carrier)

		_, span := e.tracer.Start(msgCtx, fmt.Sprintf("%s.v1 receive", msg.Topic))

		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			span.End()

			e.logger.Error("failed to read message", zap.Error(err))
			continue
		}

		if len(msg.Headers) == 0 {
			span.SetStatus(codes.Error, "expect header with action type")
			span.RecordError(err)
			span.End()

			e.logger.Error("expect header with action type", zap.String("key", string(msg.Key)))
			continue
		}

		if msg.Headers[0].Key != "action" {
			span.SetStatus(codes.Error, "expect header")
			span.RecordError(err)
			span.End()

			e.logger.Error("expect header", zap.String("key", string(msg.Headers[0].Key)))
			continue
		}

		span.End()

		_, span = e.tracer.Start(
			msgCtx,
			fmt.Sprintf("%s.v1 process", msg.Topic),
			trace.WithAttributes(
				attribute.String("action", string(msg.Headers[0].Value)),
			),
		)

		switch string(msg.Headers[0].Value) {
		case "delete":
			avatarID, err := uuid.Parse(string(msg.Value))
			span.SetAttributes(attribute.String("avatar id", avatarID.String()))

			if err != nil {
				span.SetStatus(codes.Error, "failed to parse avatar ID")
				span.RecordError(err)
				span.End()

				e.logger.Error("failed to parse avatar ID", zap.String("key", string(msg.Value)))
				continue
			}
			callback(model.DeleteEvent{AvatarID: avatarID})
		case "upload":
			avatarID, err := uuid.Parse(string(msg.Value))
			span.SetAttributes(attribute.String("avatar id", avatarID.String()))

			if err != nil {
				span.SetStatus(codes.Error, "failed to parse avatar ID")
				span.RecordError(err)
				span.End()

				e.logger.Error("failed to parse avatar ID", zap.String("key", string(msg.Value)))
				continue
			}
			callback(model.UploadEvent{AvatarID: avatarID})
		default:
			span.SetStatus(codes.Error, "unexpected header with action type")
			span.RecordError(err)

			e.logger.Error("unexpected header with action type", zap.String("key", string(msg.Headers[0].Key)))
		}

		span.End()
	}
}
