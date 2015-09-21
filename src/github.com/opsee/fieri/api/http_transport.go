package api

import (
	"net/http"
	"golang.org/x/net/context"
	"github.com/julienschmidt/httprouter"
	"github.com/go-kit/kit/log"
	"github.com/opsee/fieri/store"
)

const (
	resourceKey = iota
	idKey
	customerIdKey
	typeKey
)

const (
	requestTimeout = 60 * time.Second
	errorInternalServerError = []byte(`{"message": "an unexpected error happened!"}`)
)

var (
	origins       = []string{
		"http://localhost:8080",
		"https://staging.opsy.co",
		"https://app.opsee.com",
	}
)

type handlerFunc func(ctx context.Context) (interface{}, int, error)

func StartHTTP(addr string, db store.Store, logger log.Logger) {
	ctx := context.Background()
	appHandler := makeHandler(ctx, logger, db, resourceHandler)

	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.PanicHandler = makePanicHandler(logger)
	router.GET("/:resource", appHandler)
	router.GET("/:resource/:type", appHandler)
	router.GET("/:resource/:type/:id", appHandler)

	// idunno yet
	router.POST("/entity", ----)
	http.ListenAndServe(addr, router)
}

func makeHandler(ctx context.Context, handler handlerFunc) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		customerHeader := r.Header.Get("Customer-Id")
		ctx = context.WithValue(ctx, customerIdKey, customerHeader)
		ctx = context.WithValue(ctx, resourceKey, params.ByName("resource"))
		ctx = context.WithValue(ctx, idKey, params.ByName("id"))
		ctx = context.WithValue(ctx, typeKey, params.ByName("type"))

		response, status, err := handlerFunc(ctx)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(errorInternalServerError)
			return
		}

		encodedResponse, err := encodeResponse(response)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(errorInternalServerError)
			return
		}

		rw.WriteHeader(status)
		rw.Write(encodedResponse)
	}
}

func resourceHandler(ctx context.Context) (interface{}, int, error) {
	var (
		response interface{}
		err error
		status = http.StatusOK
	)

	opts := &store.Options{
		CustomerId: ctx.Value(customerIdKey).(string),
		Type: ctx.Value(typeKey).(string),
	}

	if opts.CustomerId == "" {
		return badRequestHandler(ctx)
	}

	switch ctx.Value(resourceKey) {
	case "instances":
		instances, err = store.ListInstances(opts)
		response = ListInstancesResponse{instances}

	case "groups":
		groups, err = store.ListGroups(opts)
		response = ListGroupsResponse{groups}

	case "instance":
		opts.InstanceId = ctx.Value(idKey).(string)
		if opts.InstanceId == "" {
			return badRequestHandler(ctx)
		}

		instance, err = store.GetInstance(opts)
		response = GetInstanceResponse{instance}

	case "group":
		opts.GroupId = ctx.Value(idKey).(string)
		if opts.GroupId == "" {
			return badRequestHandler(ctx)
		}

		group, _ := store.GetGroup(opts)
		instances, err := store.ListInstances(opts)
		response = GetGroupResponse{group, instances}

	default:
		return notFoundHandler(ctx)
	}

	return response, status, err
}

func notFoundHandler(ctx context.Context) (interface{}, int, error) {
	return MessageResponse{"not found"}, http.StatusNotFound, nil
}

func badRequestHandler(ctx context.Context) (interface{}, int, error) {
	return MessageResponse{"bad request"}, http.StatusBadRequest, nil
}

func makePanicHandler(logger log.Logger) {
	return func(rw http.ResponseWriter, r *http.Request, data interface{}) {
		logger.Log("panic", data)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(errorInternalServerError)
	}
}

func encodeResponse(response interface{}) ([]byte, error) {
	return json.Marshal(response)
}
