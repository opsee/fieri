package consumer

import (
	"encoding/json"
	"fmt"
	"github.com/nsqio/go-nsq"
	"github.com/opsee/fieri/store"
	"log"
	"time"
)

type Nsq struct {
	consumer *nsq.Consumer
}

type nsqHandler struct {
	logger *log.Logger
	db     store.Store
}

func NewNsq(lookupds []string, db store.Store, logger *log.Logger, concurrency int, topic string) (Consumer, error) {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, Channel, config)
	if err != nil {
		return nil, err
	}

	handler := &nsqHandler{logger: logger, db: db}
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

	switch event.MessageType {
	case "Instance":
		err = h.handleInstance(event, "InstanceId", "ec2")
	case "DBInstance":
		err = h.handleInstance(event, "DBInstanceIdentifier", "rds")
	case "SecurityGroup":
		err = h.handleGroup(event, "GroupId", "security")
	case "DBSecurityGroup":
		err = h.handleGroup(event, "DBSecurityGroupName", "rds-security")
	case "LoadBalancerDescription":
		err = h.handleGroup(event, "LoadBalancerName", "elb")
	}

	return err
}

func (h *nsqHandler) handleInstance(event *Event, identifier, instanceType string) error {
	id, messageBody, err := explodeMessageData(event, identifier)
	if err != nil {
		return err
	}

	fmt.Printf("put instance: %s %s %s %s", id, event.CustomerId, instanceType, string(messageBody))
	instance := store.NewInstance(id, event.CustomerId, instanceType, messageBody)
	return h.db.PutInstance(instance)
}

func (h *nsqHandler) handleGroup(event *Event, identifier, groupType string) error {
	id, messageBody, err := explodeMessageData(event, identifier)
	if err != nil {
		return err
	}

	fmt.Printf("put group: %s %s %s %s", id, event.CustomerId, groupType, string(messageBody))
	group := store.NewGroup(id, event.CustomerId, groupType, messageBody)
	return h.db.PutGroup(group)
}

func explodeMessageData(event *Event, identifier string) (string, []byte, error) {
	messageBody := []byte(event.MessageBody)
	blob := make(map[string]interface{})

	err := json.Unmarshal(messageBody, &blob)
	if err != nil {
		return "", nil, err
	}

	id, ok := blob[identifier]
	if !ok {
		return "", nil, fmt.Errorf("missing %s", identifier)
	}

	return id.(string), messageBody, nil
}

func (h *nsqHandler) log(msgs ...interface{}) {
	if h.logger != nil {
		h.logger.Println(msgs)
	}
}
