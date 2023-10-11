create table if not exists refresh_token (
	id int primary key not null auto_increment,
	user_id varchar(36) not null,
	app_id varchar(36) not null, 
	token varchar(36) not null,
	
	foreign key (user_id) references user(id),
	foreign key (app_id) references application(id)
);

create unique index idx_refresh_token on refresh_token (token);

grant select, insert, update, delete on `datadb`.`refresh_token` to `auth_user`@`%`;
