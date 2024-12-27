create table if not exists users (
    username text unique primary key not null,
    created timestamp
);