package onboarder

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/nsqio/go-nsq"
	"github.com/opsee/awscan"
	"github.com/opsee/fieri/store"
	"github.com/satori/go.uuid"
	"reflect"
)

type Onboarder interface {
	Onboard(*OnboardRequest) *OnboardResponse
}

type OnboardRequest struct {
	CustomerId string `json:"customer_id"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	Region     string `json:"region"`
	Email      string `json:"email"`
	RequestId  string `json:"request_id"`
}

type Event struct {
	RequestId  string `json:"request_id"`
	CustomerId string `json:"customer_id"`
	EventType  string `json:"event_type"`
	Error      string `json:"error"`
}

type OnboardResponse struct {
	RequestId string `json:"request_id"`
}

type onboarder struct {
	db       store.Store
	producer *nsq.Producer
	logger   log.Logger
	topic    string
}

func NewOnboarder(db store.Store, producer *nsq.Producer, logger log.Logger, topic string) *onboarder {
	return &onboarder{
		db:       db,
		producer: producer,
		logger:   log.NewContext(logger).With("onboarding", true),
		topic:    topic,
	}
}

func (o *onboarder) Onboard(request *OnboardRequest) *OnboardResponse {
	request.RequestId = uuid.NewV4().String()
	go o.scan(request)
	return &OnboardResponse{request.RequestId}
}

func (o *onboarder) scan(request *OnboardRequest) {
	logger := log.NewContext(o.logger).With("scan-request-id", request.RequestId)
	topic := fmt.Sprintf("%s.%s", request.CustomerId, o.topic)

	disco := awscan.NewDiscoverer(
		awscan.NewScanner(
			&awscan.Config{
				AccessKeyId: request.AccessKey,
				SecretKey:   request.SecretKey,
				Region:      request.Region,
			},
		),
	)

	for event := range disco.Discover() {
		outEvent := &Event{
			CustomerId: request.CustomerId,
			RequestId:  request.RequestId,
		}

		if event.Err != nil {
			logger.Log("error", event.Err)
			outEvent.Error = event.Err.Error()
		} else {
			// FIXME: this is a hack because there is no conversion of aws obj -> store.Entity, only marshal/unmarshal
			messageType := reflect.ValueOf(event.Result).Elem().Type().Name()
			messageBody, err := json.Marshal(event.Result)
			if err != nil {
				logger.Log("error", err)
				outEvent.Error = err.Error()
				goto publish
			}

			entity, err := store.NewEntity(messageType, request.CustomerId, messageBody)
			if err != nil {
				logger.Log("error", err)
				outEvent.Error = err.Error()
				goto publish
			}

			ent, err := o.db.PutEntity(entity)
			if err != nil {
				logger.Log("error", err)
				outEvent.Error = err.Error()
				goto publish
			}

			outEvent.EventType = ent.Type
		}

	publish:
		msg, err := json.Marshal(outEvent)
		if err != nil {
			logger.Log("error", err)
			continue
		}

		o.producer.Publish(topic, msg)
		logger.Log("resource-type", outEvent.EventType)
	}

	// alright, we done
	doneEvent := &Event{
		CustomerId: request.CustomerId,
		RequestId:  request.RequestId,
		EventType:  "Done",
	}
	doneMsg, _ := json.Marshal(doneEvent)
	o.producer.Publish(topic, doneMsg)
	logger.Log("done", true)
}
