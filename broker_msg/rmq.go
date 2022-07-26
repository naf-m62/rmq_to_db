package brokermsg

import (
	"context"
	"log"
	"sync"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type rmq struct {
	config     *RmqConfig
	rmqConn    *amqp.Connection
	rmqChannel *amqp.Channel
}

func newRmq(config *RmqConfig) (r *rmq, err error) {
	r = &rmq{
		config: config,
	}
	if err = r.prepareRmq(config); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *rmq) prepareRmq(config *RmqConfig) (err error) {
	if r.rmqConn, err = amqp.Dial(config.Dsn); err != nil {
		return errors.Wrap(err, "amqp.Dial error")
	}
	defer func() {
		if err != nil {
			r.Shutdown()
		}
	}()

	if r.rmqChannel, err = r.rmqConn.Channel(); err != nil {
		return errors.Wrap(err, "can't create channel")
	}

	log.Println("declaring exchange", r.config.Exchanger)
	if err = r.rmqChannel.ExchangeDeclare(
		r.config.Exchanger,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return errors.Wrap(err, "can't exchange declare")
	}

	var (
		queueName = r.getQueueName()
		queue     amqp.Queue
	)
	log.Println("declaring queue", queueName)
	if queue, err = r.rmqChannel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return errors.Wrap(err, "can't queue declare")
	}

	log.Println("queue binding")
	if err = r.rmqChannel.QueueBind(
		queue.Name,
		r.config.RoutingKey,
		r.config.Exchanger,
		false,
		nil,
	); err != nil {
		return errors.Wrap(err, "can't queue bind")
	}

	return nil
}

func (r *rmq) Handler(ctx context.Context, wg *sync.WaitGroup, handler func(context.Context, []byte) error) {
	wg.Add(1)
	defer func() {
		r.Shutdown()
		wg.Done()
	}()

	log.Println("start handle")
	wg.Add(1)
	go r.handler(ctx, wg, handler)

	<-ctx.Done()
	log.Println("Interrupting...")
}

func (r *rmq) handler(ctx context.Context, wg *sync.WaitGroup, handler func(context.Context, []byte) error) {
	defer wg.Done()

	var (
		d   <-chan amqp.Delivery
		err error
	)
	if d, err = r.rmqChannel.Consume(
		r.getQueueName(),
		"",
		false,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Println("can't create consume:", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-d:
			if err = handler(ctx, msg.Body); err != nil {
				continue
			}
			if err = msg.Ack(false); err != nil {
				log.Println("ack error:", err, "MessageId:", msg.MessageId)
			}
		}
	}
}

func (r *rmq) Shutdown() {
	var err error
	if err = r.rmqChannel.Close(); err != nil {
		log.Println("can't close channel", err)
	}
	if err = r.rmqConn.Close(); err != nil {
		log.Fatalln("can't close connection", err)
	}
	log.Println("shutdown successful")
}

func (r *rmq) getQueueName() string {
	var queueName = r.config.PreferQueueName
	if queueName == "" {
		queueName = r.config.Exchanger + r.config.RoutingKey + "Handler"
	}
	return queueName
}
