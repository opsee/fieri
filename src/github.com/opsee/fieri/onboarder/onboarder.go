package onboarder

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/opsee/awscan"
	"github.com/opsee/fieri/store"
	"github.com/satori/go.uuid"
	"github.com/yeller/yeller-golang"
	"reflect"
)

type Onboarder interface {
	Onboard(*OnboardRequest) *OnboardResponse
}

type OnboardRequest struct {
	AccessKey             string `json:"access_key"`
	SecretKey             string `json:"secret_key"`
	Region                string `json:"region"`
	CustomerId            string `json:"customer_id"`
	UserId                int    `json:"user_id"`
	RequestId             string `json:"request_id"`
	SecurityGroupCount    int    `json:"security_group_count"`
	DBSecurityGroupCount  int    `json:"db_security_group_count"`
	LoadBalancerCount     int    `json:"load_balancer_count"`
	AutoscalingGroupCount int    `json:"autoscaling_group_count"`
	InstanceCount         int    `json:"instance_count"`
	DBInstanceCount       int    `json:"db_instance_count"`
	GroupErrorCount       int    `json:"group_error_count"`
	InstanceErrorCount    int    `json:"instance_error_count"`
	LastError             string `json:"last_error"`
	CheckCount            int    `json:"check_count"`
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
	notifier Notifier
}

const (
	instanceErrorThreshold = 0.3
)

func NewOnboarder(db store.Store, notifier Notifier) *onboarder {
	return &onboarder{
		db:       db,
		notifier: notifier,
	}
}

func (o *onboarder) Onboard(request *OnboardRequest) *OnboardResponse {
	request.RequestId = uuid.NewV4().String()
	go o.scan(request)
	return &OnboardResponse{request.RequestId}
}

func (o *onboarder) scan(request *OnboardRequest) {
	var (
		err       error
		instances = make(map[string]bool)
		logger    = log.WithFields(log.Fields{
			"request-id":  request.RequestId,
			"customer-id": request.CustomerId,
			"region":      request.Region,
			"user-id":     request.UserId,
		})
		disco = awscan.NewDiscoverer(
			awscan.NewScanner(
				&awscan.Config{
					AccessKeyId: request.AccessKey,
					SecretKey:   request.SecretKey,
					Region:      request.Region,
				},
			),
		)
	)

	// just to be safe, get rid of keys out of the onboard request so we don't accidentally log them somewhere
	request.AccessKey = ""
	request.SecretKey = ""

	for event := range disco.Discover() {
		if event.Err != nil {
			switch event.Err.(*awscan.DiscoveryError).Type {
			case awscan.InstanceType, awscan.DBInstanceType:
				request.InstanceErrorCount++
			default:
				request.GroupErrorCount++
			}

			o.handleError(event.Err, request)
		} else {
			messageType := reflect.ValueOf(event.Result).Elem().Type().Name()
			messageBody, err := json.Marshal(event.Result)
			if err != nil {
				o.handleError(err, request)
				continue
			}

			entity, err := store.NewEntity(messageType, request.CustomerId, messageBody)
			if err != nil {
				o.handleError(err, request)
				continue
			}

			ent, err := o.db.PutEntity(entity)
			if err != nil {
				o.handleError(err, request)
				continue
			}

			switch messageType {
			case awscan.InstanceType:
				// we'll have to de-dupe instances so use a ghetto set (map)
				instances[ent.Entity.(*store.Instance).Id] = true
				request.InstanceCount = card(instances)
			case awscan.DBInstanceType:
				request.DBInstanceCount++
			case awscan.SecurityGroupType:
				request.SecurityGroupCount++
			case awscan.DBSecurityGroupType:
				request.DBSecurityGroupCount++
			case awscan.AutoScalingGroupType:
				request.AutoscalingGroupCount++
			case awscan.LoadBalancerType:
				request.LoadBalancerCount++
			}

			logger.WithField("resource-type", messageType).Info("customer resource discovered")
		}
	}

	if request.TooManyErrors() {
		err = o.notifier.NotifyError(request)
		if err != nil {
			o.handleError(err, request)
		}

		o.handleError(fmt.Errorf("too many aws errors"), request)
		return
	}

	err = o.notifier.NotifySuccess(request)
	if err != nil {
		o.handleError(err, request)
	}
}

func (o *onboarder) handleError(err error, request *OnboardRequest) {
	request.LastError = err.Error()
	log.WithField("scan-request-id", request.RequestId).WithError(err).Error("onboarding error")
	yeller.NotifyInfo(err, map[string]interface{}{"onboard_request": request})
}

func (r *OnboardRequest) TooManyErrors() bool {
	total := r.InstanceCount + r.DBInstanceCount
	if total == 0 {
		return r.GroupErrorCount > 0
	}

	return r.GroupErrorCount > 0 ||
		float64(r.InstanceErrorCount)/float64(total) > instanceErrorThreshold
}

func card(m map[string]bool) int {
	i := 0
	for _, _ = range m {
		i++
	}
	return i
}
