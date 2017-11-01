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
