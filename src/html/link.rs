/*
 * Copyright 2016-2018 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use db::models::*;
use maud::*;

pub struct Link {
    url: String,
    text: String,
}

impl<'a> From<&'a Machine> for Link {
    fn from(m: &Machine) -> Link {
        Link {
            url: format!["{}machine/{}", super::route_prefix(), m.name],
            text: m.name.clone(),
        }
    }
}

impl<'a> From<&'a Microarchitecture> for Link {
    fn from(p: &Microarchitecture) -> Link {
        Link {
            url: p.url.clone().unwrap_or(String::from("#")),
            text: p.name.clone(),
        }
    }
}

impl<'a> From<&'a Processor> for Link {
    fn from(p: &Processor) -> Link {
        Link {
            url: p.url.clone().unwrap_or(String::from("#")),
            text: p.name.clone(),
        }
    }
}

impl<'a> From<&'a Reservation> for Link {
    fn from(r: &Reservation) -> Link {
        Link {
            url: format!["{}reservation/{}", super::route_prefix(), r.id()],
            text: format!["{}", r.id()],
        }
    }
}

impl<'a> From<&'a User> for Link {
    fn from(u: &User) -> Link {
        Link {
            url: format!["{}user/{}", super::route_prefix(), u.username],
            text: u.username.clone(),
        }
    }
}

impl Render for Link {
    fn render(&self) -> Markup {
        html! { a href=(self.url) { (self.text) } }
    }
}
