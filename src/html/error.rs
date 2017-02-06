use super::bootstrap;
use super::diesel;

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
        bootstrap::render(self.kind(), None, None, html! {
            h1 (self.kind())
            p (self.msg())
        })
        .respond()
    }
}
