create table groups_instances (
  customer_id UUID not null,
  group_name character varying(128) not null,
  instance_id character varying(128) not null,
  foreign key (customer_id, group_name) references groups (customer_id, name) on delete cascade,
  foreign key (customer_id, instance_id) references instances (customer_id, id) on delete cascade,
  unique (customer_id, group_name, instance_id)
);
