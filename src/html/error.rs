use super::bootstrap;
use super::diesel;
use super::github;
use super::hyper;
use super::native_tls;
use super::rustc_serialize;

use chrono;
use maud; // TODO: use a Bootstrap::ResultType or somesuch
use maud::Render;
use rocket::*;
use rocket::response::{Responder, Response};

use std::env;
use std::error::Error as StdError;


#[derive(Debug)]
pub enum Error {
    /// Error authenticating user.
    AuthError(String),

    /// Login is required to access a resource.
    AuthRequired,

    /// There is a misconfiguration of the Clowder server itself.
    ConfigError(String),

    /// An otherwise-uninterpreted error occurred when interacting with the database.
    DatabaseError(diesel::result::Error),

    /// The user made an invalid request.
    BadRequest(String),

    /// We received invalid date from somewhere.
    InvalidData(String),

    /// There was a problem communicating with a remote host.
    NetError(hyper::Error),

    /// The user is not permitted to perform the requested action.
    NotAuthorized(String),
}

impl Error {
    fn kind(&self) -> &str {
        match self {
            &Error::AuthError(_) => "Authorization error",
            &Error::AuthRequired => "Authorization required",
            &Error::BadRequest(_) => "Bad request",
            &Error::ConfigError(_) => "Configuration error",
            &Error::DatabaseError(_) => "Database error",
            &Error::InvalidData(_) => "Invalid data",
            &Error::NetError(_) => "Network error",
            &Error::NotAuthorized(_) => "Authorization error",
        }
    }

    fn msg(self) -> String {
        match self {
            Error::AuthError(msg) => msg,
            Error::AuthRequired => String::from("Authorization required"),
            Error::BadRequest(msg) => msg,
            Error::ConfigError(msg) => msg,
            Error::DatabaseError(e) => e.description().to_string(),
            Error::InvalidData(msg) => msg,
            Error::NetError(e) => e.description().to_string(),
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

impl From<env::VarError> for Error {
    fn from(err: env::VarError) -> Error {
        Error::ConfigError(format!["problem with environment variable (or .env file): {}", err])
    }
}

impl From<github::Error> for Error {
    fn from(err: github::Error) -> Error {
        Error::AuthError(format!["GitHub error: {}", err])
    }
}

impl From<hyper::Error> for Error {
    fn from(err: hyper::Error) -> Error {
        Error::NetError(err)
    }
}

impl From<native_tls::Error> for Error {
    fn from(err: native_tls::Error) -> Error {
        Error::ConfigError(format!["unable to create TLS client: {}", err])
    }
}

impl From<rustc_serialize::json::DecoderError> for Error {
    fn from(err: rustc_serialize::json::DecoderError) -> Error {
        Error::InvalidData(format!["JSON error: {}", err.description()])
    }
}

impl<'r> Responder<'r> for Error {
    fn respond(self) -> response::Result<'r> {
        bootstrap::Page::new(self.kind())
                        .content(html! {
                            h1 (self.kind())
                            h2 (self.msg())
                        })
            .render()
            .respond()
    }
}


/// The error catcher for unauthorized accesses prompts for HTTP basic authentication.
#[error(401)]
fn unauthorized<'r>(_req: &Request) -> Response<'r> {
    const OAUTH_URL: &'static str = "https://github.com/login/oauth/authorize";

    let content = match env::var("CLOWDER_GH_CLIENT_ID") {
        Ok(ref id) => {
            bootstrap::ModalDialog::new("login")
                .title("Login required")
                .body(html! {
                    p {
                        i.fa.fa-github aria-hidden="true" {}
                        (maud::PreEscaped("&nbsp;"))
                        a href={ (OAUTH_URL) "?client_id=" (id) } "Sign in with GitHub"
                    }
                })
                .closeable(false)
                .start_open(true)
                .render()
        },

        Err(ref e) => {
            html! {
                h1 "Login required"
                p "Login required and GitHub redirection not configured:"
                pre (e)
            }
        },
    };

    let full_content =
        bootstrap::Page::new("401 Unauthorized")
            .content(content)
            .render()
            .into_string()
            ;

    use std::io;

    let mut response = Response::new();
    response.set_sized_body(io::Cursor::new(full_content));

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
