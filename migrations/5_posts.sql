create table if not exists post (
	id varchar(36) primary key not null,
	title varchar(255) not null,
	content text not null,
	author varchar(36) not null,
	created_at timestamp not null default current_timestamp,
	updated_at timestamp not null default current_timestamp
);

grant select, insert, update, delete on `datadb`.`post` to `app_user`@`%`;
