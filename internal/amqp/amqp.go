package amqp

import (
	"github.com/streadway/amqp"
	"strings"
)

type (
	// AMQP interpret (implement) AMQP interface definition
	AMQP struct {
		conn            *amqp.Connection
		ch              *amqp.Channel
		passiveExchange bool
	}
)

// New returns a new AMQP configured, or returning an non-nil err
// if an error occurred while creating connection or channel.
func New(dsn string, pex bool) (*AMQP, error) {
	conn, ch, err := createConnAndChan(dsn)
	if err != nil {
		return nil, err
	}

	return &AMQP{
		conn:            conn,
		ch:              ch,
		passiveExchange: pex,
	}, nil
}

// Publish wraps amqp.Publish method
func (ai *AMQP) Publish(exchange, key string, opts amqp.Publishing) error {
	return ai.ch.Publish(exchange, key, false, false, opts)
}

// CreateConsumer creates a amqp consumer. Most interesting declare args are:
func (ai *AMQP) CreateConsumer(exchange, key, kind, queue string, durable bool, declareArgs, bindArgs amqp.Table) (<-chan amqp.Delivery, error) {
	if err := ai.WithExchange(exchange, kind, durable); err != nil {
		return nil, err
	}

	var q amqp.Queue
	var err error
	if strings.HasPrefix(queue, "rabbus.gen-") {
		q, err = ai.ch.QueueDeclare(queue, durable, true, true, false, declareArgs)
	} else {
		q, err = ai.ch.QueueDeclare(queue, durable, false, false, false, declareArgs)
	}

	if err != nil {
		return nil, err
	}

	if err := ai.ch.QueueBind(q.Name, key, exchange, false, bindArgs); err != nil {
		return nil, err
	}

	return ai.ch.Consume(q.Name, "", false, false, false, false, nil)
}

// WithExchange creates a amqp exchange
func (ai *AMQP) WithExchange(exchange, kind string, durable bool) error {
	if ai.passiveExchange {
		return ai.ch.ExchangeDeclarePassive(exchange, kind, durable, false, false, false, nil)
	}

	return ai.ch.ExchangeDeclare(exchange, kind, durable, false, false, false, nil)
}

// WithQos wrapper over amqp.Qos method
func (ai *AMQP) WithQos(count, size int, global bool) error {
	return ai.ch.Qos(count, size, global)
}

// NotifyClose wrapper over notifyClose method
func (ai *AMQP) NotifyClose(c chan *amqp.Error) chan *amqp.Error {
	return ai.conn.NotifyClose(c)
}

// Close closes the running amqp connection and channel
func (ai *AMQP) Close() error {
	if err := ai.ch.Close(); err != nil {
		return err
	}

	if ai.conn != nil {
		return ai.conn.Close()
	}

	return nil
}

func createConnAndChan(dsn string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	return conn, ch, nil
}
