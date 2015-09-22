package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/julienschmidt/httprouter"
	"github.com/opsee/fieri/store"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	requestTimeout = 60 * time.Second
)

var (
	errMissingCustomerId    = errors.New("missing customer id header (Customer-Id).")
	errMalformedRequestBody = errors.New("malformed request body.")
)

type handlerFunc func(ctx context.Context, request interface{}, svc store.Store) (interface{}, int, error)
type decodeFunc func(r *http.Request, p httprouter.Params) (interface{}, error)
type panicFunc func(rw http.ResponseWriter, r *http.Request, data interface{})

type MessageResponse struct {
	Message string `json:"message"`
}

func StartHTTP(addr string, svc store.Store, logger log.Logger) {
	ctx := context.Background()

	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.PanicHandler = makePanicHandler(logger)
	router.OPTIONS("/*any", wrapHandler(ctx, logger, svc, decodeIdentity, okHandler))
	router.GET("/health", wrapHandler(ctx, logger, svc, decodeIdentity, okHandler))
	router.GET("/instances", wrapHandler(ctx, logger, svc, decodeInstanceRequest, instancesHandler))
	router.GET("/instances/:type", wrapHandler(ctx, logger, svc, decodeInstancesRequest, instancesHandler))
	router.GET("/instance/:type/:id", wrapHandler(ctx, logger, svc, decodeInstancesRequest, instanceHandler))
	router.GET("/groups", wrapHandler(ctx, logger, svc, decodeGroupRequest, groupsHandler))
	router.GET("/groups/:type", wrapHandler(ctx, logger, svc, decodeGroupsRequest, groupsHandler))
	router.GET("/group/:type/:id", wrapHandler(ctx, logger, svc, decodeGroupsRequest, groupHandler))
	router.POST("/:type", wrapHandler(ctx, logger, svc, decodeEntityRequest, entityHandler))
	http.ListenAndServe(addr, router)
}

func wrapHandler(ctx context.Context, logger log.Logger, svc store.Store, decoder decodeFunc, handler handlerFunc) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		req, err := decoder(r, params)
		if err != nil {
			renderBadRequest(rw, logger, err)
			return
		}

		response, status, err := handler(ctx, req, svc)
		if err != nil {
			renderServerError(rw, logger, err)
			return
		}

		encodedResponse, err := encodeResponse(response)
		if err != nil {
			renderServerError(rw, logger, err)
			return
		}

		rw.WriteHeader(status)
		rw.Write(encodedResponse)
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
		return nil, errMalformedRequestBody
	}

	return entity, nil
}

func okHandler(ctx context.Context, request interface{}, svc store.Store) (interface{}, int, error) {
	return map[string]bool{"ok": true}, http.StatusOK, nil
}

func instancesHandler(ctx context.Context, request interface{}, svc store.Store) (interface{}, int, error) {
	storeChan := make(chan *store.InstanceResponse)
	go func() {
		response, err := svc.ListInstances(request.(*store.InstancesRequest))
		if err != nil {
			return nil, 0, err
		}
	}()

	

	return response, http.StatusOK, nil
}

func instanceHandler(ctx context.Context, request interface{}, svc store.Store) (interface{}, int, error) {
	response, err := svc.GetInstance(request.(*store.InstanceRequest))
	if err != nil {
		return nil, 0, err
	}

	return response, http.StatusOK, nil
}

func groupsHandler(ctx context.Context, request interface{}, svc store.Store) (interface{}, int, error) {
	response, err := svc.ListGroups(request.(*store.GroupsRequest))
	if err != nil {
		return nil, 0, err
	}

	return response, http.StatusOK, nil
}

func groupHandler(ctx context.Context, request interface{}, svc store.Store) (interface{}, int, error) {
	response, err := svc.GetGroup(request.(*store.GroupRequest))
	if err != nil {
		return nil, 0, err
	}

	return response, http.StatusOK, nil
}

func entityHandler(ctx context.Context, request interface{}, svc store.Store) (interface{}, int, error) {
	response, err := svc.PutEntity(request)
	if err != nil {
		return nil, 0, err
	}

	return response, http.StatusCreated, nil
}

func makePanicHandler(logger log.Logger) panicFunc {
	return func(rw http.ResponseWriter, r *http.Request, data interface{}) {
		renderServerError(rw, logger, data)
	}
}

func renderServerError(rw http.ResponseWriter, logger log.Logger, err interface{}) {
	logger.Log("error", err)
	msg, _ := encodeResponse(MessageResponse{"An unexpected error happened."})
	rw.WriteHeader(http.StatusInternalServerError)
	rw.Write(msg)
}

func renderBadRequest(rw http.ResponseWriter, logger log.Logger, err interface{}) {
	logger.Log("bad-request", err)
	msg, _ := encodeResponse(MessageResponse{fmt.Sprint("Bad request: ", err)})
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write(msg)
}

func encodeResponse(response interface{}) ([]byte, error) {
	return json.Marshal(response)
}
