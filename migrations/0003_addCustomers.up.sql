create table customers (
  id UUID not null,
  last_sync timestamp with time zone,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
	primary key (id)
);

create trigger trg_customers_updated_at before update on customers for each row execute procedure update_time();

insert into customers (id, last_sync) select customer_id, updated_at from instances group by customer_id, updated_at order by updated_at desc limit 1;
