package brokermsg

type Config struct {
	Name string     `yaml:"name"`
	Rmq  *RmqConfig `yaml:"rmq"`
}

type RmqConfig struct {
	Dsn             string `yaml:"dsn"`
	Exchanger       string `yaml:"exchanger"`
	RoutingKey      string `yaml:"routingKey"`
	PreferQueueName string `yaml:"preferQueueName"`
}
