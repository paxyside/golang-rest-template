package config

import (
	"emperror.dev/errors"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURI string `mapstructure:"DB_URI"`
	RabbitMQURI string `mapstructure:"AMQP_URI"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, errors.Wrap(err, "godotenv.Load")
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")

	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}

	return &cfg, nil
}
