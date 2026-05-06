package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// App represents the application instance configuration.
type App struct {
	Env  Env    `env:"ENV" env-description:"app environment" env-required:"true"`
	Name string `env:"APP_NAME" env-description:"app name" env-required:"true"`
	HTTP struct {
		Addr string `env:"APP_HTTP_ADDR" env-description:"http listen address" env-required:"true"`
	}
	JWT struct {
		Secret string        `env:"APP_JWT_SECRET" env-description:"jwt secret key" env-required:"true"`
		TTL    time.Duration `env:"APP_JWT_TTL" env-description:"jwt time to live" env-required:"true"`
	}
	DB struct {
		Conn string `env:"APP_DB_CONNECT" env-description:"database connection string" env-required:"true"`
	}
	S3 struct {
		Addr            string `env:"APP_S3_ADDR" env-description:"s3 address" env-required:"true"`
		AccessKeyID     string `env:"APP_S3_ACCESS_KEY_ID" env-description:"s3 access key id" env-required:"true"`
		SecretAccessKey string `env:"APP_S3_SECRET_ACCESS_KEY" env-description:"s3 secret access key" env-required:"true"`
		BucketName      string `env:"APP_S3_BUCKET_NAME" env-description:"s3 bucket name" env-required:"true"`
	}
	Kafka struct {
		Addr  string `env:"APP_KAFKA_ADDR" env-description:"kafka address" env-required:"true"`
		Topic string `env:"APP_KAFKA_TOPIC" env-description:"kafka topic" env-required:"true"`
	}
}

// New creates a new instance of the app's configuration.
func New() (App, error) {
	cfg := App{}
	err := cleanenv.ReadEnv(&cfg)

	return cfg, err
}

// IsDev returns true when the current environment is set to development.
func (a App) IsDev() bool {
	return a.Env == Dev
}

// IsProd returns true when the current environment is set to production.
func (a App) IsProd() bool {
	return a.Env == Prod
}
