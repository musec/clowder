/*
 * Copyright 2016-2019 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use super::bootstrap;
use super::diesel;
use super::hyper;
use super::native_tls;
use super::rustc_serialize;

use chrono;
use maud;
use maud::{html, Render};
use rocket;

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

impl std::fmt::Display for Error {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        match self {
            &Error::AuthError(ref msg) => write![f, "Failed to authenticate: {}", msg],
            &Error::AuthRequired => write![f, "Authorization required"],
            &Error::BadRequest(ref req) => write![f, "{}", req],
            &Error::ConfigError(ref msg) => write![f, "{}", msg],
            &Error::DatabaseError(ref e) => write![f, "{:?}", e],
            &Error::InvalidData(ref msg) => write![f, "{}", msg],
            &Error::NetError(ref e) => write![f, "{:?}", e],
            &Error::NotAuthorized(ref action) => write![f, "Not authorized to {}", action],
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
        Error::ConfigError(format![
            "problem with environment variable (or .env file): {}",
            err
        ])
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

impl Into<bootstrap::Page> for Error {
    fn into(self) -> bootstrap::Page {
        bootstrap::Page::new(self.kind())
            .content(html! {
                h1 { (self.kind()) }
                h2 { (self.msg()) }
            })
            .link_prefix(super::route_prefix())
    }
}

/// The error catcher for unauthorized accesses prompts for HTTP basic authentication.
#[catch(401)]
pub fn unauthorized(_req: &rocket::Request) -> bootstrap::Page {
    const OAUTH_URL: &'static str = "https://github.com/login/oauth/authorize";

    let content = match env::var("CLOWDER_GH_CLIENT_ID") {
        Ok(ref id) => bootstrap::ModalDialog::new("login")
            .title("Login required")
            .body(html! {
                p {
                    a.btn.btn-secondary.large href={ (OAUTH_URL) "?client_id=" (id) } {
                        i.fa.fa-github aria-hidden="true" {}
                        (maud::PreEscaped("&nbsp;"))
                        "Sign in with GitHub"
                    }
                }
            })
            .footer(html! {
                p.footnote {
                    "Contact "
                    a href="https://www.engr.mun.ca/~anderson/" { "Jonathan Anderson" }
                    " for authorization to use this system."
                }
            })
            .closeable(false)
            .start_open(true)
            .render(),

        Err(ref e) => {
            html! {
                h1 { "Login required" }
                p { "Login required and GitHub redirection not configured:" }
                pre { (e) }
            }
        }
    };

    bootstrap::Page::new("401 Unauthorized")
        .content(content)
        .link_prefix(super::route_prefix())
}

/// 403 forbidden means that (re-)authenticating won't help.
#[catch(403)]
pub fn forbidden() -> bootstrap::Page {
    bootstrap::Page::new("403 Forbidden")
        .content(html! {
            h2 { ("403 Forbidden") }
            p { "You are not authorized to access this resource." }
        })
        .link_prefix(super::route_prefix())
}

/// The 404 handler renders a slightly nicer-looking page than the stock Rocket handler.
#[catch(404)]
pub fn not_found(req: &rocket::Request) -> bootstrap::Page {
    bootstrap::Page::new("404 Not Found")
        .content(html! {
            h2 { ("404 Not Found") }
            p { "The resource " code { (req.uri()) } " could not be found." }
        })
        .link_prefix(super::route_prefix())
}

/// The 500 ISE (Internal Server Error) handler doesn't provide any more information than the
/// stock Rocket handler, but it also looks nicer.
#[catch(500)]
pub fn internal_server_error(_req: &rocket::Request) -> bootstrap::Page {
    bootstrap::Page::new("500 Internal Server Error")
        .content(html! {
            h2 { ("500 Internal Server Error") }
            p { "There is an error in Clowder!" }
        })
        .link_prefix(super::route_prefix())
}
