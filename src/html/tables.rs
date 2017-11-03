use chrono::Utc;
use chrono_humanize::HumanTime;
use db::models::*;
use db::schema::*;
use diesel::result::Error as DieselError;
use diesel::{FindDsl,FirstDsl};
use html::link::Link;
use maud::*;

use super::Context;

type MarkupOrDieselError = Result<Markup, DieselError>;


///
/// The header of a Bootstrap-compatible HTML table.
///
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


///
/// An HTML table that shows various properties of machines.
///
/// This table will always show each machine's name, but it can also show:
///
///  - macroarchitecture (e.g., "x86_64")
///  - microarchitecture (e.g., "Sandy Bridge")
///  - processor variant (e.g., "Xeon E3-1240 v5")
///  - processor clock speed
///  - number of physical cores
///  - size of physical memory
///
/// The default is to show all of these values, but this can be disabled by calling various
/// builder methods, e.g.:
///
/// ```rust
/// let machines = FullMachine::all(&db_connection)?;
/// let markup = MachineTable::new(machines).show_microarch(false).render();
/// ```
///
pub struct MachineTable {
    machines: Vec<FullMachine>,

    show_arch: bool,
    show_cores: bool,
    show_freq: bool,
    show_memory: bool,
    show_microarch: bool,
    show_processor_name: bool,
}

impl MachineTable {
    pub fn new<MV>(machines: MV) -> MachineTable
        where MV: Into<Vec<FullMachine>>
    {
        MachineTable {
            machines: machines.into(),
            show_arch: true,
            show_cores: true,
            show_freq: true,
            show_memory: true,
            show_microarch: true,
            show_processor_name: true,
        }
    }

    fn headers(&self) -> Vec<&str> {
        [
            vec![ "Name" ],
            if self.show_arch { vec![ "Arch" ] } else { vec![] },
            if self.show_processor_name { vec![ "Proc" ] } else { vec![] },
            if self.show_microarch { vec![ "Microarch" ] } else { vec![] },
            if self.show_cores { vec![ "Cores" ] } else { vec![] },
            if self.show_freq { vec![ "Freq" ] } else { vec![] },
            if self.show_memory { vec![ "Memory" ] } else { vec![] },
        ]
        .concat()
    }

    fn render_machine(&self, m: &FullMachine) -> Markup {
        html! {
            tr {
                td (Link::from(m.machine()))
                @if self.show_arch { td (m.architecture().name) }
                @if self.show_processor_name {
                    td {
                        @if let Some(ref url) = m.processor().url {
                            a href=(url) (m.processor().name)
                        } @else {
                            (m.processor().name)
                        }
                    }
                }
                @if self.show_microarch {
                    td {
                        @let microarch = m.microarchitecture();
                        @if let Some(ref url) = microarch.url {
                            a href=(url) (microarch.name)
                        } @else {
                            (microarch.name)
                        }
                    }
                }
                @if self.show_cores { td.numeric (m.cores()) }
                @if self.show_freq { td.numeric { (m.freq_ghz()) " GHz" } }
                @if self.show_memory { td.numeric { (m.memory_gb()) " GiB" } }
            }
        }
    }

    pub fn show_arch(mut self, s: bool) -> MachineTable {
        self.show_arch = s;
        self
    }

    pub fn show_cores(mut self, s: bool) -> MachineTable {
        self.show_cores = s;
        self
    }

    pub fn show_freq(mut self, s: bool) -> MachineTable {
        self.show_freq = s;
        self
    }

    pub fn show_memory(mut self, s: bool) -> MachineTable {
        self.show_memory = s;
        self
    }

    pub fn show_microarch(mut self, s: bool) -> MachineTable {
        self.show_microarch = s;
        self
    }

    pub fn show_processor_name(mut self, s: bool) -> MachineTable {
        self.show_processor_name = s;
        self
    }
}

impl Render for MachineTable {
    fn render(&self) -> Markup {
        html! {
            table.table.table-responsive {
                (TableHeader::from_str(&self.headers()))

                tbody {
                    @for m in &self.machines {
                        (self.render_machine(m))
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
