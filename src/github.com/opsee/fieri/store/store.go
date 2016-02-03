package store

import (
	"encoding/json"
	"errors"
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

type InstanceData struct {
	State            map[string]interface{}
	Monitoring       map[string]interface{}
	PublicDnsName    string
	LaunchTime       *time.Time
	PublicIpAddress  string
	PrivateIpAddress string
	InstanceId       string
	ImageId          string
	PrivateDnsName   string
	KeyName          string
	SecurityGroups   []map[string]interface{}
	SubnetId         string
	InstanceType     string
	Placement        map[string]interface{}
	Tags             []map[string]interface{}
}

type DBInstanceData struct {
	PubliclyAccessible               bool
	VpcSecurityGroups                []map[string]interface{}
	InstanceCreateTime               *time.Time
	OptionGroupMemberships           []map[string]interface{}
	Engine                           string
	MultiAZ                          bool
	LatestRestorableTime             *time.Time
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

type AutoScalingGroupData struct {
	AutoScalingGroupName    string
	AvailabilityZones       []string
	CreatedTime             *time.Time
	DefaultCooldown         int
	DesiredCapacity         int
	HealthCheckGracePeriod  int
	HealthCheckType         string
	InstanceId              string
	Instances               []map[string]interface{}
	LaunchConfigurationName string
	LoadBalancerNames       []map[string]interface{}
	MaxSize                 int
	MinSize                 int
	Name                    string
	Status                  string
	Tags                    []map[string]interface{}
	VPCZoneIdentifier       string
	TerminationPolicies     []string
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
	CreatedTime               *time.Time
	AvailabilityZones         []string
	Scheme                    string
	SourceSecurityGroup       map[string]interface{}
}

type RouteTableData struct {
	Associations    []map[string]interface{}
	RouteTableId    string
	VpcId           string
	Tags            []map[string]interface{}
	Routes          []map[string]interface{}
	PropagatingVgws []map[string]interface{}
}

type SubnetData struct {
	VpcId                   string
	Tags                    []map[string]interface{}
	CidrBlock               string
	MapPublicIpOnLaunch     bool
	DefaultForAz            bool
	State                   string
	AvailabilityZone        string
	SubnetId                string
	AvailableIpAddressCount int
}

const (
	InstanceEntityType         = "Instance"
	DBInstanceEntityType       = "DBInstance"
	SecurityGroupEntityType    = "SecurityGroup"
	DBSecurityGroupEntityType  = "DBSecurityGroup"
	AutoScalingGroupEntityType = "AutoScalingGroup"
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
		instanceData := &InstanceData{}
		if err = json.Unmarshal(blob, instanceData); err != nil {
			break
		}
		if instanceData.InstanceId == "" {
			err = ErrMissingInstanceId
			break
		}
		entity = NewInstance(customerId, instanceData)

	case DBInstanceEntityType:
		dbInstanceData := &DBInstanceData{}
		if err = json.Unmarshal(blob, dbInstanceData); err != nil {
			break
		}
		if dbInstanceData.DBInstanceIdentifier == "" {
			err = ErrMissingInstanceId
			break
		}
		entity = NewInstance(customerId, dbInstanceData)

	case SecurityGroupEntityType:
		secGroupData := &SecurityGroupData{}
		if err = json.Unmarshal(blob, secGroupData); err != nil {
			break
		}
		if secGroupData.GroupId == "" {
			err = ErrMissingGroupId
			break
		}
		entity = NewGroup(customerId, secGroupData)

	case DBSecurityGroupEntityType:
		dbSecGroupData := &DBSecurityGroupData{}
		if err = json.Unmarshal(blob, dbSecGroupData); err != nil {
			break
		}
		if dbSecGroupData.DBSecurityGroupName == "" {
			err = ErrMissingGroupId
			break
		}
		entity = NewGroup(customerId, dbSecGroupData)

	case ELBEntityType:
		elbData := &ELBData{}
		if err = json.Unmarshal(blob, elbData); err != nil {
			break
		}
		if elbData.LoadBalancerName == "" {
			err = ErrMissingGroupId
			break
		}
		entity = NewGroup(customerId, elbData)

	case AutoScalingGroupEntityType:
		autoscalingData := &AutoScalingGroupData{}
		if err = json.Unmarshal(blob, autoscalingData); err != nil {
			break
		}

		if autoscalingData.AutoScalingGroupName == "" {
			err = ErrMissingGroupId
			break
		}
		entity = NewGroup(customerId, autoscalingData)

	case RouteTableEntityType:
		routeTableData := &RouteTableData{}
		if err = json.Unmarshal(blob, routeTableData); err != nil {
			break
		}

		if routeTableData.RouteTableId == "" {
			err = ErrMissingRouteTableId
			break
		}
		entity = NewRouteTable(customerId, routeTableData)

	case SubnetEntityType:
		subnetData := &SubnetData{}
		if err = json.Unmarshal(blob, subnetData); err != nil {
			break
		}

		if subnetData.SubnetId == "" {
			err = ErrMissingSubnetId
			break
		}
		entity = NewSubnet(customerId, subnetData)
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

	case *AutoScalingGroupData:
		autoscaling := groupData.(*AutoScalingGroupData)
		instances := make([]*Instance, len(autoscaling.Instances))
		for i, instance := range autoscaling.Instances {
			dat := &InstanceData{}
			if id, ok := instance["InstanceId"].(string); ok {
				dat.InstanceId = id
			}
			instances[i] = NewInstance(customerId, dat)
		}
		jsonD, _ = json.Marshal(autoscaling)
		group = &Group{
			CustomerId: customerId,
			Name:       autoscaling.AutoScalingGroupName,
			Type:       AutoScalingGroupStoreType,
			Data:       jsonD,
			Instances:  instances,
		}

	}
	return group
}

func NewRouteTable(customerId string, routeTableData *RouteTableData) *RouteTable {
	jsonD, _ := json.Marshal(routeTableData)

	return &RouteTable{
		Id:         routeTableData.RouteTableId,
		CustomerId: customerId,
		Data:       jsonD,
	}
}

func NewSubnet(customerId string, subnetData *SubnetData) *Subnet {
	jsonD, _ := json.Marshal(subnetData)

	return &Subnet{
		Id:         subnetData.SubnetId,
		CustomerId: customerId,
		Data:       jsonD,
	}
}

func (i *Instance) MarshalJSON() ([]byte, error) {
	return i.Data, nil
}

func (g *Group) MarshalJSON() ([]byte, error) {
	return g.Data, nil
}
