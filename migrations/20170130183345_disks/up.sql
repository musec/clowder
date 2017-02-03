create table disks (
	id serial primary key not null,
	machine_id integer not null,
	vendor text,
	model text,
	capacity_gb integer not null,
	ssd boolean not null,

	foreign key (machine_id) references machines(id)
);
