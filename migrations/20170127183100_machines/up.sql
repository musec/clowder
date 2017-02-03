create table machines (
	id serial primary key not null,
	name varchar not null,
	arch varchar not null,
	microarch varchar not null,
	cores integer not null,
	memory_gb integer not null
)
