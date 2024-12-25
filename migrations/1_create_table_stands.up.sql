create table if not exists stands (
		id uuid primary key not null,
		name text unique,
		owner_id int references users(id) on delete cascade,
		owner_username text,
		released bool,
		time_claimed timestamp
    );