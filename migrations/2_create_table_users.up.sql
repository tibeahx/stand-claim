create table if not exists users (
    id int primary key not null,
    username text unique not null
)