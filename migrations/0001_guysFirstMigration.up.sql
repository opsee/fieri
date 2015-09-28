CREATE FUNCTION update_time() RETURNS trigger LANGUAGE plpgsql AS $$
	BEGIN
		NEW.updated_at := CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
$$;



create type instance_type as enum ('ec2', 'rds', 'elb');
create table instances (
	id character varying(128) not null,
	customer_id UUID not null,
	type instance_type not null,
	data jsonb not null,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
	primary key (customer_id, id)
);

create index idx_instances_customers on instances (customer_id);
create trigger trg_instances_updated_at before update on instances for each row execute procedure update_time();



create type group_type as enum ('security', 'rds-security', 'elb', 'autoscaling', 'tag');
create table groups (
	name character varying(128) not null,
	customer_id UUID not null,
	type group_type not null,
	data jsonb not null,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
	primary key (customer_id, name)
);

create index idx_groups_customers on groups (customer_id);
create trigger trg_groups_updated_at before update on groups for each row execute procedure update_time();



