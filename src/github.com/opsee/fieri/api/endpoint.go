package api

import (
	"golang.org/x/net/context"
	"github.com/go-kit/kit/endpoint"
)

func makeInstanceEndpoint(svc service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		opts := request.(*store.Options)
		instance, err := svc.GetInstance(opts)
		
	}
}

