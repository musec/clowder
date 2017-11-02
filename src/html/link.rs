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

impl<'a> From<&'a Reservation> for Link {
    fn from(r: &Reservation) -> Link {
        Link {
            url: format!["{}reservation/{}", super::route_prefix(), r.id],
            text: format!["{}", r.id],
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
        html! { a href=(self.url) (self.text) }
    }
}

