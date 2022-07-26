module rmq_pg

go 1.18

require (
	github.com/lib/pq v1.10.6
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.4.0
)

require github.com/streadway/amqp v1.0.0

require (
	github.com/gomodule/redigo v1.8.9 // indirect
	gopkg.in/redis.v5 v5.2.9 // indirect
)
