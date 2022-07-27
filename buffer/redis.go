package buffer

import (
	"time"

	redisLib "github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

type redis struct {
	config *ConfigRedis
	client *redisLib.Pool
}

const redisTimeout = 30 * time.Second

var ErrTypeAssertions = errors.New("Type assertions errors")

func newRedis(config *ConfigRedis) (Buffer, error) {
	r := &redis{
		config: config,
		client: &redisLib.Pool{
			Dial: func() (redisLib.Conn, error) {
				return redisLib.Dial("tcp", config.Addr)
			},
			IdleTimeout: time.Duration(config.IdleTimeout) * time.Second,
		},
	}
	_, err := r.client.Dial()

	return r, err
}

func (r *redis) SaveInEnd(msg []byte) (count int64, err error) {
	var redisConn = r.client.Get()
	defer redisConn.Close()

	var countI interface{}
	if countI, err = redisLib.DoWithTimeout(
		redisConn,
		redisTimeout,
		"RPUSH",
		r.config.Key,
		msg,
	); err != nil {
		return 0, errors.Wrap(err, "RPUSH error")
	}

	var ok bool
	if count, ok = countI.(int64); !ok {
		return 0, errors.Wrap(ErrTypeAssertions, "can't convert to int64")
	}
	return count, nil
}

func (r *redis) SaveBatchInFront(msgList [][]byte) (err error) {
	var redisConn = r.client.Get()
	defer redisConn.Close()

	var msgListWithKey = make([]interface{}, 0, len(msgList)+1)
	msgListWithKey = append(msgListWithKey, r.config.Key)

	for _, bytes := range msgList {
		msgListWithKey = append(msgListWithKey, bytes)
	}

	if _, err = redisLib.DoWithTimeout(
		redisConn,
		redisTimeout,
		"LPUSH",
		msgListWithKey...,
	); err != nil {
		return errors.Wrap(err, "LPUSH error")
	}
	return nil
}

func (r *redis) GetAllValues() (msgList [][]byte, err error) {
	var redisConn = r.client.Get()
	defer redisConn.Close()

	var countI interface{}
	if countI, err = redisLib.DoWithTimeout(
		redisConn,
		redisTimeout,
		"LLEN",
		r.config.Key,
	); err != nil {
		return nil, errors.Wrap(err, "LLEN error")
	}

	var (
		ok    bool
		count int64
	)
	if count, ok = countI.(int64); !ok {
		return nil, errors.Wrap(ErrTypeAssertions, "can't get count values in key")
	}

	var msgI interface{}
	if msgI, err = redisLib.DoWithTimeout(
		redisConn,
		redisTimeout,
		"LPOP",
		r.config.Key,
		count,
	); err != nil {
		return nil, errors.Wrap(err, "LPOP error")
	}

	var msgSI []interface{}
	if msgSI, ok = msgI.([]interface{}); !ok {
		return nil, errors.Wrap(ErrTypeAssertions, "can't convert to []interface")
	}

	for _, k := range msgSI {
		var b []byte
		if b, ok = k.([]byte); !ok {
			return nil, errors.Wrap(ErrTypeAssertions, "can't convert to []byte")
		}
		if len(b) == 0 {
			continue
		}
		msgList = append(msgList, b)
	}

	return msgList, nil
}
