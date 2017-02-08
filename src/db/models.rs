use chrono::UTC;
use chrono::datetime::DateTime;
use db::schema::*;
use diesel::*;
use diesel::pg::PgConnection as Connection;


#[derive(Associations, Identifiable, Queryable)]
#[has_many(reservations)]
pub struct User {
    pub id: i32,
    pub username: String,
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
}

impl User {
    pub fn all(c: &Connection) -> Result<Vec<User>, result::Error> {
        use self::users::dsl::*;
        users.order(username).load(c)
    }

    pub fn with_username(uname: &str, c: &Connection) -> Result<User, result::Error> {
        use self::users::dsl::*;
        users.filter(username.eq(uname)).first(c)
    }
}

#[derive(Associations, Identifiable, Queryable)]
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
    pub fn all(c: &Connection) -> Result<Vec<Machine>, result::Error> {
        use self::machines::dsl::*;
        machines.order(name).load(c)
    }

    pub fn with_name(machine_name: &str, c: &Connection) -> Result<Machine, result::Error> {
        use self::machines::dsl::*;
        machines.filter(name.eq(machine_name)).first(c)
    }
}

#[derive(Associations, Identifiable, Queryable)]
#[belongs_to(Machine)]
pub struct Disk {
    pub id: i32,
    pub machine_id: i32,
    pub vendor: Option<String>,
    pub model: Option<String>,
    pub capacity_gb: i32,
    pub ssd: bool,
}

#[derive(Associations, Identifiable, Queryable)]
#[belongs_to(Machine)]
pub struct Nic {
    pub id: i32,
    pub machine_id: i32,
    pub vendor: Option<String>,
    pub model: Option<String>,
    pub mac_address: String,
    pub speed_gbps: i32,
}

#[derive(Associations, Identifiable, Queryable)]
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

    pub fn insert(self, conn: &Connection) -> Result<Reservation, result::Error> {
        insert(&self).into(reservations::table)
            .get_result(conn)
    }
}
