package server

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

const envPrefix = ""

type serverConfig struct {
	IdleTimeout  time.Duration `envconfig:"HTTP_SERVER_IDLE_TIMEOUT" default:"60s"`
	Port         int           `envconfig:"PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"HTTP_SERVER_READ_TIMEOUT" default:"1s"`
	WriteTimeout time.Duration `envconfig:"HTTP_SERVER_WRITE_TIMEOUT" default:"2s"`
}

func LoadConfig() (serverConfig, error) {
	var config serverConfig
	err := envconfig.Process(envPrefix, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
