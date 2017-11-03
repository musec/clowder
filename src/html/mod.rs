use std::collections::HashSet;

use chrono::{DateTime,Utc};
use chrono_humanize::HumanTime;
use ::db;
use db::models::*;
use ::diesel;
use diesel::pg::PgConnection as Connection;
use hyper;
use marksman_escape::Escape;
use maud::*;
use native_tls;
use super::rocket;
use rocket::*;
use rocket::request::{FlashMessage, Form, FromFormValue};
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
use self::error::Error;
use self::link::Link;


/// Contextual information about the current page rendering.
pub struct Context {
    /// Who is viewing the page
    user: User,

    /// Database connection for additional queries
    conn: Connection,
}

impl<'a, 'r> request::FromRequest<'a, 'r> for Context {
    type Error = error::Error;

    fn from_request(req: &'a Request<'r>)
            -> request::Outcome<Context, Self::Error> {

        let conn = db::establish_connection();
        let user = auth::authenticate(req.cookies(), &conn);

        match user {
            Ok(u) => Outcome::Success(Context { user: u, conn: conn }),
            Err(e) => {
                let failure = match e {
                    Error::AuthRequired => (http::Status::Unauthorized, e),
                    _ => (http::Status::InternalServerError, e),
                };

                Outcome::Failure(failure)
            },
        }
    }
}


/// All of the routes that we can handle.
pub fn all_routes() -> Vec<Route> {
    routes! {
        index,
        github_callback, logout,
        machine, machines,
        reservation, reservation_create_page, reservation_create,
        reservation_end, reservation_end_confirm, reservations,
        static_files::static_css, static_files::static_images, static_files::static_js,
        user, user_update, users,
    }
}

/// All of our error catchers (which wrap HTTP errors and present nicer UIs for them).
pub fn error_catchers() -> Vec<Catcher> {
    errors! {
        error::not_found, error::unauthorized, error::internal_server_error,
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


/// Render a normal (i.e., non-error) page of content.
pub fn render<S,M>(title: S, ctx: &Context, flash: Option<FlashMessage>, content: M) -> Markup
    where S: Into<String>, M: Into<Markup>
{
    let user = &ctx.user;
    let route_prefix = route_prefix();
    let prefix = |s| format!["{}{}", route_prefix, s];

    let mut nav_links = vec![
        bootstrap::NavItem::link(prefix("machines"), "Machines"),
        bootstrap::NavItem::link(prefix("reservations"), "Reservations"),
    ];

    if let Ok(true) = user.can_alter_users(&ctx.conn) {
        nav_links.push(bootstrap::NavItem::link(prefix("users"), "Users"));
    }

    bootstrap::Page::new(title)
                    .content(content.into())
                    .flash(flash)
                    .link_prefix(&route_prefix as &str)
                    .nav(nav_links)
                    .user(&user.username, &user.name)
                    .render()
}


#[get("/")]
fn index(ctx: Context) -> Result<Markup, Error> {
    let machines = FullMachine::all(&ctx.conn)?;

    // TODO: use multiple joins once Diesel supports it
    let reservations = Reservation::all(true, &ctx.conn)?
                                   .into_iter()
                                   .map(|(r,m)| (r, Some(m), None))
                                   .collect();

    Ok(render("Clowder", &ctx, None, html! {
        div.row {
            div class="col-md-6" {
                h4 "Machine inventory"
                (tables::MachineTable::new(machines)
                    .show_arch(false)
                    .show_cores(true)
                    .show_freq(false)
                    .show_memory(true)
                    .show_microarch(true)
                    .show_processor_name(false))
            }

            div class="col-md-6" {
                h4 "Current reservations"
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

#[derive(FromForm)]
struct GithubCallbackData {
    code: String,
    state: Option<String>,
}

#[get("/gh-callback?<query>")]
fn github_callback(query: GithubCallbackData, cookies: http::Cookies) -> Result<Redirect, Error> {
    let mut gh = github::Client::new(env::var("CLOWDER_GH_CLIENT_ID")?)?
        .set_secret(env::var("CLOWDER_GH_CLIENT_SECRET")?)
        .set_oauth_code(query.code);
        ;

    let email = gh.user().map(|u| u.email().to_string())?;
    let conn = db::establish_connection();

    User::with_email(&email, &conn)
        .map_err(|_| Error::AuthError(format!["Unknown user: {}", email]))
        .map(|user| auth::set_user_cookie(cookies, user.username))
        .map(|_| Redirect::to(&route_prefix()))
}

#[get("/logout")]
fn logout(_ctx: Context, cookies: http::Cookies) -> Redirect {
    auth::logout(cookies);
    Redirect::to(&route_prefix())
}

#[get("/machine/<machine_name>")]
fn machine(machine_name: String, ctx: Context) -> Result<Markup, Error> {
    let conn = &ctx.conn;

    FullMachine::with_name(&machine_name, conn)
        .map_err(Error::DatabaseError)
        .and_then(|m| Ok(render(format!["Clowder: {}", m.name()], &ctx, None, html! {
        div.row h2 (m.name())

        div.row {
            div class="col-md-7" {
                p {
                    (m.architecture().name) " (" (m.microarchitecture().name) "), "
                    (m.cores()) " cores, " (m.memory_gb()) " GiB RAM"
                }

                p a href={ (route_prefix()) "reservation/create/?machine=" (m.name()) }
                    "Reserve this machine"
            }

            div class="col-md-5" {
                h3 "Reservations"

                table.table.table-responsive {
                    (tables::TableHeader::from_str(
                        &[ "", "User", "Started", "Ends" ]))

                    tbody {
                        @for (ref r, ref u) in Reservation::for_machine(&m.machine(), conn)? {
                            tr {
                                td (Link::from(r))
                                td (Link::from(u))
                                td (HumanTime::from(r.start()))
                                td (r.scheduled_end.map(|e| HumanTime::from(e).to_string())
                                                   .unwrap_or(String::new()))
                            }
                        }
                    }
                }
            }
        }
    })))
}

#[get("/machines")]
fn machines(ctx: Context) -> Result<Markup, Error> {
    FullMachine::all(&ctx.conn)
        .map_err(Error::DatabaseError)
        .map(|machines| tables::MachineTable::new(machines))
        .map(|table| render("Clowder: Machines", &ctx, None, table.render()))
}

#[get("/reservation/<id>")]
fn reservation(id: i32, ctx: Context, flash: Option<FlashMessage>) -> Result<Markup, Error> {
    let (r, machine, user) = Reservation::get(id, &ctx.conn)?;

    let can_end = match (r.scheduled_start, r.actual_end) {
        (s, None) if s <= Utc::now() => true,
        (_, _) => false,
    };

    Ok(render(format!["Clowder: reservation {}", r.id], &ctx, flash, html! {
        h2 { "Reservation " (r.id) }

        table.lefty {
            tr { th "User"       td (Link::from(&user)) }
            tr { th "Machine"    td (Link::from(&machine)) }
            tr { th "Starts"     td (r.scheduled_start) }
            tr {
                th "Ends"
                td (match r.scheduled_end {
                    Some(d) => d.to_string(),
                    None => String::new(),
                })
            }
            tr {
                th "Ended"
                td (match r.actual_end {
                    Some(d) => d.to_string(),
                    None => String::new(),
                })
            }
            tr {
                th "NFS root"
                td (match r.nfs_root {
                    Some(r) => r,
                    None => String::new(),
                })
            }
            tr {
                th "PXE path"
                td (match r.pxe_path {
                    Some(p) => p,
                    None => String::new(),
                })
            }
            @if can_end {
                tr {
                    th ""
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

#[post("/reservation/create", data = "<form>")]
fn reservation_create(form: Form<ReservationForm>, ctx: Context) -> Result<Redirect, Error> {
    let res = form.get();

    let user = try![User::with_username(&res.user, &ctx.conn)];
    let machine = try![Machine::with_name(&res.machine, &ctx.conn)];

    let dates: Vec<&str> = res.dates.split(" - ").collect();
    if dates.len() != 2 {
        return Err(Error::BadRequest(format!["expected two dates, not '{:?}'", dates]));
    }

    let start = try![DateTime::parse_from_str(dates[0], "%H:%M%:z %e %b %Y")];
    let end = try![DateTime::parse_from_str(dates[1], "%H:%M%:z %e %b %Y")];

    let mut rb = ReservationBuilder::new(&user, &machine, start.with_timezone(&Utc));
    rb.end(end.with_timezone(&Utc));

    if res.pxe.len() > 0 { rb.pxe(res.pxe.clone()); }
    if res.pxe.len() > 0 { rb.nfs(res.nfs.clone()); }

    rb.insert(&ctx.conn)
        .map(|r| Redirect::to(&format!["{}reservation/{}", route_prefix(), r.id]))
        .map_err(Error::DatabaseError)
}

#[derive(Debug, FromForm)]
struct ReservationQuery {
    user: Option<String>,
    machine: Option<String>,
}

#[get("/reservation/create?<res>")]
fn reservation_create_page(res: ReservationQuery, ctx: Context) -> Result<Markup, Error> {
    let users = try![User::all(&ctx.conn)];
    let user_options = users.iter()
        .map(|ref u| forms::SelectOption::new(u.username.clone(), u.name.clone())
                                         .selected(u.username == ctx.user.username))
        .collect::<Vec<_>>()
        ;

    let machines = try![Machine::all(&ctx.conn)];
    let machine_options = machines.iter()
        .map(|ref m| forms::SelectOption::new(m.name.clone(), m.name.clone())
                                         .selected(res.machine.as_ref()
                                                              .map(|n| n == &m.name)
                                                              .unwrap_or(false)))
        .collect::<Vec<_>>()
        ;

    Ok(render("Create reservation", &ctx, None, html! {
        h2 "Reserve a machine"

        form action="." method="post" {
            table {
                tr {
                    th "User"
                    td (forms::Select::new("user").set_options(user_options))
                }
                tr {
                    th "Machine"
                    td (forms::Select::new("machine").set_options(machine_options))
                }
                tr {
                    th "Dates"
                    td (forms::Input::new("dates").class("daterange").size(45))
                }
                tr {
                    th "PXE loader"
                    td (forms::Input::new("pxe").size(45))
                }
                tr {
                    th "NFS root"
                    td (forms::Input::new("nfs").size(45))
                }
                tr {
                    th /
                    td (forms::SubmitButton::new().label("Reserve"))
                }
            }
        }
    }))
}

#[get("/reservation/end/<id>")]
fn reservation_end(id: i32, ctx: Context) -> Result<Markup, Error> {
    let (r, machine, user) = Reservation::get(id, &ctx.conn)?;

    Ok(render(format!["Clowder: end reservation {}", r.id], &ctx, None, html! {
        h2 "End reservation"

        (bootstrap::callout("warning", "Are you sure you want to end this reservation?",
            html! {
                table {
                    tr { th "User"       td (Link::from(&user)) }
                    tr { th "Machine"    td (Link::from(&machine)) }
                    tr { th "Starts"     td (r.scheduled_start) }
                    tr {
                        th "Ends"
                        td (match r.scheduled_end { Some(d) => d.to_string(), _ => String::new() })
                    }
                    tr {
                        th "Ended"
                        td (match r.actual_end { Some(d) => d.to_string(), None => String::new() })
                    }
                    tr {
                        th "NFS root"
                        td (match r.nfs_root { Some(r) => r, None => String::new() })
                    }
                    tr {
                        th "PXE path"
                        td (match r.pxe_path { Some(p) => p, None => String::new(),
                        })
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
fn reservation_end_confirm(res_id: i32, ctx: Context) -> Result<Flash<Redirect>, Error> {
    Reservation::get(res_id, &ctx.conn)
        .and_then(|(r,_,_)| r.end(&ctx.conn))
        .map_err(Error::DatabaseError)
        .map(|r| Flash::new(Redirect::to(&format!["{}reservation/{}", route_prefix(), r.id()]),
                            "info", &format!["Ended reservation {}", r.id()]))
}

#[get("/reservations")]
fn reservations(ctx: Context) -> Result<Markup, Error> {
    // TODO: use multiple joins once Diesel supports it
    let reservations = Reservation::all(false, &ctx.conn)?
                                   .into_iter()
                                   .map(|(r, m)| (r, Some(m), None))
                                   .collect();

    Ok(render("Clowder: Reservations", &ctx, None,
              tables::ReservationTable::new(reservations).render()))
}

#[get("/user/<name>")]
fn user(name: String, ctx: Context) -> Result<Markup, Error> {
    let user = try![User::with_username(&name, &ctx.conn)];
    let superuser = ctx.user.can_alter_users(&ctx.conn)?;

    let name = user.name.as_str();
    let myself = user.id == ctx.user.id;
    let writable = myself || superuser;

    let emails = user.emails(&ctx.conn)?;

    let roles =
        Role::all(&ctx.conn)?
            .into_iter()
            .map(|role| {
                let inhabited = user.inhabits_role(&role, &ctx.conn).unwrap_or(false);
                (role.name, inhabited)
            })
            ;

    let reservations = Reservation::for_user(&user, &ctx.conn)?
                                   .into_iter()
                                   .map(|(r,m)| (r, Some(m), None))
                                   .collect();

    Ok(render(name, &ctx, None, html! {
        h2 (name)

        div.row {
            div class="col-md-6" {
                form action={ (route_prefix()) "user/update/" (user.username) } method="post" {
                    table.table.table-responsive {
                        tbody {
                            tr { th "Username" td (user.username) }
                            tr { th "Name"
                                td (forms::Input::new("name")
                                                 .value(user.name.clone())
                                                 .size(18)
                                                 .writable(writable))
                            }
                            tr { th "Email"
                                td
                                    @for address in emails {
                                        (address)
                                        br {}
                                    }
                            }
                            tr {
                                th "Roles"
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
                                        ul.list-unstyled
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
                            tr td colspan="2"
                                input.btn.btn-block.btn-warning type="submit"
                                    value="Update user details" /
                        }
                    }
                }
            }

            div class="col-md-6"
                (tables::ReservationTable::new(reservations).show_user(false))
        }
    }))
}


#[get("/users")]
fn users(ctx: Context) -> Result<Markup, Error> {
    let conn = &ctx.conn;

    let can_view = ctx.user.can_alter_users(conn).unwrap_or(false);
    let can_edit = ctx.user.can_alter_users(conn).unwrap_or(false);

    if !can_view {
        return Err(Error::NotAuthorized(
            format!["User '{}' not permitted to view/alter other users", ctx.user.username]));
    }

    let users = User::all(conn)?;
    let roles = Role::all(conn)?;

    Ok(render("Users", &ctx, None, html! {
        h2 ("Users")

        table.table.table-responsive {
            thead.thead-default tr {
                th {}
                th "Username"
                th "Name"
                th "Email"
                th "Roles"
                th {}
            }

            tbody {
                @for ref user in &users {
                    form action={ (route_prefix()) "user/update/" (user.username) } method="post" {
                        tr {
                            th (user.id)
                            td (user.username.clone())
                            td (forms::Input::new("name")
                                    .value(user.name.clone())
                                    .size(15)
                                    .writable(can_edit))
                            td (user.emails(conn)?.into_iter().collect::<Vec<_>>().join(" "))
                            td (forms::Select::new("roles")
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
                            td (forms::SubmitButton::new().label("Update"))
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

        for (k, v) in form_items {
            let key: &str = &*k;
            let value =
                String::from_form_value(v)
                       .map_err(rocket::http::RawStr::as_str)
                       .map_err(String::from)
                       .map_err(Error::InvalidData)?;

            match key {
                "name" => update.name = value,
                "emails" => { update.emails.insert(value); },
                "roles" => { update.roles.insert(value); },
                _ => {
                    return Err(Error::InvalidData(format!["invalid form data name: '{}'", key]));
                },
            }
        }

        Ok(update)
    }
}

#[post("/user/update/<who>", data = "<form>")]
fn user_update(who: String, ctx: Context, form: Form<UserUpdate>)
    -> Result<Flash<Redirect>, Error> {

    let conn = &ctx.conn;

    let mut user = try! {
        User::with_username(&who, conn)
             .map_err(|err| Error::BadRequest(format!["No such user: '{}' ({})", who, err]))
    };

    let superuser = ctx.user.can_alter_users(conn).unwrap_or(false);

    if !(user.id == ctx.user.id || superuser) {
        return Err(Error::NotAuthorized(String::from("update other users' details")))
    }

    let f = form.get();

    // Has the user requested a name change?
    if user.name != f.name {
        user = user.change_name(f.name.clone(), conn)?;
    }

    // Only admin users can (currently) modify email addresses, since they are almost akin to
    // login credentials. We will revisit this in the future.
    if superuser {
        user.set_emails(&f.emails, conn)?;
    }

    // Only admin users can change users' roles.
    if superuser {
        user.set_roles(&f.roles, conn)?;
    }

    Ok(Flash::new(Redirect::to(&format!["{}user/{}", route_prefix(), user.username]),
                  "info", &format!["Updated {}'s details", user.username]))
}
