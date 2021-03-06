/*
 * Copyright 2016-2018 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use chrono::{DateTime, Utc};
use db::schema::*;
use diesel;
use diesel::pg::PgConnection as Connection;
use diesel::*;
use itertools::Itertools;
use std::collections::HashSet;

type DieselResult<T> = Result<T, diesel::result::Error>;

allow_tables_to_appear_in_same_query! { github_accounts, users }
allow_tables_to_appear_in_same_query! { machines, microarchitectures }
allow_tables_to_appear_in_same_query! { machines, architectures }
allow_tables_to_appear_in_same_query! { machines, processors }
allow_tables_to_appear_in_same_query! { machines, users }
allow_tables_to_appear_in_same_query! { microarchitectures, architectures }
allow_tables_to_appear_in_same_query! { processors, architectures }
allow_tables_to_appear_in_same_query! { processors, microarchitectures }
allow_tables_to_appear_in_same_query! { reservations, machines }
allow_tables_to_appear_in_same_query! { reservations, users }
allow_tables_to_appear_in_same_query! { role_assignments, roles }

#[derive(Debug, Identifiable, Queryable)]
pub struct User {
    pub id: i32,
    pub username: String,
    pub name: String,
}

impl User {
    pub fn all(c: &Connection) -> DieselResult<Vec<User>> {
        use self::users::dsl::*;
        users.order(username).load(c)
    }

    pub fn get(uid: i32, c: &Connection) -> DieselResult<User> {
        use db::schema::users::dsl::*;
        users.find(uid).first(c)
    }

    pub fn with_email(address: &str, c: &Connection) -> DieselResult<User> {
        use self::emails::dsl::*;

        emails
            .filter(email.eq(address))
            .first(c)
            .map(|e: Email| e.user_id)
            .and_then(|uid| User::get(uid, c))
    }

    ///
    /// Change a User's name in the database. This consumes the existing User object and returns
    /// a new User object with the updated name.
    ///
    pub fn change_name<S>(self, new_name: S, c: &Connection) -> DieselResult<User>
    where
        S: Into<String>,
    {
        use self::users::dsl::*;

        diesel::update(&self)
            .set(name.eq(new_name.into()))
            .get_result::<User>(c)
    }

    pub fn emails(&self, c: &Connection) -> DieselResult<HashSet<String>> {
        use db::schema::emails::dsl::*;
        emails
            .filter(user_id.eq(self.id))
            .load(c)
            .map(|user_emails| user_emails.into_iter().map(|e: Email| e.email).collect())
    }

    ///
    /// Update complete set of email addresses associated with this user.
    /// This will delete existing email addresses if they are not in the new set.
    ///
    pub fn set_emails(&self, new_emails: &HashSet<String>, c: &Connection) -> DieselResult<()> {
        let current_emails = self.emails(c)?;

        use self::emails::dsl::*;

        for e in &current_emails {
            if !new_emails.contains(e) {
                diesel::delete(emails.filter(email.eq(e))).execute(c)?;
            }
        }

        for e in new_emails.difference(&current_emails) {
            Email::insert(self, e.clone(), c)?;
        }

        Ok(())
    }

    ///
    /// Update complete set of roles assigned to this user.
    /// This will delete existing role assignments if they are not in the new set.
    ///
    pub fn set_roles(&self, new_roles: &HashSet<String>, c: &Connection) -> DieselResult<()> {
        let current_roles = self.roles(c)?;

        let role_names = current_roles
            .iter()
            .map(|ref role| role.name.clone())
            .collect::<HashSet<_>>();

        use self::role_assignments::dsl::*;

        for role in &current_roles {
            if !new_roles.contains(&role.name) {
                let to_delete = role_assignments
                    .filter(role_id.eq(role.id))
                    .filter(user_id.eq(self.id));

                diesel::delete(to_delete).execute(c)?;
            }
        }

        for ref role_name in new_roles.difference(&role_names) {
            Role::with_name(role_name, c)
                .and_then(|role| RoleAssignment::insert(self, &role, c))?;
        }

        Ok(())
    }

    pub fn username(&self) -> &str {
        &self.username
    }

    pub fn with_username(uname: &str, c: &Connection) -> DieselResult<User> {
        use self::users::dsl::*;
        users.filter(username.eq(uname)).first(c)
    }

    pub fn inhabits_role(&self, role: &Role, c: &Connection) -> DieselResult<bool> {
        use self::role_assignments::dsl::*;
        role_assignments
            .inner_join(roles::table)
            .filter(user_id.eq(self.id))
            .filter(role_id.eq(role.id))
            .count()
            .get_result::<i64>(c)
            .map(|count| count > 0)
    }

    pub fn roles(&self, c: &Connection) -> DieselResult<Vec<Role>> {
        use self::role_assignments::dsl::*;
        role_assignments
            .inner_join(roles::table)
            .filter(user_id.eq(self.id))
            .load(c)
            .map(|roles: Vec<(RoleAssignment, Role)>| roles.into_iter().map(|(_, r)| r).collect())
    }

    pub fn can_alter_machines(&self, c: &Connection) -> DieselResult<bool> {
        self.has_role(c, |ref role| role.can_alter_machines)
    }

    pub fn can_alter_users(&self, c: &Connection) -> DieselResult<bool> {
        self.has_role(c, |ref role| role.can_alter_users)
    }

    pub fn can_create_machines(&self, c: &Connection) -> DieselResult<bool> {
        self.has_role(c, |ref role| role.can_create_machines)
    }

    pub fn can_delete_machines(&self, c: &Connection) -> DieselResult<bool> {
        self.has_role(c, |ref role| role.can_delete_machines)
    }

    pub fn can_view_users(&self, c: &Connection) -> DieselResult<bool> {
        self.has_role(c, |ref role| role.can_view_users)
    }

    /// Does any of this user's roles satisfy a predicate?
    fn has_role<Pred>(&self, c: &Connection, predicate: Pred) -> DieselResult<bool>
    where
        Pred: Fn(&Role) -> bool,
    {
        self.roles(c).map(|roles| roles.iter().any(predicate))
    }
}

#[derive(Associations, Debug, Identifiable, Queryable)]
pub struct GithubAccount {
    pub id: i32,
    pub user_id: i32,
    pub github_username: String,
}

impl GithubAccount {
    pub fn get(gh_username: &str, conn: &Connection) -> DieselResult<(GithubAccount, User)> {
        use self::github_accounts::dsl::*;
        github_accounts
            .inner_join(users::table)
            .filter(github_username.eq(gh_username))
            .first(conn)
    }
}

#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(User)]
pub struct Email {
    pub id: i32,
    pub user_id: i32,
    pub email: String,
}

#[derive(Debug, Insertable)]
#[table_name = "emails"]
struct EmailInserter {
    user_id: i32,
    email: String,
}

impl Email {
    pub fn insert<S>(user: &User, email: S, conn: &Connection) -> DieselResult<Email>
    where
        S: Into<String>,
    {
        diesel::insert_into(emails::table)
            .values(&EmailInserter {
                user_id: user.id,
                email: email.into(),
            })
            .get_result(conn)
    }
}

#[derive(Associations, Debug, Identifiable, Queryable)]
pub struct Role {
    pub id: i32,
    pub name: String,
    pub can_alter_machines: bool,
    pub can_alter_users: bool,
    pub can_create_machines: bool,
    pub can_delete_machines: bool,
    pub can_view_users: bool,
}

impl Role {
    pub fn all(c: &Connection) -> DieselResult<Vec<Role>> {
        use self::roles::dsl::*;
        roles.order(name).load(c)
    }

    pub fn with_name(role_name: &str, c: &Connection) -> DieselResult<Role> {
        use self::roles::dsl::*;
        roles.filter(name.eq(role_name)).first(c)
    }
}

#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(Role)]
#[belongs_to(User)]
pub struct RoleAssignment {
    pub id: i32,
    pub user_id: i32,
    pub role_id: i32,
}

#[derive(Debug, Insertable)]
#[table_name = "role_assignments"]
struct RoleAssigner {
    user_id: i32,
    role_id: i32,
}

impl RoleAssignment {
    pub fn insert(user: &User, role: &Role, conn: &Connection) -> DieselResult<RoleAssignment> {
        diesel::insert_into(role_assignments::table)
            .values(&RoleAssigner {
                user_id: user.id,
                role_id: role.id,
            })
            .get_result(conn)
    }
}

#[derive(Debug, Identifiable, Queryable)]
pub struct Architecture {
    pub id: i32,
    pub name: String,
}

#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(Architecture, foreign_key = "arch_id")]
pub struct Microarchitecture {
    pub id: i32,
    pub arch_id: i32,
    pub name: String,
    pub url: Option<String>,
}

impl Microarchitecture {
    pub fn arch(&self, c: &Connection) -> DieselResult<Architecture> {
        use self::architectures::dsl::*;
        architectures.find(self.arch_id).first(c)
    }
}

#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(Microarchitecture, foreign_key = "microarch_id")]
pub struct Processor {
    pub id: i32,
    pub microarch_id: i32,
    pub name: String,
    pub cores: i32,
    pub threads: i32,
    pub freq_ghz: f64,
    pub url: Option<String>,
}

impl Processor {
    pub fn all(c: &Connection) -> DieselResult<Vec<Processor>> {
        use self::processors::dsl::*;
        processors.order(name).load(c)
    }

    pub fn get(processor_id: i32, c: &Connection) -> DieselResult<Processor> {
        use self::processors::dsl::*;
        processors.find(processor_id).first(c)
    }
}

#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(Processor)]
pub struct Machine {
    pub id: i32,
    pub name: String,
    pub processor_id: i32,
    pub memory_gb: i32,
}

impl Machine {
    pub fn all(c: &Connection) -> DieselResult<Vec<Machine>> {
        use self::machines::dsl::*;
        machines.order(name).load(c)
    }

    pub fn disks(&self, c: &Connection) -> DieselResult<Vec<Disk>> {
        Disk::belonging_to(self).load(c)
    }

    pub fn nics(&self, c: &Connection) -> DieselResult<Vec<Nic>> {
        Nic::belonging_to(self).load(c)
    }

    pub fn with_name(machine_name: &str, c: &Connection) -> DieselResult<Machine> {
        use self::machines::dsl::*;
        machines.filter(name.eq(machine_name)).first(c)
    }
}

#[derive(Debug, Insertable)]
#[table_name = "machines"]
pub struct MachineBuilder {
    name: String,
    processor_id: Option<i32>,
    memory_gb: Option<i32>,
}

impl MachineBuilder {
    pub fn new(name: String) -> MachineBuilder {
        MachineBuilder {
            name: name,
            processor_id: None,
            memory_gb: None,
        }
    }

    pub fn insert(self, conn: &Connection) -> DieselResult<Machine> {
        insert_into(machines::table).values(&self).get_result(conn)
    }

    pub fn memory_gb(mut self, mem: i32) -> MachineBuilder {
        self.memory_gb = Some(mem);
        self
    }

    pub fn processor(mut self, p: &Processor) -> MachineBuilder {
        self.processor_id = Some(p.id);
        self
    }
}

///
/// A FullMachine is a complete representation of a machine and all of its architectural details.
///
pub struct FullMachine {
    machine: Machine,
    processor: Processor,
    microarch: Microarchitecture,
    arch: Architecture,
}

type FullMachineJoin = (Machine, (Processor, (Microarchitecture, Architecture)));

impl FullMachine {
    fn from(data: FullMachineJoin) -> FullMachine {
        let (machine, (processor, (microarch, arch))) = data;

        FullMachine {
            machine: machine,
            processor: processor,
            microarch: microarch,
            arch: arch,
        }
    }

    pub fn all(c: &Connection) -> DieselResult<Vec<FullMachine>> {
        use self::machines::dsl::*;
        let m = machines
            .order(name)
            .inner_join(
                processors::table
                    .inner_join(microarchitectures::table.inner_join(architectures::table)),
            )
            .load(c)?;

        Ok(m.into_iter().map(FullMachine::from).collect())
    }

    pub fn with_name(machine_name: &str, c: &Connection) -> DieselResult<FullMachine> {
        use self::machines::dsl::*;
        machines
            .filter(name.eq(machine_name))
            .inner_join(
                processors::table
                    .inner_join(microarchitectures::table.inner_join(architectures::table)),
            )
            .first(c)
            .map(FullMachine::from)
    }

    pub fn architecture(&self) -> &Architecture {
        &self.arch
    }

    pub fn cores(&self) -> i32 {
        self.processor.cores
    }

    pub fn freq_ghz(&self) -> f64 {
        self.processor.freq_ghz
    }

    pub fn machine(&self) -> &Machine {
        &self.machine
    }

    pub fn memory_gb(&self) -> i32 {
        self.machine.memory_gb
    }

    pub fn microarchitecture(&self) -> &Microarchitecture {
        &self.microarch
    }

    pub fn name(&self) -> &str {
        &self.machine.name
    }

    pub fn processor(&self) -> &Processor {
        &self.processor
    }
}

#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(Machine)]
pub struct Disk {
    pub id: i32,
    pub machine_id: i32,
    pub vendor: Option<String>,
    pub model: Option<String>,
    pub capacity_gb: i32,
    pub ssd: bool,
}

impl Disk {
    pub fn short_description(&self) -> String {
        let v = self
            .vendor
            .as_ref()
            .map(|vendor| format!["{} ", vendor])
            .unwrap_or(String::new());

        let ssd = if self.ssd { "SSD" } else { "non-SSD" };

        format!["{}{} GiB ({})", v, self.capacity_gb, ssd]
    }

    pub fn vendor_name(&self) -> &str {
        self.vendor
            .as_ref()
            .map(|s| s.as_str())
            .unwrap_or("<unknown vendor>")
    }
}

#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(Machine)]
pub struct Nic {
    pub id: i32,
    pub machine_id: i32,
    pub vendor: Option<String>,
    pub model: Option<String>,
    pub mac_address: String,
    pub speed_gbps: i32,
}

impl Nic {
    pub fn short_description(&self) -> String {
        let vendor: String = self
            .vendor
            .as_ref()
            .map(|vendor| format!["{} ", vendor])
            .unwrap_or(String::new());

        let model: String = self
            .model
            .as_ref()
            .map(|m| format!["{} ", m])
            .unwrap_or(String::new());

        format![
            "{} — {}{} Gbps",
            self.mac_formatted(),
            vendor + &model,
            self.speed_gbps
        ]
    }

    pub fn mac_formatted(&self) -> String {
        self.mac_address
            .chars()
            .chunks(2)
            .into_iter()
            .map(squash_chars)
            .collect::<Vec<_>>()
            .join(":")
    }

    pub fn vendor_name(&self) -> &str {
        self.vendor
            .as_ref()
            .map(|s| s.as_str())
            .unwrap_or("<unknown vendor>")
    }
}

fn squash_chars<C>(chunk: C) -> String
where
    C: Iterator<Item = char>,
{
    chunk.fold(String::new(), |mut s, c| {
        s.push(c);
        s
    })
}

#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(Machine)]
#[belongs_to(User)]
pub struct Reservation {
    pub id: i32,
    pub user_id: i32,
    pub machine_id: i32,
    pub scheduled_start: DateTime<Utc>,
    pub scheduled_end: Option<DateTime<Utc>>,
    pub actual_end: Option<DateTime<Utc>>,
    pub pxe_path: Option<String>,
    pub nfs_root: Option<String>,
}

type FullReservation = (Reservation, Machine, User);

impl Reservation {
    /// Find all reservations, ordered by end time.
    pub fn all(only_current: bool, c: &Connection) -> DieselResult<Vec<FullReservation>> {
        use self::reservations::dsl::*;

        let query = reservations
            .inner_join(machines::table)
            .inner_join(users::table)
            .order(scheduled_end.desc())
            .order(machine_id);

        if only_current {
            query.filter(actual_end.is_null()).load(c)
        } else {
            query.load(c)
        }
    }

    ///
    /// Find all of a machine's reservations (and the User that reserved it in each case).
    ///
    pub fn for_machine(m: &Machine, c: &Connection) -> DieselResult<Vec<(Reservation, User)>> {
        use self::reservations::dsl::*;
        reservations
            .inner_join(users::table)
            .filter(machine_id.eq(m.id()))
            .filter(user_id.eq(users::dsl::id))
            .order(actual_end.desc())
            .order(scheduled_end.desc())
            .load(c)
    }

    ///
    /// Find all of a user's machine reservations (and the details of those machines).
    ///
    pub fn for_user(user: &User, c: &Connection) -> DieselResult<Vec<(Reservation, Machine)>> {
        use self::reservations::dsl::*;
        reservations
            .inner_join(machines::table)
            .filter(user_id.eq(user.id()))
            .order(actual_end.desc())
            .order(scheduled_end.desc())
            .load(c)
    }

    pub fn get(res_id: i32, c: &Connection) -> DieselResult<FullReservation> {
        use db::schema::reservations::dsl::*;

        reservations
            .inner_join(machines::table)
            .inner_join(users::table)
            .filter(id.eq(res_id))
            .first(c)
    }

    ///
    /// Mark this reservation as "ended".
    ///
    /// The only way to mark a reservation as actually concluded (as opposed to scheduled for
    /// completion) is to mark it as completed right now.
    ///
    pub fn end(self, c: &Connection) -> DieselResult<Reservation> {
        diesel::update(&self)
            .set(reservations::actual_end.eq(Some(Utc::now())))
            .get_result::<Reservation>(c)
    }

    pub fn id(&self) -> i32 {
        self.id
    }

    pub fn start(&self) -> DateTime<Utc> {
        self.scheduled_start
    }

    pub fn finish(&self) -> Option<DateTime<Utc>> {
        match (self.scheduled_end, self.actual_end) {
            (Some(s), _) => Some(s),
            (None, Some(a)) => Some(a),
            (None, None) => None,
        }
    }
}

#[derive(Debug, Insertable)]
#[table_name = "reservations"]
pub struct ReservationBuilder {
    user_id: i32,
    machine_id: i32,
    scheduled_start: DateTime<Utc>,
    scheduled_end: Option<DateTime<Utc>>,
    actual_end: Option<DateTime<Utc>>,
    pxe_path: Option<String>,
    nfs_root: Option<String>,
}

impl ReservationBuilder {
    pub fn new(user: &User, machine: &Machine, start: DateTime<Utc>) -> ReservationBuilder {
        ReservationBuilder {
            user_id: user.id,
            machine_id: machine.id,
            scheduled_start: start,
            scheduled_end: None,
            actual_end: None,
            pxe_path: None,
            nfs_root: None,
        }
    }

    pub fn end(&mut self, time: DateTime<Utc>) -> &mut ReservationBuilder {
        self.scheduled_end = Some(time);
        self
    }

    pub fn pxe(&mut self, path: String) -> &mut ReservationBuilder {
        self.pxe_path = Some(path);
        self
    }

    pub fn nfs(&mut self, path: String) -> &mut ReservationBuilder {
        self.nfs_root = Some(path);
        self
    }

    pub fn insert(self, conn: &Connection) -> DieselResult<Reservation> {
        insert_into(reservations::table)
            .values(&self)
            .get_result(conn)
    }
}
