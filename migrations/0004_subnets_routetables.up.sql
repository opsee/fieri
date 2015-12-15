create table route_tables (
  id character varying(128) not null,
  customer_id UUID not null,
  data jsonb not null,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
	primary key (customer_id, id)
);

create index idx_route_tables_customers on route_tables (customer_id);
create trigger trg_route_tables_updated_at before update on route_tables for each row execute procedure update_time();

create table subnets (
  id character varying(128) not null,
  customer_id UUID not null,
  data jsonb not null,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
	primary key (customer_id, id)
);

create index idx_subnets_customers on subnets (customer_id);
create trigger trg_subnets_updated_at before update on subnets for each row execute procedure update_time();
