/*
 * Copyright 2017 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use chrono::Utc;
use chrono_humanize::HumanTime;
use db::models::*;
use html::link::Link;
use maud::*;


///
/// The header of a Bootstrap-compatible HTML table.
///
pub struct TableHeader (Vec<String>);

impl TableHeader {
    pub fn new(strs: &[&str]) -> TableHeader {
        TableHeader(strs.into_iter()
                        .map(|s| s.to_string())
                        .collect())
    }

    pub fn add_if<S>(mut self, condition: bool, s: S) -> TableHeader
        where S: Into<String>
    {
        if condition {
            self.0.push(s.into());
        }

        self
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

    fn header(&self) -> TableHeader {
        TableHeader::new(&[ "Name" ])
            .add_if(self.show_arch, "Arch")
            .add_if(self.show_processor_name, "Proc")
            .add_if(self.show_microarch, "Microarch")
            .add_if(self.show_cores, "Cores")
            .add_if(self.show_freq, "Freq")
            .add_if(self.show_memory, "Memory")
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
                (self.header())

                tbody {
                    @for m in &self.machines {
                        (self.render_machine(m))
                    }
                }
            }
        }
    }
}


type ReservationData = (Reservation, Option<Machine>, Option<User>);

///
/// An HTML table that shows information about reservations.
///
/// This table can show:
///
///  - the user
///  - the machine
///  - scheduled start and end times
///  - actual end status
///
/// The default is to show all of these values, but this can be disabled by calling various
/// builder methods, e.g.:
///
/// ```rust
/// let reservations = Reservation::all(&db_connection)?;
/// let markup = ReservationTable::new(reservations).show_ended(false).render();
/// ```
///
pub struct ReservationTable {
    reservations: Vec<ReservationData>,

    show_actual_end: bool,
    show_machine: bool,
    show_scheduled_end: bool,
    show_scheduled_start: bool,
    show_user: bool,
}

impl ReservationTable {
    pub fn new(reservations: Vec<ReservationData>) -> ReservationTable {
        ReservationTable {
            reservations: reservations,
            show_actual_end: true,
            show_machine: true,
            show_scheduled_end: true,
            show_scheduled_start: true,
            show_user: true,
        }
    }

    fn header(&self) -> TableHeader {
        TableHeader::new(&[ "#" ])
            .add_if(self.show_machine, "Machine")
            .add_if(self.show_user, "User")
            .add_if(self.show_scheduled_start, "Start")
            .add_if(self.show_scheduled_end, "Scheduled end")
            .add_if(self.show_actual_end, "Actually ended")
    }

    fn render(&self, data: &(Reservation, Option<Machine>, Option<User>)) -> Markup {
        let &(ref r, ref m, ref u) = data;

        let now = Utc::now();
        let row_class = {
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

        html! {
            tr class=(row_class) {
                td (Link::from(r))

                @if self.show_machine {
                    td {
                        @if let &Some(ref machine) = m {
                            (Link::from(machine))
                        }
                    }
                }

                @if self.show_user {
                    td {
                        @if let &Some(ref user) = u {
                            (Link::from(user))
                        }
                    }
                }

                @if self.show_scheduled_start {
                    td {
                        (HumanTime::from(r.start()))
                    }
                }

                @if self.show_scheduled_end {
                    td {
                        @if let Some(ref time) = r.scheduled_end {
                            (HumanTime::from(*time))
                        }
                    }
                }

                @if self.show_actual_end {
                    td {
                        @if let Some(ref time) = r.actual_end {
                            (HumanTime::from(*time))
                        }
                    }
                }
            }
        }
    }

    pub fn show_actual_end(mut self, s: bool) -> ReservationTable {
        self.show_actual_end = s;
        self
    }

    pub fn show_machine(mut self, s: bool) -> ReservationTable {
        self.show_machine = s;
        self
    }

    pub fn show_scheduled_end(mut self, s: bool) -> ReservationTable {
        self.show_scheduled_end = s;
        self
    }

    pub fn show_scheduled_start(mut self, s: bool) -> ReservationTable {
        self.show_scheduled_start = s;
        self
    }

    pub fn show_user(mut self, s: bool) -> ReservationTable {
        self.show_user = s;
        self
    }
}

impl Render for ReservationTable {
    fn render(&self) -> Markup {
        html! {
            table.table.table-responsive {
                (self.header())

                tbody {
                    @for r in &self.reservations {
                        (self.render(r))
                    }
                }
            }
        }
    }
}
