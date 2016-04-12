package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	opsee_aws "github.com/opsee/basic/schema/aws"
	opsee_aws_autoscaling "github.com/opsee/basic/schema/aws/autoscaling"
	opsee_aws_ec2 "github.com/opsee/basic/schema/aws/ec2"
	opsee_aws_elb "github.com/opsee/basic/schema/aws/elb"
	opsee_aws_rds "github.com/opsee/basic/schema/aws/rds"
	"time"
)

type Store interface {
	Start()
	PutEntity(interface{}) (*EntityResponse, error)
	GetInstance(*InstanceRequest) (*InstanceResponse, error)
	ListInstances(*InstancesRequest) (*InstancesResponse, error)
	CountInstances(*InstancesRequest) (*CountResponse, error)
	GetGroup(*GroupRequest) (*GroupResponse, error)
	GetCustomer(*CustomerRequest) (*CustomerResponse, error)
	ListGroups(*GroupsRequest) (*GroupsResponse, error)
	CountGroups(*GroupsRequest) (*CountResponse, error)
}

type InstanceRequest struct {
	CustomerId string `json:"customer_id"`
	InstanceId string `json:"instance_id"`
	Type       string `json:"type"`
}

type InstancesRequest struct {
	CustomerId string `json:"customer_id"`
	GroupId    string `json:"group_id"`
	Type       string `json:"type"`
}

type GroupRequest struct {
	CustomerId string `json:"customer_id"`
	GroupId    string `json:"group_id"`
	Type       string `json:"type"`
}

type GroupsRequest struct {
	CustomerId string `json:"customer_id"`
	Type       string `json:"type"`
}

type InstanceResponse struct {
	Instance *Instance `json:"instance"`
}

type InstancesResponse struct {
	Instances []*InstanceResponse `json:"instances"`
}

type GroupResponse struct {
	Group         *Group              `json:"group"`
	Instances     []*InstanceResponse `json:"instances,omitempty"`
	InstanceCount int                 `json:"instance_count"`
}

type GroupsResponse struct {
	Groups []*GroupResponse `json:"groups"`
}

type CustomerRequest struct {
	Id string `json:"id"`
}

type CustomerResponse struct {
	Customer *Customer `json:"customer"`
}

type EntityResponse struct {
	Entity interface{} `json:"entity"`
}

type CountResponse struct {
	Count int `json:"count"`
}

type Customer struct {
	Id        string    `json:"id"`
	LastSync  time.Time `json:"last_sync" db:"last_sync"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Instance struct {
	Id         string    `json:"id"`
	CustomerId string    `json:"customer_id" db:"customer_id"`
	Type       string    `json:"type"`
	Data       []byte    `json:"data"`
	Groups     []*Group  `json:"-" db:""`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type Group struct {
	Name          string      `json:"name"`
	CustomerId    string      `json:"customer_id" db:"customer_id"`
	Type          string      `json:"type"`
	Data          []byte      `json:"data"`
	InstanceCount int         `json:"instance_count" db:"instance_count"`
	Instances     []*Instance `json:"-" db:""`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
}

type RouteTable struct {
	Id         string    `json:"id"`
	CustomerId string    `json:"customer_id" db:"customer_id"`
	Data       []byte    `json:"data"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type Subnet struct {
	Id         string    `json:"id"`
	CustomerId string    `json:"customer_id" db:"customer_id"`
	Data       []byte    `json:"data"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

const (
	InstanceEntityType         = "Instance"
	DBInstanceEntityType       = "DBInstance"
	SecurityGroupEntityType    = "SecurityGroup"
	DBSecurityGroupEntityType  = "DBSecurityGroup"
	AutoScalingGroupEntityType = "Group"
	ELBEntityType              = "LoadBalancerDescription"
	RouteTableEntityType       = "RouteTable"
	SubnetEntityType           = "Subnet"

	InstanceStoreType         = "ec2"
	DBInstanceStoreType       = "rds"
	SecurityGroupStoreType    = "security"
	DBSecurityGroupStoreType  = "rds-security"
	AutoScalingGroupStoreType = "autoscaling"
	ELBStoreType              = "elb"
)

var (
	ErrMissingInstanceId   = errors.New("must provide instance id")
	ErrMissingGroupId      = errors.New("must provide group id")
	ErrMissingRouteTableId = errors.New("must provide route table id")
	ErrMissingSubnetId     = errors.New("must provide subnet id")
	ErrMissingCustomerId   = errors.New("must provide customer id")
	ErrMissingType         = errors.New("must provide type")
	ErrMissingBody         = errors.New("must provide body")
)

func NewEntity(entityType, customerId string, blob []byte) (interface{}, error) {
	var (
		err    error
		entity interface{}
	)

	switch entityType {
	case InstanceEntityType:
		instanceData := &opsee_aws_ec2.Instance{}
		if err = json.Unmarshal(blob, instanceData); err != nil {
			break
		}
		entity, err = NewInstance(customerId, instanceData)

	case DBInstanceEntityType:
		dbInstanceData := &opsee_aws_rds.DBInstance{}
		if err = json.Unmarshal(blob, dbInstanceData); err != nil {
			break
		}
		entity, err = NewInstance(customerId, dbInstanceData)

	case SecurityGroupEntityType:
		secGroupData := &opsee_aws_ec2.SecurityGroup{}
		if err = json.Unmarshal(blob, secGroupData); err != nil {
			break
		}
		entity, err = NewGroup(customerId, secGroupData)

	case ELBEntityType:
		elbData := &opsee_aws_elb.LoadBalancerDescription{}
		if err = json.Unmarshal(blob, elbData); err != nil {
			break
		}
		entity, err = NewGroup(customerId, elbData)

	case AutoScalingGroupEntityType:
		autoscalingData := &opsee_aws_autoscaling.Group{}
		if err = json.Unmarshal(blob, autoscalingData); err != nil {
			break
		}
		entity, err = NewGroup(customerId, autoscalingData)

	case RouteTableEntityType:
		routeTableData := &opsee_aws_ec2.RouteTable{}
		if err = json.Unmarshal(blob, routeTableData); err != nil {
			break
		}
		entity, err = NewRouteTable(customerId, routeTableData)

	case SubnetEntityType:
		subnetData := &opsee_aws_ec2.Subnet{}
		if err = json.Unmarshal(blob, subnetData); err != nil {
			break
		}
		entity, err = NewSubnet(customerId, subnetData)
	}

	return entity, err
}

func NewInstance(customerId string, instanceData interface{}) (*Instance, error) {
	var (
		instance *Instance
		groups   []*Group
		jsonD    []byte
		err      error
	)

	switch t := instanceData.(type) {
	case *opsee_aws_ec2.Instance:
		if t.InstanceId == nil {
			return nil, ErrMissingInstanceId
		}

		groups = make([]*Group, 0, len(t.SecurityGroups))

		for _, group := range t.SecurityGroups {
			gr := &opsee_aws_ec2.SecurityGroup{}
			opsee_aws.CopyInto(gr, group)

			g, err := NewGroup(customerId, gr)
			if err != nil {
				continue
			}
			groups = append(groups, g)
		}

		jsonD, err = json.Marshal(t)
		instance = &Instance{
			Id:         aws.StringValue(t.InstanceId),
			CustomerId: customerId,
			Type:       InstanceStoreType,
			Groups:     groups,
			Data:       jsonD,
		}

	case *opsee_aws_rds.DBInstance:
		if t.DBInstanceIdentifier == nil {
			return nil, ErrMissingInstanceId
		}

		groups = make([]*Group, 0, len(t.VpcSecurityGroups))

		for _, group := range t.VpcSecurityGroups {
			gr := &opsee_aws_ec2.SecurityGroup{}
			opsee_aws.CopyInto(gr, group)

			g, err := NewGroup(customerId, gr)
			if err != nil {
				continue
			}

			groups = append(groups, g)
		}

		jsonD, err = json.Marshal(t)
		instance = &Instance{
			Id:         aws.StringValue(t.DBInstanceIdentifier),
			CustomerId: customerId,
			Type:       DBInstanceStoreType,
			Groups:     groups,
			Data:       jsonD,
		}
	default:
		err = fmt.Errorf("unsupported instance type: %#v", t)
	}

	if err != nil {
		return nil, err
	}

	return instance, nil
}

func NewGroup(customerId string, groupData interface{}) (*Group, error) {
	var (
		group *Group
		jsonD []byte
		err   error
	)

	switch t := groupData.(type) {
	case *opsee_aws_ec2.SecurityGroup:
		if t.GroupId == nil {
			return nil, ErrMissingGroupId
		}

		jsonD, err = json.Marshal(t)
		group = &Group{
			CustomerId: customerId,
			Name:       aws.StringValue(t.GroupId),
			Type:       SecurityGroupStoreType,
			Data:       jsonD,
		}

	case *opsee_aws_elb.LoadBalancerDescription:
		if t.LoadBalancerName == nil {
			return nil, ErrMissingGroupId
		}

		instances := make([]*Instance, 0, len(t.Instances))
		for _, instance := range t.Instances {
			inst := &opsee_aws_ec2.Instance{}
			opsee_aws.CopyInto(inst, instance)

			ii, err := NewInstance(customerId, inst)
			if err != nil {
				continue
			}

			instances = append(instances, ii)
		}

		jsonD, err = json.Marshal(t)
		group = &Group{
			CustomerId: customerId,
			Name:       aws.StringValue(t.LoadBalancerName),
			Type:       ELBStoreType,
			Data:       jsonD,
			Instances:  instances,
		}

	case *opsee_aws_autoscaling.Group:
		if t.AutoScalingGroupName == nil {
			return nil, ErrMissingGroupId
		}

		instances := make([]*Instance, 0, len(t.Instances))
		for _, instance := range t.Instances {
			inst := &opsee_aws_ec2.Instance{}
			opsee_aws.CopyInto(inst, instance)

			ii, err := NewInstance(customerId, inst)
			if err != nil {
				continue
			}

			instances = append(instances, ii)
		}

		jsonD, err = json.Marshal(t)
		group = &Group{
			CustomerId: customerId,
			Name:       aws.StringValue(t.AutoScalingGroupName),
			Type:       AutoScalingGroupStoreType,
			Data:       jsonD,
			Instances:  instances,
		}
	default:
		err = fmt.Errorf("unsupported group type: %#v", t)
	}

	if err != nil {
		return nil, err
	}

	return group, nil
}

func NewRouteTable(customerId string, routeTableData *opsee_aws_ec2.RouteTable) (*RouteTable, error) {
	jsonD, err := json.Marshal(routeTableData)

	if err != nil {
		return nil, err
	}

	return &RouteTable{
		Id:         aws.StringValue(routeTableData.RouteTableId),
		CustomerId: customerId,
		Data:       jsonD,
	}, nil
}

func NewSubnet(customerId string, subnetData *opsee_aws_ec2.Subnet) (*Subnet, error) {
	jsonD, err := json.Marshal(subnetData)

	if err != nil {
		return nil, err
	}

	return &Subnet{
		Id:         aws.StringValue(subnetData.SubnetId),
		CustomerId: customerId,
		Data:       jsonD,
	}, nil
}

func (i *Instance) MarshalJSON() ([]byte, error) {
	return i.Data, nil
}

func (g *Group) MarshalJSON() ([]byte, error) {
	return g.Data, nil
}
