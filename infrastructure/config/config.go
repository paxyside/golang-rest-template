package config

import (
	"emperror.dev/errors"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		return errors.Wrap(err, "godotenv.Load")
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return errors.Wrap(err, "viper.ReadInConfig")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("")

	return nil
}
