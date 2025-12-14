package configs

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Database  Database
	Service   Service
	AuthToken string `envconfig:"AUTH_TOKEN" required:"true"`
	Port      string `envconfig:"PORT" default:":8080"`
}

type Database struct {
	Host              string `envconfig:"DB_HOST" default:"localhost"`
	Port              string `envconfig:"DB_PORT" required:"true"`
	User              string `envconfig:"DB_USER" required:"true"`
	Password          string `envconfig:"DB_PASSWORD" required:"true"`
	Name              string `envconfig:"DB_NAME" required:"true"`
	MaxOpenConnection int    `envconfig:"DB_MAX_OPEN_CONNECTION" default:"10"`
	MaxLifeTime       int    `envconfig:"DB_MAX_OPEN_LIFE_TIME" default:"30"`
}

type Service struct {
	ReadTimeout  int `envconfig:"SERVICE_READ_TIMEOUT" default:"10"`
	WriteTimeout int `envconfig:"SERVICE_WRITE_TIMEOUT" default:"10"`
}

func NewParsedConfig() (Config, error) {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
