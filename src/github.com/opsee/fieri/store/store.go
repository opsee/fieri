package store

import (
	"time"
)

type Store interface {
	PutInstance(*Instance) error
	GetInstance(string, string) (*Instance, error)
	ListInstances(string) ([]*Instance, error)
	ListInstancesByType(string, string) ([]*Instance, error)
	DeleteInstances() error
	PutGroup(*Group) error
	GetGroup(string, string) (*Group, error)
	ListGroups(string) ([]*Group, error)
	ListGroupsByType(string, string) ([]*Group, error)
	DeleteGroups() error
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
