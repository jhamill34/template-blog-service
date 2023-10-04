create table if not exists application (
	id varchar(36) primary key not null,
	client_id varchar(36) not null,
	hashed_client_secret text not null,
	redirect_uri text not null,
	name varchar(255) not null,
	description text,
	created_at timestamp not null default current_timestamp,
	updated_at timestamp not null default current_timestamp
)
