use super::bootstrap;
use super::diesel;

use maud; // TODO: use a Bootstrap::ResultType or somesuch
use maud::Render;
use rocket::*;
use rocket::response::Responder;


#[derive(Debug)]
pub enum Error {
    /// An otherwise-uninterpreted error occurred when interacting with the database.
    DatabaseError(diesel::result::Error),

    /// The user made an invalid request.
    BadRequest(String),

    /// The user is not permitted to perform the requested action.
    NotAuthorized(String),
}

impl Error {
    fn kind(&self) -> &str {
        match self {
            _ => "foo",
        }
    }

    fn msg(&self) -> String {
        match self {
            _ => String::new(),
        }
    }
}

impl From<diesel::result::Error> for Error {
    fn from(err: diesel::result::Error) -> Error {
        Error::DatabaseError(err)
    }
}

impl<'r> Responder<'r> for Error {
    fn respond(self) -> response::Result<'r> {
        bootstrap::Page::new(self.kind())
                        .content(html! { p (self.msg()) })
            .render()
            .respond()
    }
}


/// The error catcher for unauthorized accesses prompts for HTTP basic authentication.
#[error(401)]
fn unauthorized(req: &Request) -> maud::Markup {
    bootstrap::Page::new("401 Unauthorized")
        .content(html! {
            h1 "401 Unauthorized"
            p { "Authorization is required to access " code (req.uri()) "." }
        })
        .render()
}

/// The 404 handler renders a slightly nicer-looking page than the stock Rocket handler.
#[error(404)]
fn not_found(req: &Request) -> maud::Markup {
    bootstrap::Page::new("404 Not Found")
        .content(html! {
            h2 ("404 Not Found")
            p { "The resource " code (req.uri()) " could not be found." }
        })
        .render()
}
