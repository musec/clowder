use super::bootstrap;
use super::diesel;

use chrono;
use maud; // TODO: use a Bootstrap::ResultType or somesuch
use maud::Render;
use rocket::*;
use rocket::response::{Responder, Response};

use std::error::Error as StdError;


#[derive(Debug)]
pub enum Error {
    /// Login is required to access a resource.
    AuthRequired,

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
            &Error::AuthRequired => "Authorization required",
            &Error::BadRequest(_) => "Bad request",
            &Error::DatabaseError(_) => "Database error",
            &Error::NotAuthorized(_) => "Authorization error",
        }
    }

    fn msg(&self) -> String {
        match self {
            Error::AuthRequired => String::from("Authorization required"),
            Error::BadRequest(msg) => msg,
            Error::DatabaseError(e) => e.description().to_string(),
            Error::NotAuthorized(msg) => msg,
        }
    }
}

impl From<chrono::ParseError> for Error {
    fn from(err: chrono::ParseError) -> Error {
        Error::BadRequest(err.description().to_string())
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
fn unauthorized<'r>(req: &Request) -> Response<'r> {
    let content = bootstrap::Page::new("401 Unauthorized")
        .content(html! {
            h1 "401 Unauthorized"
            p { "Authorization is required to access " code (req.uri()) "." }
        })
        .render()
        .into_string()
        ;

    use std::io::Cursor;

    let mut response = Response::new();
    response.set_status(http::Status::Unauthorized);
    response.set_header(http::Header::new("WWW-Authenticate", "Basic realm=\"Clowder\""));
    response.set_sized_body(Cursor::new(content));

    response
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
