package consumer

import (
	"encoding/json"
	"fmt"
	kvlog "github.com/go-kit/kit/log"
	"github.com/nsqio/go-nsq"
	"github.com/opsee/fieri/store"
	"time"
)

type Nsq struct {
	consumer *nsq.Consumer
}

type nsqHandler struct {
	logger kvlog.Logger
	db     store.Store
}

func NewNsq(lookupds []string, db store.Store, kvlogger kvlog.Logger, concurrency int, topic string) (Consumer, error) {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, Channel, config)
	if err != nil {
		return nil, err
	}

	if kvlogger == nil {
		kvlogger = kvlog.NewNopLogger()
	}

	handler := &nsqHandler{logger: kvlogger, db: db}
	consumer.AddConcurrentHandlers(handler, concurrency)
	consumer.ConnectToNSQLookupds(lookupds)

	return &Nsq{consumer: consumer}, nil
}

func (c *Nsq) Stop() error {
	c.consumer.Stop()

	var err error

	select {
	case <-c.consumer.StopChan:
		err = nil
	case <-time.After(5 * time.Second):
		err = fmt.Errorf("timed out waiting for consumer shutdown")
	}

	return err
}

func (h *nsqHandler) HandleMessage(m *nsq.Message) error {
	event := &Event{}
	err := json.Unmarshal(m.Body, event)
	if err != nil {
		return err
	}

	entity, err := store.NewEntity(event.MessageType, event.CustomerId, event.MessageBody)
	if err != nil {
		return err
	}

	return h.db.PutEntity(entity)
}
