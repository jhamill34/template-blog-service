create table if not exists user_permission (
	id int primary key not null auto_increment,
	user_id varchar(36) not null,
	resource text not null,
	action text not null,
	effect text not null,

	foreign key (user_id) references user(id)
);

grant select, insert, update, delete on `datadb`.`user_permission` to `auth_user`@`%`;
