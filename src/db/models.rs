use chrono::UTC;
use chrono::datetime::DateTime;
use db::schema::*;


#[derive(Associations, Identifiable, Queryable)]
#[has_many(reservations)]
pub struct User {
    pub id: i32,
    pub username: String,
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
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
