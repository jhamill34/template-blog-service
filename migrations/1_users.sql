create table if not exists user (
	id varchar(36) primary key not null,
	name text not null,
	email varchar(256) not null,
	hashed_password text not null,
	verified boolean default false,
	created_at datetime default current_timestamp,
	updated_at datetime default current_timestamp
);

create unique index idx_user_email on user (email);

grant select, insert, update, delete on `datadb`.`user` to `auth_user`@`%`;

