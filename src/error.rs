/*
 * Copyright 2016-2019 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use super::diesel;
use super::hyper;
use super::native_tls;
use super::rustc_serialize;

use chrono;

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
    pub fn kind(&self) -> &str {
        match self {
            &Error::AuthError(_) => "Authentication error",
            &Error::AuthRequired => "Authorization required",
            &Error::BadRequest(_) => "Bad request",
            &Error::ConfigError(_) => "Configuration error",
            &Error::DatabaseError(_) => "Database error",
            &Error::InvalidData(_) => "Invalid data",
            &Error::NetError(_) => "Network error",
            &Error::NotAuthorized(_) => "Authorization error",
        }
    }
}

impl std::error::Error for Error {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            &Error::DatabaseError(ref e) => Some(e),
            &Error::NetError(ref e) => Some(e),
            _ => None,
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
