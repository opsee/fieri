package consumer

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/nsqio/go-nsq"
	"github.com/opsee/fieri/store"
	"github.com/yeller/yeller-golang"
	"time"
)

type Nsq struct {
	consumer *nsq.Consumer
}

type nsqHandler struct {
	db store.Store
}

func NewNsq(lookupds []string, db store.Store, concurrency int, topic string) (Consumer, error) {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, Channel, config)
	if err != nil {
		return nil, err
	}

	handler := &nsqHandler{db: db}
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
		h.handleError(m, err)
		return nil
	}

	entity, err := store.NewEntity(event.MessageType, event.CustomerId, []byte(event.MessageBody))
	if err != nil {
		h.handleError(m, err)
		return nil
	}

	_, err = h.db.PutEntity(entity)
	if err != nil {
		h.handleError(m, err)
		return nil
	}

	return nil
}

func (h *nsqHandler) handleError(m *nsq.Message, err error) {
	log.WithFields(log.Fields{"err": err.Error(), "message": string(m.Body)}).Warn("error processing nsq message")
	yeller.NotifyInfo(err, map[string]interface{}{"message": string(m.Body)})
}
