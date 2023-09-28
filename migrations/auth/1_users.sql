create table if not exists user (
	id text primary key not null,
	name text not null,
	email text not null,
	hashed_password text not null,
	created_at datetime default current_timestamp,
	updated_at datetime default current_timestamp
);

insert into user 
	( id, name, email, hashed_password ) 
values 
	(
		"4667297c-26b0-47dd-b566-edfcac16f3a7", 
		"root",
		"root@example.com", 
		"$argon2id$v=19$m=32768,t=3,p=4$Op4KBD95M7mtNob5lIXOHQ$sapnaoDRlAni1SodVlhMOb9kEeJ3WWlswNAH18rfEZ8"
	);
	

