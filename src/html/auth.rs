/*
 * Copyright 2016-2017 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use db::models::*;
use diesel::pg::PgConnection as Connection;
use rocket::http::{Cookie, Cookies};
use rocket::request;
use std::env;

use html::error::Error;
use super::rocket;

/// The name of the cookie we set (with authenticated encryption) for the user's username.
static AUTH_COOKIE_NAME: &'static str = "clowder_user";


///
/// A struct that authenticates users given a MAC'ed cookie or a debug auth bypass
/// (e.g., `CLOWDER_FAKE_GITHUB_USERNAME`).
///
struct Authenticator {
    conn: Connection,
}

impl Authenticator {
    fn new() -> Authenticator {
        Authenticator { conn: super::db::establish_connection() }
    }

    ///
    /// Attempt to authenticate a Clowder user using a MAC'ed cookie, or if that fails, whatever
    /// fallback authentication methods are permitted by local policy (e.g., fake/test auth data).
    ///
    fn authenticate(self, cookies: &mut Cookies) -> Result<AuthContext, Error> {
        let user = cookies.get_private(AUTH_COOKIE_NAME)
            .ok_or(Error::AuthRequired)
            .and_then(|ref username| self.lookup_user(username.value()))
            .or_else(|_| self.try_fake_auth())?;

        Ok(AuthContext {
            conn: self.conn,
            user: user,
        })
    }

    ///
    /// Attempt to look up a user (by Clowder username) in the user database.
    ///
    fn lookup_user(&self, clowder_username: &str) -> Result<User, Error> {
        User::with_username(&clowder_username, &self.conn).map_err(Error::DatabaseError)
    }

    ///
    /// Attempt to authenticate the user with fake (bypass) authentication methods, which may be:
    ///
    /// `CLOWDER_FAKE_GITHUB_USERNAME`
    /// : if set in the environment (or `.env`), treat this as a verified GitHub username
    ///
    fn try_fake_auth(&self) -> Result<User, Error> {
        env::var_os("CLOWDER_FAKE_GITHUB_USERNAME")
            .and_then(|s| s.to_str().map(str::to_string))
            .ok_or(Error::AuthRequired)
            .and_then(|username| {
                GithubAccount::get(&username, &self.conn)
                    .map_err(Error::DatabaseError)
                    .map(|(_, user)| user)
            })
    }
}


///
/// Context for a logged-in user with the right to access the database.
///
pub struct AuthContext {
    /// Database connection
    pub conn: Connection,

    /// The authenticated user
    pub user: User,
}

impl<'a, 'r> request::FromRequest<'a, 'r> for AuthContext {
    type Error = super::error::Error;

    fn from_request(req: &'a request::Request<'r>) -> request::Outcome<AuthContext, super::Error> {
        let auth_context = Authenticator::new().authenticate(&mut req.cookies());

        match auth_context {
            Ok(ctx) => rocket::outcome::Outcome::Success(ctx),
            Err(e) => {
                let failure = match e {
                    Error::AuthRequired => (rocket::http::Status::Unauthorized, e),
                    _ => (rocket::http::Status::InternalServerError, e),
                };

                rocket::outcome::Outcome::Failure(failure)
            }
        }
    }
}


/// Log the user out by clearing their auth cookie.
pub fn logout<'c>(mut jar: Cookies) {
    jar.get_private(AUTH_COOKIE_NAME).map(|c| jar.remove_private(c));
}

/// Generate a cookie that attests to a logged-in user's username.
pub fn set_user_cookie<'c, S: Into<String>>(mut jar: Cookies, username: S) -> Result<(), Error> {
    jar.add_private(Cookie::new(String::from(AUTH_COOKIE_NAME), username.into()));
    Ok(())
}
