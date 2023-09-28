create table if not exists user (
	id text primary key not null,
	name text not null,
	email text not null,
	hashed_password text not null,
	created_at datetime default current_timestamp,
	updated_at datetime default current_timestamp
);

