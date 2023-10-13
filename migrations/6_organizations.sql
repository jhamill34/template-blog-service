create table if not exists organization (
	id varchar(36) primary key not null,
	name varchar(255) not null,
	description text
);

create unique index idx_organization_name on organization (name);

create table if not exists organization_permission (
	id int primary key not null auto_increment,
	org_id varchar(36) not null,
	resource text not null,
	action text not null,
	effect text not null,

	foreign key (org_id) references organization(id)
);

create table if not exists organization_user (
	id int primary key not null auto_increment,
	org_id varchar(36) not null,
	user_id varchar(36) not null,

	foreign key (org_id) references organization(id),
	foreign key (user_id) references user(id)
);

create unique index idx_organization_user_ids on organization_user (org_id, user_id);

grant select, insert, update, delete on `datadb`.`organization` to `auth_user`@`%`;
grant select, insert, update, delete on `datadb`.`organization_permission` to `auth_user`@`%`;
grant select, insert, update, delete on `datadb`.`organization_user` to `auth_user`@`%`;

