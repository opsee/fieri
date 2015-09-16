package store

import (
	"errors"
	"time"
)

type Store interface {
	PutInstance(*Instance) error
	GetInstance(*Options) (*Instance, error)
	ListInstances(*Options) ([]*Instance, error)
	CountInstances(*Options) (int, error)
	DeleteInstances() error
	PutGroup(*Group) error
	GetGroup(*Options) (*Group, error)
	ListGroups(*Options) ([]*Group, error)
	CountGroups(*Options) (int, error)
	DeleteGroups() error
}

type Options struct {
	CustomerId string `json:"customer_id"`
	InstanceId string `json:"instance_id"`
	GroupId    string `json:"group_id"`
	Type       string `json:"type"`
}

type Instance struct {
	Id         string    `json:"id"`
	CustomerId string    `json:"customer_id" db:"customer_id"`
	Type       string    `json:"type"`
	Data       []byte    `json:"data"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type Group struct {
	Name       string    `json:"name"`
	CustomerId string    `json:"customer_id" db:"customer_id"`
	Type       string    `json:"type"`
	Data       []byte    `json:"data"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

var (
	ErrMissingInstanceId = errors.New("must provide instance id")
	ErrMissingGroupId    = errors.New("must provide group id")
	ErrMissingCustomerId = errors.New("must provide customer id")
)

func NewInstance(id, customerId, instanceType string, data []byte) *Instance {
	return &Instance{
		Id:         id,
		CustomerId: customerId,
		Type:       instanceType,
		Data:       data,
	}
}

func NewGroup(name, customerId, groupType string, data []byte) *Group {
	return &Group{
		Name:       name,
		CustomerId: customerId,
		Type:       groupType,
		Data:       data,
	}
}
