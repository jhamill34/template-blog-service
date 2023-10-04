create table if not exists role (
	id int primary key not null auto_increment,
	name text not null
);

create table if not exists role_permission (
	id int primary key not null auto_increment,
	role_id int not null,
	resource text not null,
	action text not null,
	effect text not null,

	foreign key (role_id) references role(id)
);

create table if not exists user_permission (
	id int primary key not null auto_increment,
	user_id varchar(36) not null,
	resource text not null,
	action text not null,
	effect text not null,

	foreign key (user_id) references user(id)
);

create table if not exists role_user (
	id int primary key not null auto_increment,

	role_id int not null,
	user_id varchar(36) not null,

	foreign key (role_id) references role(id),
	foreign key (user_id) references user(id)
);
