use chrono::UTC;
use chrono::datetime::DateTime;
use db::schema::*;
use diesel;
use diesel::*;
use diesel::pg::PgConnection as Connection;

type DieselResult<T> = Result<T, diesel::result::Error>;


#[derive(Associations, Debug, Identifiable, Queryable)]
#[has_many(reservations)]
pub struct User {
    pub id: i32,
    pub username: String,
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
}

impl User {
    pub fn all(c: &Connection) -> DieselResult<Vec<User>> {
        use self::users::dsl::*;
        users.order(username).load(c)
    }

    pub fn with_email(address: &str, c: &Connection) -> DieselResult<User> {
        use self::users::dsl::*;
        users.filter(email.eq(address)).first(c)
    }

    pub fn with_username(uname: &str, c: &Connection) -> DieselResult<User> {
        use self::users::dsl::*;
        users.filter(username.eq(uname)).first(c)
    }

    pub fn inhabits_role(&self, role: &Role, c: &Connection) -> DieselResult<bool> {
        use self::role_assignments::dsl::*;
        role_assignments.inner_join(roles::table)
            .filter(user_id.eq(self.id))
            .filter(role_id.eq(role.id))
            .count()
            .get_result::<i64>(c)
            .map(|count| count > 0)
    }

    pub fn roles(&self, c: &Connection) -> DieselResult<Vec<Role>> {
        use self::role_assignments::dsl::*;
        role_assignments.inner_join(roles::table)
            .filter(user_id.eq(self.id))
            .load(c)
            .map(|roles: Vec<(RoleAssignment, Role)>|
                roles.into_iter()
                    .map(|(_, r)| r)
                    .collect())
    }

    pub fn can_alter_users(&self, c: &Connection) -> DieselResult<bool> {
        self.has_role(c, |ref role| role.can_alter_users)
    }

    pub fn can_view_users(&self, c: &Connection) -> DieselResult<bool> {
        self.has_role(c, |ref role| role.can_view_users)
    }

    /// Does any of this user's roles satisfy a predicate?
    fn has_role<Pred>(&self, c: &Connection, predicate: Pred) -> DieselResult<bool>
        where Pred: Fn(&Role) -> bool
    {
        use self::role_assignments::dsl::*;
        role_assignments.inner_join(roles::table)
            .filter(user_id.eq(self.id))
            .load(c)
            .map(|roles: Vec<(RoleAssignment, Role)>|
                 roles.into_iter().any(|(_, r)| predicate(&r)))
    }
}


#[derive(Associations, Debug, Identifiable, Queryable)]
#[has_many(role_assignments)]
pub struct Role {
    pub id: i32,
    pub name: String,
    pub can_view_users: bool,
    pub can_alter_users: bool,
}


#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(Role)]
#[belongs_to(User)]
pub struct RoleAssignment {
    pub id: i32,
    pub user_id: i32,
    pub role_id: i32,
}

#[derive(Associations, Debug, Identifiable, Queryable)]
#[has_many(disks)]
#[has_many(nics)]
#[has_many(reservations)]
pub struct Machine {
    pub id: i32,
    pub name: String,
    pub arch: String,
    pub microarch: String,
    pub cores: i32,
    pub memory_gb: i32,
}

impl Machine {
    pub fn all(c: &Connection) -> DieselResult<Vec<Machine>> {
        use self::machines::dsl::*;
        machines.order(name).load(c)
    }

    pub fn with_name(machine_name: &str, c: &Connection) -> DieselResult<Machine> {
        use self::machines::dsl::*;
        machines.filter(name.eq(machine_name)).first(c)
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

#[derive(Associations, Debug, Identifiable, Queryable)]
#[belongs_to(Machine)]
#[belongs_to(User)]
pub struct Reservation {
    pub id: i32,
    pub user_id: i32,
    pub machine_id: i32,
    pub scheduled_start: DateTime<UTC>,
    pub scheduled_end: Option<DateTime<UTC>>,
    pub actual_end: Option<DateTime<UTC>>,
    pub pxe_path: Option<String>,
    pub nfs_root: Option<String>,
}

impl Reservation {
    pub fn start(&self) -> DateTime<UTC> {
        self.scheduled_start
    }

    pub fn finish(&self) -> Option<DateTime<UTC>> {
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
    scheduled_start: DateTime<UTC>,
    scheduled_end: Option<DateTime<UTC>>,
    actual_end: Option<DateTime<UTC>>,
    pxe_path: Option<String>,
    nfs_root: Option<String>,
}

impl ReservationBuilder {
    pub fn new(user: &User, machine: &Machine, start: DateTime<UTC>) -> ReservationBuilder {
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

    pub fn end(&mut self, time: DateTime<UTC>) -> &mut ReservationBuilder {
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
        insert(&self).into(reservations::table)
            .get_result(conn)
    }
}
