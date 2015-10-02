package onboarder

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
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
	AccessKey  string         `json:"access_key"`
	SecretKey  string         `json:"secret_key"`
	Region     string         `json:"region"`
	CustomerId string         `json:"customer_id"`
	UserId     int            `json:"user_id"`
	RequestId  string         `json:"request_id"`
	Result     *OnboardResult `json:"result"`
}

type OnboardResult struct {
	SecurityGroupCount    int `json:"security_group_count"`
	DBSecurityGroupCount  int `json:"db_security_group_count"`
	LoadBalancerCount     int `json:"load_balancer_count"`
	AutoscalingGroupCount int `json:"autoscaling_group_count"`
	InstanceCount         int `json:"instance_count"`
	DBInstanceCount       int `json:"db_instance_count"`
	GroupErrorCount       int `json:"group_error_count"`
	InstanceErrorCount    int `json:"instance_error_count"`
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
	logger   log.Logger
	notifier Notifier
}

const (
	discoveryTemplate      = "discovery-completion"
	errorTemplate          = "discovery-failure"
	instanceErrorThreshold = 0.3
)

func NewOnboarder(db store.Store, logger log.Logger, notifier Notifier) *onboarder {
	return &onboarder{
		db:       db,
		logger:   log.NewContext(logger).With("onboarding", true),
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
		instances = make(map[string]bool)
		logger    = log.NewContext(o.logger).With("scan-request-id", request.RequestId)
		disco     = awscan.NewDiscoverer(
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

	// set up a result struct
	request.Result = &OnboardResult{}

	for event := range disco.Discover() {
		if event.Err != nil {
			switch event.Err.(*awscan.DiscoveryError).Type {
			case awscan.InstanceType, awscan.DBInstanceType:
				request.Result.InstanceErrorCount++
			default:
				request.Result.GroupErrorCount++
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
				request.Result.InstanceCount = card(instances)
			case awscan.DBInstanceType:
				request.Result.DBInstanceCount++
			case awscan.SecurityGroupType:
				request.Result.SecurityGroupCount++
			case awscan.DBSecurityGroupType:
				request.Result.DBSecurityGroupCount++
			case awscan.AutoScalingGroupType:
				request.Result.AutoscalingGroupCount++
			case awscan.LoadBalancerType:
				request.Result.LoadBalancerCount++
			}

			logger.Log("resource-type", messageType)
		}
	}

	if request.TooManyErrors() {
		o.notifier.NotifyEmail(request.UserId, errorTemplate, map[string]interface{}{})
		o.handleError(fmt.Errorf("too many aws errors"), request)
		return
	}

	emailVars := map[string]interface{}{
		"aws_region":              request.Region,
		"instance_count":          request.Result.InstanceCount,
		"db_instance_count":       request.Result.DBInstanceCount,
		"security_group_count":    request.Result.SecurityGroupCount,
		"db_security_group_count": request.Result.DBSecurityGroupCount,
		"load_balancer_count":     request.Result.LoadBalancerCount,
		"autoscaling_group_count": request.Result.AutoscalingGroupCount,
		"checks_count":            0,
	}

	_, err := o.notifier.NotifyEmail(request.UserId, discoveryTemplate, emailVars)
	if err != nil {
		o.handleError(err, request)
	}

	err = o.notifier.NotifySlack(emailVars)
	if err != nil {
		o.handleError(err, request)
	}
}

func (o *onboarder) handleError(err error, request *OnboardRequest) {
	log.NewContext(o.logger).With("scan-request-id", request.RequestId).Log("error", err.Error())
	yeller.NotifyInfo(err, map[string]interface{}{"onboard_request": request})
}

func (r *OnboardRequest) TooManyErrors() bool {
	return r.Result.GroupErrorCount > 0 ||
		float64(r.Result.InstanceErrorCount)/float64(r.Result.InstanceCount+r.Result.DBInstanceCount) > instanceErrorThreshold
}

func card(m map[string]bool) int {
	i := 0
	for _, _ = range m {
		i++
	}
	return i
}
