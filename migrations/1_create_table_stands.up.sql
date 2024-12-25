create table if not exists stands (
		id uuid primary key not null,
		name text unique,
		owner_id int,
		released bool,
		owner_username text,
		time_claimed timestamp,
    );