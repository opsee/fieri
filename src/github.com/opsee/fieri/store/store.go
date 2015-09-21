package store

import (
	"golang.org/x/net/context"
	"encoding/json"
	"errors"
	"time"
)

type Store interface {
	PutEntity(interface{}) error
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
	Groups     []*Group  `json:"-" db:""`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type Group struct {
	Name       string      `json:"name"`
	CustomerId string      `json:"customer_id" db:"customer_id"`
	Type       string      `json:"type"`
	Data       []byte      `json:"data"`
	Instances  []*Instance `json:"-" db:""`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

type InstanceData struct {
	Monitoring         map[string]interface{}
	PublicDnsName      string
	LaunchTime         time.Time
	PublicIpAddress    string
	PrivateIpAddress   string
	InstanceId         string
	ImageId            string
	PrivateDnsName     string
	KeyName            string
	SecurityGroups     []map[string]interface{}
	SubnetId           string
	InstanceType       string
	Placement          map[string]interface{}
	IamInstanceProfile map[string]interface{}
	Tags               []map[string]interface{}
	AmiLaunchIndex     int
}

type DBInstanceData struct {
	PubliclyAccessible               bool
	VpcSecurityGroups                []map[string]interface{}
	InstanceCreateTime               time.Time
	OptionGroupMemberships           []map[string]interface{}
	Engine                           string
	MultiAZ                          bool
	LatestRestorableTime             time.Time
	DBSecurityGroups                 []map[string]interface{}
	DBParameterGroups                []map[string]interface{}
	DBSubnetGroup                    map[string]interface{}
	SecondaryAvailabilityZone        string
	ReadReplicaDBInstanceIdentifiers []string
	AllocatedStorage                 int
	DBName                           string
	Endpoint                         map[string]interface{}
	DBInstanceStatus                 string
	EngineVersion                    string
	AvailabilityZone                 string
	StorageType                      string
	DbiResourceId                    string
	CACertificateIdentifier          string
	Iops                             int
	StorageEncrypted                 bool
	DBInstanceClass                  string
	DbInstancePort                   int
	DBInstanceIdentifier             string
}

type SecurityGroupData struct {
	Description string
	GroupName   string
	GroupId     string
}

type DBSecurityGroupData struct {
	OwnerId                    string
	DBSecurityGroupDescription string
	DBSecurityGroupName        string
	EC2SecurityGroups          []map[string]interface{}
}

type ELBData struct {
	Subnets                   []string
	CanonicalHostedZoneNameID string
	CanonicalHostedZoneName   string
	ListenerDescriptions      []map[string]interface{}
	HealthCheck               map[string]interface{}
	VPCId                     string
	Instances                 []map[string]interface{}
	DNSName                   string
	SecurityGroups            []string
	LoadBalancerName          string
	CreatedTime               time.Time
	AvailabilityZones         []string
	Scheme                    string
	SourceSecurityGroup       map[string]interface{}
}

const (
	InstanceEntityType        = "Instance"
	DBInstanceEntityType      = "DBInstance"
	SecurityGroupEntityType   = "SecurityGroup"
	DBSecurityGroupEntityType = "DBSecurityGroup"
	ELBEntityType             = "LoadBalancerDescription"

	InstanceStoreType        = "ec2"
	DBInstanceStoreType      = "rds"
	SecurityGroupStoreType   = "security"
	DBSecurityGroupStoreType = "rds-security"
	ELBStoreType             = "elb"
)

var (
	ErrMissingInstanceId = errors.New("must provide instance id")
	ErrMissingGroupId    = errors.New("must provide group id")
	ErrMissingCustomerId = errors.New("must provide customer id")
)

func NewEntity(entityType, customerId, blob string) (interface{}, error) {
	var (
		err    error
		entity interface{}
	)

	switch entityType {
	case InstanceEntityType:
		instanceData := &InstanceData{}
		if err = json.Unmarshal([]byte(blob), instanceData); err != nil {
			break
		}
		if instanceData.InstanceId == "" {
			err = ErrMissingInstanceId
			break
		}
		entity = NewInstance(customerId, instanceData)

	case DBInstanceEntityType:
		dbInstanceData := &DBInstanceData{}
		if err = json.Unmarshal([]byte(blob), dbInstanceData); err != nil {
			break
		}
		if dbInstanceData.DBInstanceIdentifier == "" {
			err = ErrMissingInstanceId
			break
		}
		entity = NewInstance(customerId, dbInstanceData)

	case SecurityGroupEntityType:
		secGroupData := &SecurityGroupData{}
		if err = json.Unmarshal([]byte(blob), secGroupData); err != nil {
			break
		}
		if secGroupData.GroupId == "" {
			err = ErrMissingGroupId
			break
		}
		entity = NewGroup(customerId, secGroupData)

	case DBSecurityGroupEntityType:
		dbSecGroupData := &DBSecurityGroupData{}
		if err = json.Unmarshal([]byte(blob), dbSecGroupData); err != nil {
			break
		}
		if dbSecGroupData.DBSecurityGroupName == "" {
			err = ErrMissingGroupId
			break
		}
		entity = NewGroup(customerId, dbSecGroupData)

	case ELBEntityType:
		elbData := &ELBData{}
		if err = json.Unmarshal([]byte(blob), elbData); err != nil {
			break
		}
		if elbData.LoadBalancerName == "" {
			err = ErrMissingGroupId
			break
		}
		entity = NewGroup(customerId, elbData)
	}

	return entity, err
}

func NewInstance(customerId string, instanceData interface{}) *Instance {
	var (
		instance *Instance
		groups   []*Group
		jsonD    []byte
	)

	switch instanceData.(type) {
	case *InstanceData:
		iData := instanceData.(*InstanceData)
		groups = make([]*Group, len(iData.SecurityGroups))

		for i, group := range iData.SecurityGroups {
			gd := &SecurityGroupData{}
			if name, ok := group["GroupName"].(string); ok {
				gd.GroupName = name
			}
			if id, ok := group["GroupId"].(string); ok {
				gd.GroupId = id
			}
			groups[i] = NewGroup(customerId, gd)
		}

		jsonD, _ = json.Marshal(iData)
		instance = &Instance{
			Id:         iData.InstanceId,
			CustomerId: customerId,
			Type:       InstanceStoreType,
			Groups:     groups,
			Data:       jsonD,
		}

	case *DBInstanceData:
		dbiData := instanceData.(*DBInstanceData)
		dbSecLen := len(dbiData.DBSecurityGroups)
		groups = make([]*Group, dbSecLen+len(dbiData.VpcSecurityGroups))

		for i, group := range dbiData.DBSecurityGroups {
			sg := &DBSecurityGroupData{}
			if name, ok := group["DBSecurityGroupName"].(string); ok {
				sg.DBSecurityGroupName = name
			}
			groups[i] = NewGroup(customerId, sg)
		}

		for i, group := range dbiData.VpcSecurityGroups {
			sg := &SecurityGroupData{}
			if id, ok := group["VpcSecurityGroupId"].(string); ok {
				sg.GroupId = id
			}
			groups[i+dbSecLen] = NewGroup(customerId, sg)
		}

		jsonD, _ = json.Marshal(dbiData)
		instance = &Instance{
			Id:         dbiData.DBInstanceIdentifier,
			CustomerId: customerId,
			Type:       DBInstanceStoreType,
			Groups:     groups,
			Data:       jsonD,
		}
	}

	return instance
}

func NewGroup(customerId string, groupData interface{}) *Group {
	var (
		group *Group
		jsonD []byte
	)

	switch groupData.(type) {
	case *SecurityGroupData:
		sg := groupData.(*SecurityGroupData)
		jsonD, _ = json.Marshal(sg)
		group = &Group{
			CustomerId: customerId,
			Name:       sg.GroupId,
			Type:       SecurityGroupStoreType,
			Data:       jsonD,
		}

	case *DBSecurityGroupData:
		dbsg := groupData.(*DBSecurityGroupData)
		jsonD, _ = json.Marshal(dbsg)
		group = &Group{
			CustomerId: customerId,
			Name:       dbsg.DBSecurityGroupName,
			Type:       DBSecurityGroupStoreType,
			Data:       jsonD,
		}

	case *ELBData:
		elb := groupData.(*ELBData)
		instances := make([]*Instance, len(elb.Instances))
		for i, instance := range elb.Instances {
			dat := &InstanceData{}
			if id, ok := instance["InstanceId"].(string); ok {
				dat.InstanceId = id
			}
			instances[i] = NewInstance(customerId, dat)
		}
		jsonD, _ = json.Marshal(elb)
		group = &Group{
			CustomerId: customerId,
			Name:       elb.LoadBalancerName,
			Type:       ELBStoreType,
			Data:       jsonD,
			Instances:  instances,
		}
	}

	return group
}

func (i *Instance) MarshalJSON() ([]byte, error) {
	return i.Data, nil
}

func (g *Group) MarshalJSON() ([]byte, error) {
	return g.Data, nil
}
