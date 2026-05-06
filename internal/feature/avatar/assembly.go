package avatar

import (
	"goph-profile/internal/feature/avatar/domain"
	"goph-profile/internal/feature/avatar/infra"
	"goph-profile/internal/feature/avatar/transport"
	"reflect"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type AvatarRepo interface {
	domain.FetchAvatarRepo
	domain.UploadAvatarRepo
}

type Feature struct {
	Handler transport.Handler
	Finish  func()
}

func New(
	kafkaAddr string,
	kafkaTopic string,
	pool *pgxpool.Pool,
	avatarRepo AvatarRepo,
	tracerProvider *trace.TracerProvider,
	meterProvider metric.MeterProvider,
) Feature {
	tracer := tracerProvider.Tracer(reflect.TypeOf(Feature{}).PkgPath())
	meter := meterProvider.Meter(reflect.TypeOf(Feature{}).PkgPath())

	repo := infra.NewRepo(pool, tracer)
	dsp := infra.NewEventDispatcher(tracer, kafkaAddr, kafkaTopic)

	return Feature{
		Handler: transport.NewHandler(
			domain.NewUseCaseGetAvatar(repo, avatarRepo),
			domain.NewUseCaseGetAvatarMetadata(repo),
			domain.NewUseCaseDeleteAvatar(repo, dsp),
			domain.NewUseCaseUploadAvatarMetadata(avatarRepo, repo, dsp),
			meter,
		),
		Finish: dsp.Close,
	}
}
