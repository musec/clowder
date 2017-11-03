create table architectures (
	id serial primary key not null,
	name varchar not null
);

insert into architectures (name) values ('i386'), ('x86_64');

create table microarchitectures (
	id serial primary key not null,
	arch_id integer not null,
	name varchar not null,
	url varchar,

	foreign key (arch_id) references architectures(id)
);

create table processors (
	id serial primary key not null,
	microarch_id integer not null,
	name varchar not null,
	cores integer not null,
	threads integer not null,
	freq_ghz float not null,
	url varchar,

	foreign key (microarch_id) references microarchitectures(id)
);

create table machines (
	id serial primary key not null,
	name varchar not null,
	processor_id integer not null,
	memory_gb integer not null,

	foreign key (processor_id) references processors(id)
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
	name text not null
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
