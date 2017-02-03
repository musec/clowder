create table nics (
	id serial primary key not null,
	machine_id integer not null,
	vendor text,
	model text,
	mac_address char(12) not null,
	speed_gbps integer not null,

	foreign key (machine_id) references machines(id)
);
