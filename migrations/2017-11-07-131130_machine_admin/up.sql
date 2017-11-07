alter table roles add column can_alter_machines boolean default false;
alter table roles add column can_create_machines boolean default false;
alter table roles add column can_delete_machines boolean default false;
