use chrono::Utc;
use chrono_humanize::HumanTime;
use db::models::*;
use db::schema::*;
use diesel::result::Error as DieselError;
use diesel::{FindDsl,FirstDsl,LoadDsl};
use html::link::Link;
use maud::*;

use super::Context;

type MarkupOrDieselError = Result<Markup, DieselError>;


pub fn machines(machines: &Vec<Machine>) -> Markup {
    html! {
        table.table.table-responsive {
            (TableHeader::from_str(
                &[ "Name", "Arch", "Microarch", "Cores", "Memory" ]))

            tbody {
                @for m in machines {
                    tr {
                        td (Link::from(m))
                        td (m.arch)
                        td (m.microarch)
                        td.numeric (m.cores)
                        td.numeric { (m.memory_gb) " GiB" }
                    }
                }
            }
        }
    }
}

pub fn reservations_with_machines(reservations: &Vec<(Reservation, Machine)>,
                                  ctx: &Context, show_actual_ends: bool)
        -> MarkupOrDieselError {

    let now = Utc::now();
    let row_class = |r: &Reservation| {
        if r.scheduled_start <= now {
            if let Some(_) = r.actual_end {
                ""
            } else if let Some(end) = r.scheduled_end {
                if end <= now {
                    "table-danger"
                } else {
                    "table-active"
                }
            } else {
                "table-active"
            }
        } else {
            "table-info"
        }
    };

    let mut headings = vec![ "", "Machine", "User", "Start", "Scheduled end" ];
    if show_actual_ends {
        headings.extend([ "Actual end" ].iter());
    };

    Ok(html! {
        table.table.table-responsive {
            (TableHeader::from_str(&headings))

            tbody {
                @for &(ref r, ref m) in reservations {
                    tr class=(row_class(r)) {
                        td (Link::from(r))
                        td (Link::from(m))

                        td ({
                            let u:User = try! {
                                users::table.find(r.user_id)
                                            .first(&ctx.conn)
                            };

                            Link::from(&u)
                        })

                        td (HumanTime::from(r.start()))
                        td (r.scheduled_end
                             .map(|t| format!["{}", HumanTime::from(t)])
                             .unwrap_or(String::new()))

                        @if show_actual_ends {
                            td (r.actual_end
                                .map(|t| format!["{}", HumanTime::from(t)])
                                .unwrap_or(String::new()))
                        }
                    }
                }
            }
        }
    })
}



pub struct TableHeader (Vec<String>);

impl TableHeader {
    pub fn from_str(strs: &[&str]) -> TableHeader {
        TableHeader(strs.iter()
                        .map(|s| s.to_string())
                        .collect::<Vec<_>>())
    }
}

impl Render for TableHeader {
    fn render(&self) -> Markup {
        html! {
            thead.thead-default {
                tr {
                    @for ref heading in &self.0 {
                        th (heading)
                    }
                }
            }
        }
    }
}


