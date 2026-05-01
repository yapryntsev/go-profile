package avatar

import (
	"goph-profile/internal/feature/avatar/domain"
	"goph-profile/internal/feature/avatar/infra"
	"goph-profile/internal/feature/avatar/transport"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AvatarRepo interface {
	domain.FetchAvatarRepo
	domain.UploadAvatarRepo
}

func New(
	pool *pgxpool.Pool,
	avatarRepo AvatarRepo,
	kafkaAddr string,
	kafkaTopic string,
) (transport.Handler, func()) {
	repo := infra.NewRepo(pool)
	dsp := infra.NewEventDispatcher(kafkaAddr, kafkaTopic)

	return transport.NewHandler(
		domain.NewUseCaseGetAvatar(repo, avatarRepo),
		domain.NewUseCaseGetAvatarMetadata(repo),
		domain.NewUseCaseDeleteAvatar(repo, dsp),
		domain.NewUseCaseUploadAvatarMetadata(avatarRepo, repo, dsp),
	), dsp.Close
}
