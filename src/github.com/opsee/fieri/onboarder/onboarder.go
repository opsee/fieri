package onboarder

import (
	// "github.com/opsee/awscan"
	"github.com/satori/go.uuid"
	"sync"
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
}

type OnboardResponse struct {
	RequestId string `json:"request_id"`
}

type onboarder struct {
	*sync.RWMutex
	requests map[string]chan interface{}
}

func NewOnboarder() *onboarder {
	return &onboarder{}
}

func (o *onboarder) Onboard(request *OnboardRequest) *OnboardResponse {
	requestId := uuid.NewV4().String()
	return &OnboardResponse{requestId}
}
