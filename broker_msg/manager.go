package brokermsg

import (
	"context"
	"fmt"
	"sync"
)

type BrokerMsg interface {
	Handler(ctx context.Context, wg *sync.WaitGroup, handler func(context.Context, []byte) error)
	Shutdown()
}

func NewBrokerMsg(config *Config) (_ BrokerMsg, err error) {
	switch config.Name {
	case "rmq":
		return newRmq(config.Rmq)
	default:
		return nil, fmt.Errorf("%s not found in avaliable broker message", config.Name)
	}
}
