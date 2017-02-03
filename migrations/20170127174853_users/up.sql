create table users (
	id serial primary key not null,
	username varchar not null,
	name text not null,
	email text not null,
	phone text
)
