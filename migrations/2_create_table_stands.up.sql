create table if not exists stands (
		name text primary key unique not null,
		owner_username text references users(username) on delete cascade,
		released bool,
		time_claimed timestamp
	);