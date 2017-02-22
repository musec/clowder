create table roles
(
	id serial primary key not null,
	name text not null,
	can_alter_users boolean not null default false
);

create table role_assignments
(
	id serial primary key not null,
	user_id integer not null,
	role_id integer not null,

	foreign key (user_id) references users(id),
	foreign key (role_id) references roles(id)
);
