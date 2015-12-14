package service

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/opsee/fieri/onboarder"
	"github.com/opsee/fieri/store"
	"github.com/yeller/yeller-golang"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
)

type handlerFunc func(ctx context.Context, request interface{}) (interface{}, int, error)
type decodeFunc func(r *http.Request, p httprouter.Params) (interface{}, error)
type panicFunc func(rw http.ResponseWriter, r *http.Request, data interface{})

func (s *service) StartHTTP(addr string) {
	ctx := context.Background()

	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.PanicHandler = s.makePanicHandler()
	router.OPTIONS("/*any", s.wrapHandler(ctx, decodeIdentity, s.okHandler))
	router.GET("/health", s.wrapHandler(ctx, decodeIdentity, s.okHandler))
	router.GET("/instances", s.wrapHandler(ctx, decodeInstancesRequest, s.instancesHandler))
	router.GET("/instances/:type", s.wrapHandler(ctx, decodeInstancesRequest, s.instancesHandler))
	router.GET("/instance/:type/:id", s.wrapHandler(ctx, decodeInstanceRequest, s.instanceHandler))
	router.GET("/groups", s.wrapHandler(ctx, decodeGroupsRequest, s.groupsHandler))
	router.GET("/groups/:type", s.wrapHandler(ctx, decodeGroupsRequest, s.groupsHandler))
	router.GET("/group/:type/:id", s.wrapHandler(ctx, decodeGroupRequest, s.groupHandler))
	router.POST("/entity/:type", s.wrapHandler(ctx, decodeEntityRequest, s.entityHandler))
	router.POST("/onboard", s.wrapHandler(ctx, decodeOnboardRequest, s.onboardHandler))
	router.GET("/customer", s.wrapHandler(ctx, decodeCustomerRequest, s.customerHandler))
	http.ListenAndServe(addr, router)
}

func (s *service) wrapHandler(ctx context.Context, decoder decodeFunc, handler handlerFunc) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx, cancel := context.WithTimeout(ctx, forwardTimeout)
		defer cancel()

		req, err := decoder(r, params)
		if err != nil {
			s.renderBadRequest(rw, r, err)
			return
		}

		forwardChan := make(chan requestForwarder)
		go func() {
			defer close(forwardChan)
			response, status, err := handler(ctx, req)
			forwardChan <- requestForwarder{response, status, err}
		}()

		select {
		case rf := <-forwardChan:
			if rf.err != nil {
				s.renderServerError(rw, r, rf.err)
				return
			}

			encodedResponse, err := encodeResponse(rf.response)
			if err != nil {
				s.renderServerError(rw, r, err)
				return
			}

			rw.WriteHeader(rf.status)
			rw.Write(encodedResponse)
			log.WithFields(log.Fields{
				"status":      rf.status,
				"path":        r.URL.RequestURI(),
				"method":      r.Method,
				"customer-id": r.Header.Get("Customer-Id"),
			}).Info("http request")

		case <-ctx.Done():
			msg, _ := encodeResponse(MessageResponse{"Backend service unavailable."})
			rw.WriteHeader(http.StatusServiceUnavailable)
			rw.Write(msg)
		}
	}
}

func decodeIdentity(r *http.Request, params httprouter.Params) (interface{}, error) {
	return struct{}{}, nil
}

func decodeInstanceRequest(r *http.Request, params httprouter.Params) (interface{}, error) {
	customerId := r.Header.Get("Customer-Id")
	if customerId == "" {
		return nil, errMissingCustomerId
	}

	return &store.InstanceRequest{
		CustomerId: customerId,
		InstanceId: params.ByName("id"),
		Type:       params.ByName("type"),
	}, nil
}

func decodeInstancesRequest(r *http.Request, params httprouter.Params) (interface{}, error) {
	customerId := r.Header.Get("Customer-Id")
	if customerId == "" {
		return nil, errMissingCustomerId
	}

	return &store.InstancesRequest{
		CustomerId: customerId,
		Type:       params.ByName("type"),
	}, nil
}

func decodeGroupRequest(r *http.Request, params httprouter.Params) (interface{}, error) {
	customerId := r.Header.Get("Customer-Id")
	if customerId == "" {
		return nil, errMissingCustomerId
	}

	return &store.GroupRequest{
		CustomerId: customerId,
		GroupId:    params.ByName("id"),
		Type:       params.ByName("type"),
	}, nil
}

func decodeGroupsRequest(r *http.Request, params httprouter.Params) (interface{}, error) {
	customerId := r.Header.Get("Customer-Id")
	if customerId == "" {
		return nil, errMissingCustomerId
	}

	return &store.GroupsRequest{
		CustomerId: customerId,
		Type:       params.ByName("type"),
	}, nil
}

func decodeEntityRequest(r *http.Request, params httprouter.Params) (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	entity, err := store.NewEntity(params.ByName("type"), r.Header.Get("Customer-Id"), body)
	if err != nil {
		log.WithError(err).WithField("body", string(body)).Error("failed decoding entity")
		return nil, errMalformedRequestBody
	}

	return entity, nil
}

func decodeOnboardRequest(r *http.Request, params httprouter.Params) (interface{}, error) {
	customerId := r.Header.Get("Customer-Id")
	if customerId == "" {
		return nil, errMissingCustomerId
	}

	request := &onboarder.OnboardRequest{}
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(request)
	if err != nil {
		return nil, errMalformedRequestBody
	}

	if request.AccessKey == "" {
		return nil, errMissingAccessKey
	}

	if request.SecretKey == "" {
		return nil, errMissingSecretKey
	}

	if request.Region == "" {
		return nil, errMissingRegion
	}

	if request.UserId == 0 {
		return nil, errMissingUserId
	}

	request.CustomerId = customerId
	return request, nil
}

func decodeCustomerRequest(r *http.Request, params httprouter.Params) (interface{}, error) {
	customerId := r.Header.Get("Customer-Id")
	if customerId == "" {
		return nil, errMissingCustomerId
	}

	request := &store.CustomerRequest{Id: customerId}
	return request, nil
}

func (s *service) okHandler(ctx context.Context, request interface{}) (interface{}, int, error) {
	return map[string]bool{"ok": true}, http.StatusOK, nil
}

func (s *service) instancesHandler(ctx context.Context, request interface{}) (interface{}, int, error) {
	response, err := s.ListInstances(request.(*store.InstancesRequest))
	if err != nil {
		return nil, 0, err
	}

	return response, http.StatusOK, nil
}

func (s *service) instanceHandler(ctx context.Context, request interface{}) (interface{}, int, error) {
	response, err := s.GetInstance(request.(*store.InstanceRequest))
	if err != nil {
		return nil, 0, err
	}

	return response, http.StatusOK, nil
}

func (s *service) groupsHandler(ctx context.Context, request interface{}) (interface{}, int, error) {
	response, err := s.ListGroups(request.(*store.GroupsRequest))
	if err != nil {
		return nil, 0, err
	}

	return response, http.StatusOK, nil
}

func (s *service) groupHandler(ctx context.Context, request interface{}) (interface{}, int, error) {
	response, err := s.GetGroup(request.(*store.GroupRequest))
	if err != nil {
		return nil, 0, err
	}

	return response, http.StatusOK, nil
}

func (s *service) entityHandler(ctx context.Context, request interface{}) (interface{}, int, error) {
	response, err := s.PutEntity(request)
	if err != nil {
		return nil, 0, err
	}

	return response, http.StatusCreated, nil
}

func (s *service) onboardHandler(ctx context.Context, request interface{}) (interface{}, int, error) {
	response := s.Onboard(request.(*onboarder.OnboardRequest))
	return response, http.StatusCreated, nil
}

func (s *service) customerHandler(ctx context.Context, request interface{}) (interface{}, int, error) {
	response, err := s.GetCustomer(request.(*store.CustomerRequest))
	if err != nil {
		return MessageResponse{"No customer exists."}, http.StatusOK, nil
	}

	return response, http.StatusOK, nil
}

func (s *service) makePanicHandler() panicFunc {
	return func(rw http.ResponseWriter, r *http.Request, data interface{}) {
		yeller.NotifyPanic(data)
		s.renderServerError(rw, r, fmt.Errorf("panic data: %#v", data))
	}
}

func (s *service) renderServerError(rw http.ResponseWriter, r *http.Request, err error) {
	log.WithError(err).WithField("request", *r).Error("internal server error")
	msg, _ := encodeResponse(MessageResponse{"An unexpected error happened."})
	rw.WriteHeader(http.StatusInternalServerError)
	rw.Write(msg)
}

func (s *service) renderBadRequest(rw http.ResponseWriter, r *http.Request, err error) {
	log.WithError(err).WithField("request", *r).Error("bad request")
	msg, _ := encodeResponse(MessageResponse{fmt.Sprint("Bad request: ", err)})
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write(msg)
}

func encodeResponse(response interface{}) ([]byte, error) {
	return json.Marshal(response)
}
