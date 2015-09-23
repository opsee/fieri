package service

import (
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/opsee/fieri/onboarder"
	"github.com/opsee/fieri/store"
	"time"
)

type Service interface {
	store.Store
	onboarder.Onboarder
}

type service struct {
	store.Store
	onboarder.Onboarder
	logger log.Logger
}

type MessageResponse struct {
	Message string `json:"message"`
}

type requestForwarder struct {
	response interface{}
	status   int
	err      error
}

const (
	forwardTimeout = 5 * time.Second
)

var (
	errMissingCustomerId    = errors.New("missing customer id header (Customer-Id).")
	errMalformedRequestBody = errors.New("malformed request body.")
	errMissingAccessKey     = errors.New("missing access_key.")
	errMissingSecretKey     = errors.New("missing secret_key.")
	errMissingEmail         = errors.New("missing email.")
)

func NewService(store store.Store, onboarder onboarder.Onboarder, logger log.Logger) *service {
	return &service{store, onboarder, logger}
}
