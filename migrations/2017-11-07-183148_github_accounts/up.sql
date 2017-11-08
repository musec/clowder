create table github_accounts (
	id serial primary key not null,
	user_id integer not null,
	github_username varchar not null,

	foreign key (user_id) references users(id)
);
