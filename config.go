package main

import (
	"rmq_pg/buffer"

	brokermsg "rmq_pg/broker_msg"
	"rmq_pg/database"
)

type Config struct {
	Settings  map[string]string `yaml:"settings"`
	Db        *database.Config  `yaml:"db"`
	BrokerMsg *brokermsg.Config `yaml:"brokerMsg"`
	Buffer    *buffer.Config    `json:"buffer"`
}
