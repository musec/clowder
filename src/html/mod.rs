/*
 * Copyright 2016-2018 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use std::collections::HashSet;

use chrono::{DateTime, Utc};
use chrono_humanize::HumanTime;
use db;
use db::models::*;
use diesel;
use hyper;
use marksman_escape::Escape;
use maud::*;
use native_tls;
use super::rocket;
use rocket::{http, request, Catcher, Route};
use rocket::request::{FlashMessage, Form};
use rocket::response::{Flash, Redirect};
use rustc_serialize;
use url;

// We do, in fact, use FromFrom, but only in a rocket-codegen derivation.
#[allow(unused_imports)]
use rocket::request::FromForm;

mod auth;
mod bootstrap;
mod error;
mod github;
mod forms;
mod link;
mod static_files;
mod tables;

use std::env;
use self::auth::AuthContext;
use self::bootstrap::Page;
use self::error::Error;
use self::link::Link;


/// All of the routes that we can handle.
pub fn all_routes() -> Vec<Route> {
    routes! {
        index,
        github_callback, logout,
        machine, machine_create, machines,
        reservation, reservation_create_page, reservation_create,
        reservation_end, reservation_end_confirm, reservations,
        static_files::static_css, static_files::static_images, static_files::static_js,
        user, user_update, users,
    }
}

/// All of our error catchers (which wrap HTTP errors and present nicer UIs for them).
pub fn error_catchers() -> Vec<Catcher> {
    catchers! {
        error::not_found, error::unauthorized, error::forbidden, error::internal_server_error,
    }
}

/// What prefix should we prepend to our links?
///
/// Specified by CLOWDER_PREFIX in environment or .env; defaults to "/".
pub fn route_prefix() -> String {
    env::var("CLOWDER_PREFIX").unwrap_or(String::from("/"))
}

/// Escape a string to make it suitable for HTML form input.
pub fn escape(dangerous: &str) -> String {
    String::from_utf8(Escape::new(dangerous.bytes()).collect())
        .unwrap_or(String::from("&lt;error&gt;"))
}


///
/// Create a normal (i.e., non-error) page to be augmented with content.
///
/// ```rust
/// page("my title", &auth)
///     .content(html! { h2 "hello" })
///     .flash(/* ... */)
///     .render()
/// ```
///
fn page<S>(title: S, auth: &AuthContext) -> Page
    where S: Into<String>
{
    let user = &auth.user;
    let route_prefix: &str = &route_prefix();
    let prefix = |s| format!["{}{}", route_prefix, s];

    let mut nav_links = vec![bootstrap::NavItem::link(prefix("machines"), "Machines"),
                             bootstrap::NavItem::link(prefix("reservations"), "Reservations")];

    if let Ok(true) = user.can_alter_users(&auth.conn) {
        nav_links.push(bootstrap::NavItem::link(prefix("users"), "Users"));
    }

    Page::new(title.into())
        .link_prefix(route_prefix)
        .nav(nav_links)
        .user(&user.username, &user.name)
}


#[get("/")]
fn index(auth: AuthContext) -> Result<Page, Error> {
    let machines = FullMachine::all(&auth.conn)?;

    let reservations = Reservation::all(true, &auth.conn)
        ?
        .into_iter()
        .map(|(r, m, u)| (r, Some(m), Some(u)))
        .collect();

    Ok(page("Clowder", &auth).content(html! {
        div.row {
            div class="col-md-6" {
                h4 { "Machine inventory" }
                (tables::MachineTable::new(machines)
                    .show_arch(false)
                    .show_cores(true)
                    .show_freq(false)
                    .show_memory(true)
                    .show_microarch(true)
                    .show_processor_name(false))
            }

            div class="col-md-6" {
                h4 { "Current reservations" }
                (tables::ReservationTable::new(reservations)
                                          .show_machine(true)
                                          .show_user(true)
                                          .show_scheduled_end(true)
                                          .show_scheduled_start(false)
                                          .show_actual_end(false))
            }
        }
    }))
}

#[get("/gh-callback?<code>")]
fn github_callback(code: String, cookies: http::Cookies) -> Result<Redirect, Error> {
    auth::github_callback(code, cookies).map(|_| Redirect::to(route_prefix()))
}

#[get("/logout")]
fn logout(_auth: AuthContext, cookies: http::Cookies) -> Redirect {
    auth::logout(cookies);
    Redirect::to(route_prefix().to_string())
}

#[get("/machine/<machine_name>")]
fn machine(machine_name: String, auth: AuthContext) -> Result<Page, Error> {
    let conn = &auth.conn;

    let m = FullMachine::with_name(&machine_name, conn)?;
    let disks = m.machine().disks(conn)?;
    let nics = m.machine().nics(conn)?;

    Ok(page(format!["Clowder: {}", m.name()], &auth).content(html! {
        div.row { h2 { (m.name()) } }

        div.row {
            div class="col-md-7" {
                dl {
                    dt { "Processor(s)" }
                    dd {
                        ul {
                            li {
                                (Link::from(m.processor()))
                                ": "
                                (Link::from(m.microarchitecture())) " " (m.architecture().name)
                                ", "
                                (m.cores()) " cores, " (m.freq_ghz()) " GHz"
                            }
                        }
                    }

                    dt { "Memory" }
                    dd { (m.memory_gb()) " GiB" }

                    dt { "Disk(s)" }
                    dd {
                        ul {
                            @for ref disk in disks {
                                li { (disk.short_description()) }
                            }
                        }
                    }

                    dt { "NIC(s)" }
                    dd {
                        ul {
                            @for ref nic in nics {
                                li { (nic.short_description()) }
                            }
                        }
                    }
                }

                p {
                    a href={ (route_prefix()) "reservation/create/?machine=" (m.name()) } {
                        "Reserve this machine"
                    }
                }
            }

            div class="col-md-5" {
                h3 { "Reservations" }

                table.table.table-responsive {
                    (tables::TableHeader::new(&[ "", "User", "Started", "Ends" ]))

                    tbody {
                        @for (ref r, ref u) in Reservation::for_machine(&m.machine(), conn)? {
                            tr {
                                td { (Link::from(r)) }
                                td { (Link::from(u)) }
                                td { (HumanTime::from(r.start())) }
                                td {
                                    (r.scheduled_end.map(|e| HumanTime::from(e).to_string())
                                                    .unwrap_or(String::new()))
                                }
                            }
                        }
                    }
                }
            }
        }
    }))
}

#[derive(Debug, FromForm)]
struct NewMachineForm {
    name: String,
    processor: i32,
    memory_gb: i32,
}

#[post("/machine/create", data = "<form>")]
fn machine_create(form: Form<NewMachineForm>, auth: AuthContext) -> Result<Redirect, Error> {
    MachineBuilder::new(form.name.clone())
        .processor(&Processor::get(form.processor, &auth.conn)?)
        .memory_gb(form.memory_gb)
        .insert(&auth.conn)
        .map(|m| Redirect::to(format!["{}machine/{}", route_prefix(), m.name]))
        .map_err(Error::DatabaseError)
}

#[get("/machines")]
fn machines(auth: AuthContext) -> Result<Page, Error> {
    let machine_creator = auth.user.can_create_machines(&auth.conn)?;
    let processor_options = Processor::all(&auth.conn)
        ?
        .iter()
        .map(|p| forms::SelectOption::new(p.id.to_string(), p.name.clone()))
        .collect::<Vec<_>>();

    FullMachine::all(&auth.conn)
        .map_err(Error::DatabaseError)
        .map(|machines| tables::MachineTable::new(machines))
        .map(|table| {
            html! {
                h2 { "Current inventory" }
                (table)

                @if machine_creator {
                    h2 { "Add new machine" }

                    form action={ (route_prefix()) "machine/create" } method="post" {
                        table {
                            tr {
                                th { "Name" }
                                td { (forms::Input::new("name")) }
                            }
                            tr {
                                th { "Processor" }
                                td {
                                    (forms::Select::new("processor")
                                                   .set_options(processor_options))
                                }
                            }
                            tr {
                                th { "Memory" }
                                td { (forms::Input::new("memory_gb")) " GiB" }
                            }
                            tr {
                                th /
                                td { (forms::SubmitButton::new().label("Add to inventory")) }
                            }
                        }
                    }
                }
            }
        })
        .map(|table| page("Clowder: Machines", &auth).content(table))
}

#[get("/reservation/<id>")]
fn reservation(id: i32, auth: AuthContext, flash: Option<FlashMessage>) -> Result<Page, Error> {
    let (r, machine, user) = Reservation::get(id, &auth.conn)?;

    let can_end = match (r.scheduled_start, r.actual_end) {
        (s, None) if s <= Utc::now() => true,
        (_, _) => false,
    };

    Ok(page(format!["Clowder: reservation {}", r.id], &auth).flash(flash).content(html! {
        h2 { "Reservation " (r.id) }

        table.lefty {
            tr { th { "User" }       td { (Link::from(&user)) } }
            tr { th { "Machine" }    td { (Link::from(&machine)) } }
            tr { th { "Starts" }     td { (r.scheduled_start) } }
            tr {
                th { "Ends" }
                td {
                    (match r.scheduled_end {
                        Some(d) => d.to_string(),
                        None => String::new(),
                    })
                }
            }
            tr {
                th { "Ended" }
                td {
                    (match r.actual_end {
                        Some(d) => d.to_string(),
                        None => String::new(),
                    })
                }
            }
            tr {
                th { "NFS root" }
                td {
                    (match r.nfs_root {
                        Some(r) => r,
                        None => String::new(),
                    })
                }
            }
            tr {
                th { "PXE path" }
                td {
                    (match r.pxe_path {
                        Some(p) => p,
                        None => String::new(),
                    })
                }
            }
            @if can_end {
                tr {
                    th {}
                    td {
                        form action={ "end/" (r.id) } method="get" {
                             input type="submit" value="End reservation" /
                        }
                    }
                }
            }
        }
    }))
}

#[derive(Debug, FromForm)]
struct ReservationForm {
    user: String,
    machine: String,
    dates: String,
    pxe: String,
    nfs: String,
}

#[post("/reservation/create", data = "<res>")]
fn reservation_create(res: Form<ReservationForm>, auth: AuthContext) -> Result<Redirect, Error> {
    let user = User::with_username(&res.user, &auth.conn)?;
    let machine = Machine::with_name(&res.machine, &auth.conn)?;

    let dates: Vec<&str> = res.dates.split(" - ").collect();
    if dates.len() != 2 {
        return Err(Error::BadRequest(format!["expected two dates, not '{:?}'", dates]));
    }

    let start = DateTime::parse_from_str(dates[0], "%H:%M%:z %e %b %Y")?;
    let end = DateTime::parse_from_str(dates[1], "%H:%M%:z %e %b %Y")?;

    let mut rb = ReservationBuilder::new(&user, &machine, start.with_timezone(&Utc));
    rb.end(end.with_timezone(&Utc));

    if res.pxe.len() > 0 {
        rb.pxe(res.pxe.clone());
    }
    if res.pxe.len() > 0 {
        rb.nfs(res.nfs.clone());
    }

    rb.insert(&auth.conn)
        .map(|r| Redirect::to(format!["{}reservation/{}", route_prefix(), r.id]))
        .map_err(Error::DatabaseError)
}

#[get("/reservation/create?<machine>")]
fn reservation_create_page(machine: Option<String>, auth: AuthContext)
    -> Result<Page, Error>
{
    let users = User::all(&auth.conn)?;
    let user_options = users.iter()
        .map(|ref u| {
            forms::SelectOption::new(u.username.clone(), u.name.clone())
                .selected(u.username == auth.user.username)
        })
        .collect::<Vec<_>>();

    let machines = Machine::all(&auth.conn)?;
    let machine_options = machines.iter()
        .map(|ref m| {
            forms::SelectOption::new(m.name.clone(), m.name.clone())
                .selected(if let &Some(ref name) = &machine { name == &m.name } else { false })
        })
        .collect::<Vec<_>>();

    Ok(page("Create reservation", &auth).content(html! {
        h2 { "Reserve a machine" }

        form action="." method="post" {
            table {
                tr {
                    th { "User" }
                    td { (forms::Select::new("user").set_options(user_options)) }
                }
                tr {
                    th { "Machine" }
                    td { (forms::Select::new("machine").set_options(machine_options)) }
                }
                tr {
                    th { "Dates" }
                    td { (forms::Input::new("dates").class("daterange").size(45)) }
                }
                tr {
                    th { "PXE loader" }
                    td { (forms::Input::new("pxe").size(45)) }
                }
                tr {
                    th { "NFS root" }
                    td { (forms::Input::new("nfs").size(45)) }
                }
                tr {
                    th /
                    td { (forms::SubmitButton::new().label("Reserve")) }
                }
            }
        }
    }))
}

#[get("/reservation/end/<id>")]
fn reservation_end(id: i32, auth: AuthContext) -> Result<Page, Error> {
    let (r, machine, user) = Reservation::get(id, &auth.conn)?;

    Ok(page(format!["Clowder: end reservation {}", r.id], &auth).content(html! {
        h2 { "End reservation" }

        (bootstrap::callout("warning", "Are you sure you want to end this reservation?",
            html! {
                table {
                    tr { th { "User" }       td { (Link::from(&user)) } }
                    tr { th { "Machine" }    td { (Link::from(&machine)) } }
                    tr { th { "Starts" }     td { (r.scheduled_start) } }
                    tr {
                        th { "Ends" }
                        td {
                            (match r.scheduled_end {
                                Some(d) => d.to_string(),
                                _ => String::new()
                            })
                        }
                    }
                    tr {
                        th { "Ended" }
                        td {
                            (match r.actual_end {
                                Some(d) => d.to_string(),
                                None => String::new()
                            })
                        }
                    }
                    tr {
                        th { "NFS root" }
                        td { (match r.nfs_root { Some(r) => r, None => String::new() }) }
                    }
                    tr {
                        th { "PXE path" }
                        td { (match r.pxe_path { Some(p) => p, None => String::new() }) }
                    }
                    tr {
                        td colspan="2" {
                            form action={ "confirm/" (r.id) } method="get" {
                                input type="submit" value="End reservation" /
                            }
                        }
                    }
                }
            }))
    }))
}

#[get("/reservation/end/confirm/<res_id>")]
fn reservation_end_confirm(res_id: i32, auth: AuthContext) -> Result<Flash<Redirect>, Error> {
    Reservation::get(res_id, &auth.conn)
        .and_then(|(r, _, _)| r.end(&auth.conn))
        .map_err(Error::DatabaseError)
        .map(|r| {
            Flash::new(Redirect::to(format!["{}reservation/{}", route_prefix(), r.id()]),
                       "info",
                       format!["Ended reservation {}", r.id()])
        })
}

#[get("/reservations")]
fn reservations(auth: AuthContext) -> Result<Page, Error> {
    let reservations = Reservation::all(false, &auth.conn)
        ?
        .into_iter()
        .map(|(r, m, u)| (r, Some(m), Some(u)))
        .collect();

    Ok(page("Clowder: Reservations", &auth).content(tables::ReservationTable::new(reservations)))
}

#[get("/user/<name>")]
fn user(name: String, auth: AuthContext) -> Result<Page, Error> {
    let user = User::with_username(&name, &auth.conn)?;
    let superuser = auth.user.can_alter_users(&auth.conn)?;

    let name = user.name.as_str();
    let myself = user.id == auth.user.id;
    let writable = myself || superuser;

    let emails = user.emails(&auth.conn)?;

    let roles = Role::all(&auth.conn)
        ?
        .into_iter()
        .map(|role| {
            let inhabited = user.inhabits_role(&role, &auth.conn).unwrap_or(false);
            (role.name, inhabited)
        });

    let reservations = Reservation::for_user(&user, &auth.conn)
        ?
        .into_iter()
        .map(|(r, m)| (r, Some(m), None))
        .collect();

    Ok(page(name, &auth).content(html! {
        h2 { (name) }

        div.row {
            div class="col-md-6" {
                form action={ (route_prefix()) "user/update/" (user.username) } method="post" {
                    table.table.table-responsive {
                        tbody {
                            tr { th { "Username" } td { (user.username) } }
                            tr {
                                th { "Name" }
                                td {
                                    (forms::Input::new("name")
                                                  .value(user.name.clone())
                                                  .size(18)
                                                  .writable(writable))
                                }
                            }
                            tr {
                                th { "Email" }
                                td {
                                    @for address in emails {
                                        (address)
                                        br {}
                                    }
                                }
                            }
                            tr {
                                th { "Roles" }
                                td {
                                    @if superuser {
                                        (forms::Select::new("roles")
                                            .multiple(true)
                                            .set_options(
                                                roles
                                                    .map(|(name, inhabited)| {
                                                        forms::SelectOption::new(&*name, &*name)
                                                            .selected(inhabited)
                                                    })
                                                    .collect()
                                            ))
                                    } @else {
                                        ul.list-unstyled {
                                            @for (name, inhabited) in roles {
                                                li {
                                                    (name)
                                                    @if inhabited {
                                                        " " i.fa.fa-check aria-hidden="true" {}
                                                    }
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                            tr {
                                td colspan="2" {
                                    input.btn.btn-block.btn-warning type="submit"
                                        value="Update user details" /
                                }
                            }
                        }
                    }
                }
            }

            div class="col-md-6" {
                (tables::ReservationTable::new(reservations).show_user(false))
            }
        }
    }))
}


#[get("/users")]
fn users(auth: AuthContext) -> Result<Page, Error> {
    let conn = &auth.conn;

    let can_view = auth.user.can_alter_users(conn).unwrap_or(false);
    let can_edit = auth.user.can_alter_users(conn).unwrap_or(false);

    if !can_view {
        return Err(Error::NotAuthorized(format!["User '{}' cannot view/alter other users",
                                                auth.user.username]));
    }

    let users = User::all(conn)?;
    let roles = Role::all(conn)?;

    Ok(page("Users", &auth).content(html! {
        h2 { ("Users") }

        table.table.table-responsive {
            thead.thead-default {
                tr {
                    th {}
                    th { "Username" }
                    th { "Name" }
                    th { "Email" }
                    th { "Roles" }
                    th {}
                }
            }

            tbody {
                @for ref user in &users {
                    form action={ (route_prefix()) "user/update/" (user.username) } method="post" {
                        tr {
                            th { (user.id) }
                            td { (user.username.clone()) }
                            td {
                                (forms::Input::new("name")
                                       .value(user.name.clone())
                                       .size(15)
                                       .writable(can_edit))
                            }
                            td { (user.emails(conn)?.into_iter().collect::<Vec<_>>().join(" ")) }
                            td {
                                (forms::Select::new("roles")
                                    .set_options(
                                        roles.iter()
                                            .map(|ref role| {
                                                let name = &*role.name;
                                                let inhabited = user.inhabits_role(role, conn)
                                                                    .unwrap_or(false);

                                                forms::SelectOption::new(name, name)
                                                                    .selected(inhabited)
                                            })
                                            .collect()
                                    )
                                    .multiple(true))
                            }
                            td { (forms::SubmitButton::new().label("Update")) }
                        }
                    }
                }
            }
        }
    }))
}

struct UserUpdate {
    name: String,
    emails: HashSet<String>,
    roles: HashSet<String>,
}

// TODO: use #[derive(FromForm)] once SergioBenitez/Rocket#205 is resolved
impl<'f> request::FromForm<'f> for UserUpdate {
    type Error = error::Error;

    fn from_form(form_items: &mut request::FormItems<'f>, _: bool) -> Result<Self, Self::Error> {
        let mut update = UserUpdate {
            name: String::new(),
            emails: HashSet::new(),
            roles: HashSet::new(),
        };

        for (key, value) in form_items.map(|i| i.key_value_decoded()) {
            match &*key {
                "name" => update.name = value,
                "emails" => {
                    update.emails.insert(value);
                }
                "roles" => {
                    update.roles.insert(value);
                }
                _ => {
                    return Err(Error::InvalidData(format!["invalid form data name: '{}'", key]));
                }
            }
        }

        Ok(update)
    }
}

#[post("/user/update/<who>", data = "<form>")]
fn user_update(who: String,
               auth: AuthContext,
               form: Form<UserUpdate>)
               -> Result<Flash<Redirect>, Error> {

    let conn = &auth.conn;

    let mut user =
        User::with_username(&who, conn)
             .map_err(|err| Error::BadRequest(format!["No such user: '{}' ({})", who, err]))?;

    let superuser = auth.user.can_alter_users(conn).unwrap_or(false);

    if !(user.id == auth.user.id || superuser) {
        return Err(Error::NotAuthorized(String::from("update other users' details")));
    }

    // Has the user requested a name change?
    if user.name != form.name {
        user = user.change_name(form.name.clone(), conn)?;
    }

    // Only admin users can (currently) modify email addresses, since they are almost akin to
    // login credentials. We will revisit this in the future.
    if superuser {
        user.set_emails(&form.emails, conn)?;
    }

    // Only admin users can change users' roles.
    if superuser {
        user.set_roles(&form.roles, conn)?;
    }

    Ok(Flash::new(Redirect::to(format!["{}user/{}", route_prefix(), user.username]),
                  "info",
                  format!["Updated {}'s details", user.username]))
}
