package service

import (
	"errors"
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
	errMissingRegion        = errors.New("missing region.")
	errMissingEmail         = errors.New("missing email.")
	errMissingRequestId     = errors.New("missing request_id.")
	errMissingUserId        = errors.New("missing user_id.")
)

func NewService(store store.Store, onboarder onboarder.Onboarder) *service {
	return &service{store, onboarder}
}
