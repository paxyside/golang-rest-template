package application

import (
	"emperror.dev/errors"
	"github.com/spf13/viper"
	"project_reference/infrastructure/database"
	"project_reference/infrastructure/rabbit"
)

type Infra struct {
	db     *database.DB
	rabbit *rabbit.RabbitMQ
}

func (i *Infra) Close() {
	i.db.Close()
	i.rabbit.Close()
}

func (i *Infra) GetDB() *database.DB {
	return i.db
}

func (i *Infra) GetMQ() *rabbit.RabbitMQ {
	return i.rabbit
}

func setupInfra() (*Infra, error) {
	var dbUri string

	if dbUri = viper.GetString("DB_URI"); dbUri == "" {
		return nil, errors.New("DB_URI is empty")
	}

	db, err := database.Init(dbUri)
	if err != nil {
		return nil, errors.Wrap(err, "database.Init")
	}

	var amqpUri string

	if amqpUri = viper.GetString("AMQP_URI"); amqpUri == "" {
		return nil, errors.New("AMQP_URI is empty")
	}

	mq, err := rabbit.NewRabbitMQ(amqpUri)
	if err != nil {
		return nil, errors.Wrap(err, "rabbit.NewRabbitMQ")
	}

	return &Infra{
		db:     db,
		rabbit: mq,
	}, nil
}
