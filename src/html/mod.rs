use std::fs::File;
use std::io;

use chrono::UTC;
use chrono_humanize::HumanTime;
use ::db;
use db::models::*;
use db::schema::*;
use ::diesel;
use diesel::*;
use diesel::pg::PgConnection as Connection;
use marksman_escape::Escape;
use maud::*;
use rocket::*;
use rocket::request::{FlashMessage, Form};
use rocket::response::{Flash, Redirect};

// We do, in fact, use FromFrom, but only in a rocket-codegen derivation.
#[allow(unused_imports)]
use rocket::request::FromForm;

mod bootstrap;
mod error;
mod forms;
mod link;
mod tables;

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
    type Error = diesel::result::Error;

    fn from_request(req: &'a Request<'r>)
            -> request::Outcome<Context, Self::Error> {

        let conn = db::establish_connection();
        let user = match users::table.first(&conn) {
            Ok(u) => u,
            Err(err) => { return Outcome::Failure((http::Status::BadRequest, err)); },
        };

        Outcome::Success(Context { user: user, conn: conn })
    }
}


/// All of the routes that we can handle.
pub fn all_routes() -> Vec<Route> {
    routes! {
        index, logout,
        machine, machines,
        reservation, reservation_end, reservation_end_confirm, reservations,
        static_css, static_js,
        user, user_update,
    }
}

pub fn escape(dangerous: &str) -> String {
    String::from_utf8(Escape::new(dangerous.bytes()).collect())
           .unwrap_or(String::from("&lt;error&gt;"))
}


/// Render a normal (i.e., non-error) page of content.
pub fn render<S>(title: S, user: &User, flash: Option<FlashMessage>, content: Markup)
    -> Markup
    where S: Into<String>
{
    bootstrap::Page::new(title)
                    .content(content)
                    .flash(flash)
                    .nav(vec! {
                        bootstrap::NavItem::link("/machines", "Machines"),
                        bootstrap::NavItem::link("/reservations", "Reservations"),
                    })
                    .user(&user.username, &user.name)
                    .render()
}


#[get("/")]
fn index(ctx: Context) -> Result<Markup, Error> {
    let conn = db::establish_connection();

    let machines = try![{
        use self::machines::dsl::*;
        machines.order(name)
                .load::<Machine>(&conn)
    }];

    // TODO: use multiple joins once Diesel supports it
    let reservations: Vec<(Reservation, Machine)> = try![{
        use self::reservations::dsl::*;
        reservations.inner_join(machines::table)
                    .filter(actual_end.is_null())
                    .order(scheduled_start.desc())
                    .load(&conn)
    }];

    Ok(render("Clowder", &ctx.user, None, html! {
        div.row {
            div class="col-md-6" {
                h4 "Machine inventory"
                (tables::machines(&machines))
            }

            div class="col-md-6" {
                h4 "Current reservations"
                (try![tables::reservations_with_machines(&reservations, &ctx, false)])
            }
        }
    }))
}

#[get("/logout")]
fn logout(ctx: Context) -> Result<Markup, Error> {
    Ok(render("Logout", &ctx.user, None,
            bootstrap::callout("warning", "Unhandled operation",
                    PreEscaped("We don't handle logout just yet.".to_string()))))
}

#[get("/machine/<machine_name>")]
fn machine(machine_name: &str, ctx: Context) -> Result<Markup, Error> {
    let m: Machine = try![{
        use self::machines::dsl::*;
        machines.filter(name.eq(machine_name))
                .first(&ctx.conn)
    }];

    let reserv: Vec<(Reservation, User)> = try![{
        use self::reservations::dsl::*;
        reservations.inner_join(users::table)
                    .filter(machine_id.eq(m.id))
                    .filter(user_id.eq(users::dsl::id))
                    .order(scheduled_start.desc())
                    .load(&ctx.conn)
    }];

    Ok(render(format!["Clowder: {}", m.name], &ctx.user, None, html! {
        div.row h2 (m.name)

        div.row {
            div class="col-md-7" {
                p {
                    (m.arch) " (" (m.microarch) "), "
                    (m.cores) " cores, " (m.memory_gb) " GiB RAM"
                }

                p a href={ "/reservation/create/?machine=" (m.name) } "Reserve this machine"
            }

            div class="col-md-5" {
                h3 "Reservations"

                table.table.table-responsive {
                    (tables::TableHeader::from_str(
                        &[ "", "User", "Started", "Ends" ]))

                    tbody {
                        @for (ref r, ref u) in reserv {
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
    }))
}

#[get("/machines")]
fn machines(ctx: Context) -> Result<Markup, Error> {
    let machines = try![{
        use self::machines::dsl::*;
        machines.order(name)
                .load::<Machine>(&ctx.conn)
    }];

    Ok(render("Clowder: Machines", &ctx.user, None, tables::machines(&machines)))
}

#[get("/reservation/<id>")]
fn reservation(id: i32, ctx: Context, flash: Option<FlashMessage>) -> Result<Markup, Error> {
    let r: Reservation = try![reservations::table.find(id).first(&ctx.conn)];
    let machine: Machine = try![machines::table.find(r.machine_id).first(&ctx.conn)];
    let user: User = try![users::table.find(r.user_id).first(&ctx.conn)];

    let can_end = match (r.scheduled_start, r.actual_end) {
        (s, None) if s <= UTC::now() => true,
        (_, _) => false,
    };

    Ok(render(format!["Clowder: reservation {}", r.id], &ctx.user, flash, html! {
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

#[get("/reservation/end/<id>")]
fn reservation_end(id: i32, ctx: Context) -> Result<Markup, Error> {
    let r: Reservation = try![reservations::table.find(id).first(&ctx.conn)];
    let machine: Machine = try![machines::table.find(r.machine_id).first(&ctx.conn)];
    let user: User = try![users::table.find(r.user_id).first(&ctx.conn)];

    Ok(render(format!["Clowder: end reservation {}", r.id], &ctx.user, None, html! {
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
    use db::schema::reservations::dsl::*;

    let r: Reservation = try![reservations.find(res_id).first(&ctx.conn)];

    try! {
        diesel::update(&r)
            .set(actual_end.eq(Some(UTC::now())))
            .get_result::<Reservation>(&ctx.conn)
    };

    Ok(Flash::new(Redirect::to(&format!["/reservation/{}", res_id]), "info",
                               &format!["Ended reservation {}", res_id]))
}

#[get("/reservations")]
fn reservations(ctx: Context) -> Result<Markup, Error> {
    // TODO: use multiple joins once Diesel supports it
    let reservations: Vec<(Reservation, Machine)> = try![{
        use db::schema::reservations::dsl::*;

        reservations
            .inner_join(machines::table)
            .order(scheduled_start.desc())
            .load(&ctx.conn)
    }];

    Ok(render("Clowder: Reservations", &ctx.user, None,
                         try![tables::reservations_with_machines(&reservations, &ctx, true)]))
}

#[get("/static/css/<filename>")]
fn static_css(filename: &str) -> io::Result<File> {
    File::open(format!["static/css/{}", filename])
}

#[get("/static/js/<filename>")]
fn static_js(filename: &str) -> io::Result<File> {
    File::open(format!["static/js/{}", filename])
}

#[get("/user/<name>")]
fn user(name: String, ctx: Context) -> Result<Markup, Error> {
    let user: User = try![users::table.filter(users::dsl::username.eq(name)).first(&ctx.conn)];
    let name = user.name.as_str();
    let myself = user.id == ctx.user.id;

    let reservations: Vec<(Reservation, Machine)> = try![{
        use db::schema::reservations::dsl::*;

        reservations
            .inner_join(machines::table)
            .filter(user_id.eq(user.id))
            .order(scheduled_start.desc())
            .load(&ctx.conn)
    }];

    Ok(render(name, &ctx.user, None, html! {
        h2 (name)

        div.row {
            div class="col-md-4" {
                form action={ "/user/update/" (user.username) } method="post" {
                    table.table.table-responsive {
                        tbody {
                            tr { th "Username" td (user.username) }
                            tr { th "Name"
                                td (forms::Input::new("name".to_string())
                                                 .value(user.name.clone())
                                                 .size(18)
                                                 .writable(myself))
                            }
                            tr { th "Email"
                                td (forms::Input::new("email".to_string())
                                                 .value(user.email.clone())
                                                 .size(18)
                                                 .writable(myself))
                            }
                            tr {
                                th "Phone"
                                td (forms::Input::new("phone".to_string())
                                                 .value(user.phone.as_ref()
                                                                  .map(String::clone)
                                                                  .unwrap_or(String::new()))
                                                 .size(18)
                                                 .writable(myself))
                            }
                            tr td colspan="2"
                                input.btn.btn-block.btn-warning type="submit"
                                    value="Update user details" /
                        }
                    }
                }
            }

            div class="col-md-8"
                (try![tables::reservations_with_machines(&reservations, &ctx, true)])
        }
    }))
}


#[derive(FromForm)]
struct UserUpdate {
    name: String,
    email: String,
    phone: Option<String>,
}

#[post("/user/update/<who>", data = "<form>")]
fn user_update(who: &str, ctx: Context, form: Form<UserUpdate>) -> Result<Flash<Redirect>, Error> {
    use self::users::dsl::*;

    let user: User = try! {
        users.filter(username.eq(who))
             .first(&ctx.conn)
             .map_err(|err| Error::BadRequest(format!["No such user: '{}' ({})", who, err]))
    };

    if user.id != ctx.user.id /* TODO: and not superuser */ {
        return Err(Error::NotAuthorized(String::from("update other users' details")))
    }

    let f = form.get();
    try! {
        diesel::update(&user)
            .set(name.eq(f.name.clone()))
            .get_result::<User>(&ctx.conn)
    };

    try! {
        diesel::update(&user)
            .set(email.eq(f.email.clone()))
            .get_result::<User>(&ctx.conn)
    };

    if let Some(ref p) = f.phone {
        try! {
            diesel::update(&user)
                .set(phone.eq(Some(p.clone())))
                .get_result::<User>(&ctx.conn)
        };
    };

    Ok(Flash::new(Redirect::to(&format!["/user/{}", user.username]),
                  "info", &format!["Updated {}'s details", user.username]))
}
