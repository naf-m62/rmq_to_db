package buffer

import (
	"fmt"
)

type Buffer interface {
	// SaveInEnd write msg to end buffer
	SaveInEnd(msg []byte) (count int64, err error)
	// SaveBatchInFront write batch to the beginning. Use for rollback if handler couldn't save to db
	SaveBatchInFront(msgList [][]byte) (err error)
	// GetAllValues get all values
	GetAllValues() (msgList [][]byte, err error)
}

func NewBuffer(config *Config) (_ Buffer, err error) {
	switch config.Name {
	case "redis":
		return newRedis(config.Redis)
	default:
		return nil, fmt.Errorf("%s not found in avaliable buffer", config.Name)
	}
}
