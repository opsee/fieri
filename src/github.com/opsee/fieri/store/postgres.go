package store

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(connection string) (Store, error) {
	db, err := sqlx.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(8)

	return &Postgres{db: db}, nil
}

func (pg *Postgres) PutEntity(entity interface{}) (*EntityResponse, error) {
	var (
		err      error
		response *EntityResponse
	)

	switch entity.(type) {
	case *Instance:
		err = pg.putInstance(entity.(*Instance))
		response = &EntityResponse{entity.(*Instance).Type}

	case *Group:
		err = pg.putGroup(entity.(*Group))
		response = &EntityResponse{entity.(*Group).Type}
	}

	return response, err
}

func (pg *Postgres) GetInstance(request *InstanceRequest) (*InstanceResponse, error) {
	if request.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	if request.InstanceId == "" {
		return nil, ErrMissingInstanceId
	}

	instance := new(Instance)
	err := pg.db.Get(instance, "select * from instances where customer_id = $1 and id = $2", request.CustomerId, request.InstanceId)
	return &InstanceResponse{instance}, err
}

func (pg *Postgres) ListInstances(request *InstancesRequest) (*InstancesResponse, error) {
	instances, err := pg.listInstances(request)
	return &InstancesResponse{instances}, err
}

func (pg *Postgres) CountInstances(request *InstancesRequest) (*CountResponse, error) {
	if request.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	var err error
	var count int

	if request.Type == "" {
		err = pg.db.Get(&count, "select count(id) from instances where customer_id = $1", request.CustomerId)
	} else {
		err = pg.db.Get(&count, "select count(id) from instances where customer_id = $1 and type = $2", request.CustomerId, request.Type)
	}

	return &CountResponse{count}, err
}

func (pg *Postgres) DeleteInstances() error {
	_, err := pg.db.Exec("delete from instances")
	return err
}

func (pg *Postgres) GetGroup(request *GroupRequest) (*GroupResponse, error) {
	if request.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	if request.GroupId == "" {
		return nil, ErrMissingGroupId
	}

	group := new(Group)
	err := pg.db.Get(group, "select * from groups where customer_id = $1 and name = $2", request.CustomerId, request.GroupId)
	if err != nil {
		return nil, err
	}

	instances, err := pg.listInstances(&InstancesRequest{CustomerId: request.CustomerId, GroupId: request.GroupId, Type: request.Type})
	return &GroupResponse{group, instances}, err
}

func (pg *Postgres) ListGroups(request *GroupsRequest) (*GroupsResponse, error) {
	if request.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	var err error
	groups := make([]*Group, 0)

	if request.Type == "" {
		err = pg.db.Select(&groups, "select * from groups where customer_id = $1", request.CustomerId)
	} else {
		err = pg.db.Select(&groups, "select * from groups where customer_id = $1 and type = $2", request.CustomerId, request.Type)
	}

	return &GroupsResponse{groups}, err
}

func (pg *Postgres) CountGroups(request *GroupsRequest) (*CountResponse, error) {
	if request.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	var err error
	var count int

	if request.Type == "" {
		err = pg.db.Get(&count, "select count(name) from groups where customer_id = $1", request.CustomerId)
	} else {
		err = pg.db.Get(&count, "select count(name) from groups where customer_id = $1 and type = $2", request.CustomerId, request.Type)
	}

	return &CountResponse{count}, err
}

func (pg *Postgres) DeleteGroups() error {
	_, err := pg.db.Exec("delete from groups")
	return err
}

func (pg *Postgres) listInstances(request *InstancesRequest) ([]*Instance, error) {
	if request.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	var err error
	instances := make([]*Instance, 0)

	if request.GroupId != "" {
		err = pg.db.Select(&instances, "select * from instances where customer_id = $1 and id in (select instance_id from groups_instances where customer_id = $1 and group_name = $2)", request.CustomerId, request.GroupId)
	} else if request.Type == "" {
		err = pg.db.Select(&instances, "select * from instances where customer_id = $1", request.CustomerId)
	} else {
		err = pg.db.Select(&instances, "select * from instances where customer_id = $1 and type = $2", request.CustomerId, request.Type)
	}

	return instances, err
}

func (pg *Postgres) putInstance(instance *Instance) error {
	query := "with update_instances as (update instances set (type, data) = (:type, :data) where id = :id and customer_id = :customer_id returning id), insert_instances as (insert into instances (id, customer_id, type, data) select :id as id, :customer_id as customer_id, :type as type, :data as data where not exists (select id from update_instances limit 1) returning id) select * from update_instances union all select * from insert_instances;"
	_, err := pg.db.NamedExec(query, instance)
	if err != nil {
		return err
	}

	// i don't really want to use transactions for this right now until a refactor
	for _, group := range instance.Groups {
		err := pg.ensureGroup(group)
		if err != nil {
			return err
		}

		_, err = pg.db.Exec("insert into groups_instances (customer_id, group_name, instance_id) select $1 as customer_id, ($2::varchar(128)) as group_name, ($3::varchar(128)) as instance_id where not exists (select instance_id from groups_instances where customer_id = $1 and group_name = $2 and instance_id = $3)", group.CustomerId, group.Name, instance.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pg *Postgres) putGroup(group *Group) error {
	query := "with update_groups as (update groups set (type, data) = (:type, :data) where name = :name and customer_id = :customer_id returning name), insert_groups as (insert into groups (name, customer_id, type, data) select :name as name, :customer_id as customer_id, :type as type, :data as data where not exists (select name from update_groups limit 1) returning name) select * from update_groups union all select * from insert_groups;"
	_, err := pg.db.NamedExec(query, group)
	if err != nil {
		return err
	}

	// i don't really want to use transactions for this right now until a refactor
	for _, instance := range group.Instances {
		err := pg.ensureInstance(instance)
		if err != nil {
			return err
		}

		_, err = pg.db.Exec("insert into groups_instances (customer_id, group_name, instance_id) select $1 as customer_id, ($2::varchar(128)) as group_name, ($3::varchar(128)) as instance_id where not exists (select instance_id from groups_instances where customer_id = $1 and group_name = $2 and instance_id = $3)", group.CustomerId, group.Name, instance.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pg *Postgres) ensureInstance(instance *Instance) error {
	_, err := pg.db.Exec("insert into instances (id, customer_id, type, data) select ($1::varchar(128)) as id, $2 as customer_id, $3 as type, $4 as data where not exists (select id from instances where id = $1 and customer_id = $2)", instance.Id, instance.CustomerId, instance.Type, instance.Data)
	return err
}

func (pg *Postgres) ensureGroup(group *Group) error {
	_, err := pg.db.Exec("insert into groups (name, customer_id, type, data) select ($1::varchar(128)) as name, $2 as customer_id, $3 as type, $4 as data where not exists (select name from groups where name = $1 and customer_id = $2)", group.Name, group.CustomerId, group.Type, group.Data)
	return err
}
