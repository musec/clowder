create table machines (
	id serial primary key not null,
	name varchar not null,
	arch varchar not null,
	microarch varchar not null,
	cores integer not null,
	memory_gb integer not null
);

create table disks (
	id serial primary key not null,
	machine_id integer not null,
	vendor text,
	model text,
	capacity_gb integer not null,
	ssd boolean not null,

	foreign key (machine_id) references machines(id)
);

create table nics (
	id serial primary key not null,
	machine_id integer not null,
	vendor text,
	model text,
	mac_address char(12) not null,
	speed_gbps integer not null,

	foreign key (machine_id) references machines(id)
);

create table users (
	id serial primary key not null,
	username varchar not null,
	name text not null,
	phone text
);

create table emails (
	id serial primary key not null,
	user_id integer not null,
	email text not null,

	foreign key (user_id) references users(id)
);

create table roles
(
	id serial primary key not null,
	name text not null,
	can_alter_users boolean not null default false,
	can_view_users boolean not null default false
);

create table role_assignments
(
	id serial primary key not null,
	user_id integer not null,
	role_id integer not null,

	foreign key (user_id) references users(id),
	foreign key (role_id) references roles(id)
);

create table reservations (
	id serial primary key not null,
	user_id integer not null,
	machine_id integer not null,
	scheduled_start timestamp with time zone not null,
	scheduled_end timestamp with time zone,
	actual_end timestamp with time zone,
	pxe_path text,
	nfs_root text,

	foreign key (user_id) references users(id),
	foreign key (machine_id) references machines(id)
);
