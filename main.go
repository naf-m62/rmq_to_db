package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	brokermsg "rmq_pg/broker_msg"
	"rmq_pg/buffer"
	"rmq_pg/database"
)

type worker struct {
	config    *Config
	repo      database.DBManager
	brokerMsg brokermsg.BrokerMsg
	buffer    buffer.Buffer
}

func main() {
	var (
		err    error
		config *Config
	)

	if config, err = parseConfig(); err != nil {
		log.Println("parseConfig error:", err)
		return
	}

	var db *sql.DB
	if db, err = database.ConnDB(config.Db); err != nil {
		log.Fatal("connection to db failed, error:", err)
	}

	var bMsg brokermsg.BrokerMsg
	if bMsg, err = brokermsg.NewBrokerMsg(config.BrokerMsg); err != nil {
		log.Fatal("connection to rmq failed, error:", err)
	}

	var buf buffer.Buffer
	if buf, err = buffer.NewBuffer(config.Buffer); err != nil {
		log.Fatal("connection to buffer error", err)
	}

	w := worker{
		config:    config,
		repo:      database.NewDBManager(db, config.Db.TableName),
		brokerMsg: bMsg,
		buffer:    buf,
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	go func() {
		w.brokerMsg.Handler(ctx, wg, w.handle)
	}()

	var stop = make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-stop
	cancel()
	wg.Wait()
}

func parseConfig() (config *Config, err error) {
	var (
		filename   string
		configData []byte
	)
	flag.StringVar(&filename, "c", "config.yaml", "config filename (shorthand)")
	flag.Parse()

	config = &Config{}
	if configData, err = ioutil.ReadFile(filename); err != nil {
		return nil, errors.Wrap(err, "can't read config file")
	}

	if err = yaml.Unmarshal(configData, &config); err != nil {
		return nil, errors.Wrap(err, "can't unmarshal config")
	}

	return config, nil
}

func (w *worker) handle(ctx context.Context, msgJson []byte) (err error) {
	var count int64
	if count, err = w.buffer.SaveInEnd(msgJson); err != nil {
		log.Println("can't save to buffer", err)
		return err
	}

	if count < w.config.Buffer.Size {
		return nil
	}

	var msgList [][]byte
	if msgList, err = w.buffer.GetAllValues(); err != nil {
		log.Println("can't get batch from buffer", err)
		return err
	}

	depth := w.getMsgDepth()

	var valueMapList []map[string]interface{}
	for _, bytes := range msgList {
		msg := make(map[string]interface{})
		if err = json.Unmarshal(bytes, &msg); err != nil {
			log.Println("can't unmarshal msg", err)
			continue
		}

		var valueMap map[string]interface{}
		valueMap = w.findKey(msg, "", valueMap, 0, depth)

		valueMapList = append(valueMapList, valueMap)
	}

	if err = w.repo.Repo().SaveBatch(ctx, w.config.Settings, valueMapList); err != nil {
		log.Println("can't save batch in db", err, "try save back to buffer")

		if errR := w.buffer.SaveBatchInFront(msgList); errR != nil {
			log.Println("can't save back to buffer", errR)
		}
		return err
	}

	return nil
}

func (w worker) getMsgDepth() (d int) {
	for k, _ := range w.config.Settings {
		if d < strings.Count(k, ".") {
			d = strings.Count(k, ".")
		}
	}
	return d + 1
}

func (w *worker) findKey(
	msg map[string]interface{},
	prefix string,
	valueMap map[string]interface{},
	current, depth int,
) map[string]interface{} {
	for k, v := range msg {
		if current >= depth {
			continue
		}

		if _, ok := v.(map[string]interface{}); !ok {
			if valueMap == nil {
				valueMap = make(map[string]interface{})
			}
			if vv, ok := v.(interface{}); ok {
				valueMap[getPrefix(prefix, k)] = vv
				continue
			}
			if vv, ok := v.([]interface{}); ok {
				valueMap[getPrefix(prefix, k)] = vv
			}
			continue
		}

		valueMap = w.findKey(v.(map[string]interface{}), getPrefix(prefix, k), valueMap, current+1, depth)
	}

	return valueMap
}

func getPrefix(initPrefix, key string) string {
	if initPrefix == "" {
		return key
	}
	return initPrefix + "." + key
}
