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

func (pg *Postgres) PutInstance(instance *Instance) error {
	query := "with update_instances as (update instances set (type, data) = (:type, :data) where id = :id and customer_id = :customer_id returning id), insert_instances as (insert into instances (id, customer_id, type, data) select :id as id, :customer_id as customer_id, :type as type, :data as data where not exists (select id from update_instances limit 1) returning id) select * from update_instances union all select * from insert_instances;"
	_, err := pg.db.NamedExec(query, instance)

	return err
}

func (pg *Postgres) GetInstance(options *Options) (*Instance, error) {
	if options.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	if options.InstanceId == "" {
		return nil, ErrMissingInstanceId
	}

	instance := new(Instance)
	err := pg.db.Get(instance, "select * from instances where customer_id = $1 and id = $2", options.CustomerId, options.InstanceId)
	return instance, err
}

func (pg *Postgres) ListInstances(options *Options) ([]*Instance, error) {
	if options.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	var err error
	instances := make([]*Instance, 0)

	if options.Type == "" {
		err = pg.db.Select(&instances, "select * from instances where customer_id = $1", options.CustomerId)
	} else {
		err = pg.db.Select(&instances, "select * from instances where customer_id = $1 and type = $2", options.CustomerId, options.Type)
	}

	return instances, err
}

func (pg *Postgres) CountInstances(options *Options) (int, error) {
	if options.CustomerId == "" {
		return 0, ErrMissingCustomerId
	}

	var err error
	var count int

	if options.Type == "" {
		err = pg.db.Get(&count, "select count(id) from instances where customer_id = $1", options.CustomerId)
	} else {
		err = pg.db.Get(&count, "select count(id) from instances where customer_id = $1 and type = $2", options.CustomerId, options.Type)
	}

	return count, err
}

func (pg *Postgres) DeleteInstances() error {
	_, err := pg.db.Exec("delete from instances")
	return err
}

func (pg *Postgres) PutGroup(group *Group) error {
	query := "with update_groups as (update groups set (type, data) = (:type, :data) where name = :name and customer_id = :customer_id returning name), insert_groups as (insert into groups (name, customer_id, type, data) select :name as name, :customer_id as customer_id, :type as type, :data as data where not exists (select name from update_groups limit 1) returning name) select * from update_groups union all select * from insert_groups;"
	_, err := pg.db.NamedExec(query, group)

	return err
}

func (pg *Postgres) GetGroup(options *Options) (*Group, error) {
	if options.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	if options.GroupId == "" {
		return nil, ErrMissingGroupId
	}

	group := new(Group)
	err := pg.db.Get(group, "select * from groups where customer_id = $1 and name = $2", options.CustomerId, options.GroupId)
	return group, err
}

func (pg *Postgres) ListGroups(options *Options) ([]*Group, error) {
	if options.CustomerId == "" {
		return nil, ErrMissingCustomerId
	}

	var err error
	groups := make([]*Group, 0)

	if options.Type == "" {
		err = pg.db.Select(&groups, "select * from groups where customer_id = $1", options.CustomerId)
	} else {
		err = pg.db.Select(&groups, "select * from groups where customer_id = $1 and type = $2", options.CustomerId, options.Type)
	}

	return groups, err
}

func (pg *Postgres) CountGroups(options *Options) (int, error) {
	if options.CustomerId == "" {
		return 0, ErrMissingCustomerId
	}

	var err error
	var count int

	if options.Type == "" {
		err = pg.db.Get(&count, "select count(name) from groups where customer_id = $1", options.CustomerId)
	} else {
		err = pg.db.Get(&count, "select count(name) from groups where customer_id = $1 and type = $2", options.CustomerId, options.Type)
	}

	return count, err
}

func (pg *Postgres) DeleteGroups() error {
	_, err := pg.db.Exec("delete from groups")
	return err
}
