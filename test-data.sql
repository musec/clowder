begin transaction;

delete from machines;
delete from reservations;
delete from users;

insert into users (username, name, email, phone) values
	('alice', 'Alice Aliceson', 'alice@example.com', '+1 (709) 123-4567'),
	('bob', 'Bob Balderson', 'bob@example.com', NULL)
	;

insert into machines (name, arch, microarch, cores, memory_gb) values
	('apple', 'x86_64', 'Sandy Bridge', 4, 16),
	('banana', 'x86_64', 'Haswell', 8, 128),
	('candy', 'arm64', 'armv7', 1, 4)
	;

insert into reservations (user_id, machine_id, scheduled_start, scheduled_end, actual_end, pxe_path, nfs_root) values
	(0, 0, timestamp with time zone '2017-01-01 03:30:00 -3:30', NULL, NULL, NULL, NULL),
	(0, 1, timestamp with time zone '2017-02-01 03:30:00 -3:30', NULL, NULL, NULL, NULL),
	(1, 2, timestamp with time zone '2017-03-01 03:30:00 -3:30', NULL, NULL, NULL, NULL)
	;

commit;
