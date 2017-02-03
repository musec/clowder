use db::models::*;
use maud::*;

pub struct Link {
    url: String,
    text: String,
}

impl Link {
    pub fn to<S1, S2>(url: S1, text: S2) -> Link
        where S1: Into<String>, S2: Into<String>
    {
        Link {
            url: url.into(),
            text: text.into(),
        }
    }
}

impl<'a> From<&'a Machine> for Link {
    fn from(m: &Machine) -> Link {
        Link {
            url: format!["/machine/{}", m.name],
            text: m.name.clone(),
        }
    }
}

impl<'a> From<&'a Reservation> for Link {
    fn from(r: &Reservation) -> Link {
        Link {
            url: format!["/reservation/{}", r.id],
            text: format!["{}", r.id],
        }
    }
}

impl<'a> From<&'a User> for Link {
    fn from(u: &User) -> Link {
        Link {
            url: format!["/user/{}", u.username],
            text: u.username.clone(),
        }
    }
}

impl Render for Link {
    fn render(&self) -> Markup {
        html! { a href=(self.url) (self.text) }
    }
}

