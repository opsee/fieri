package api

import (
	"github.com/opsee/fieri/store"
)

type Request struct {
	CustomerId string `json:"customer_id"`
	Type string `json:"type"`
}

type GetInstanceRequest struct {
	Request
	InstanceId string `json:"instance_id"`
}

type ListInstancesRequest struct {
	Request
}

type GetGroupRequest struct {
	Request
	GroupId string `json:"group_id"`
}

type ListGroupsRequest struct {
	Request
}

type GetInstanceResponse struct {
	*store.Instance
}

type ListInstancesResponse struct {
	Instances []*store.Instance `json:"instances"`
}

type GetGroupResponse struct {
	Group *store.Group `json:"group"`
	Instances []*store.Instance `json:"instances"`
}

type ListGroupsResponse struct {
	[]*store.Group `json:"group"`
}

type MessageResponse struct {
	Message string
}

type PutEntityRequest struct {
	interface{}
}

type Service interface {
	GetInstance(*GetInstanceRequest) (*GetInstanceResponse, error)
	ListInstances(*ListInstancesRequest) (*ListInstancesResponse, error)
	GetGroup(*GetGroupRequest) (*GetGroupResponse, error)
	GetGroups(*GetGroupsRequest) (*GetGroupsResponse, error)
	PutEntity(*PutEntityRequest) error
}

type service struct {}

func (svc service) GetInstance(req *GetInstanceRequest) (*GetInstanceResponse, error) {
	
}

func (svc service) ListInstances(req *ListInstancesRequest) (*ListInstancesResponse, error) {
	
}

func (svc service) GetGroup(req *GetGroupRequest) (*GetGroupResponse, error) {
	
}

func (svc service) GetGroups(req *GetGroupsRequest) (*GetGroupsResponse, error) {
	
}


func (svc service) PutEntity(ent *PutEntityRequest) error {
	return store.PutEntity(ent)
}
