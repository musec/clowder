/*
 * Copyright 2016-2017, 2019 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use db::models::*;
use diesel::pg::PgConnection as Connection;
use diesel::result::Error as DieselError;
use rocket::http::Cookies;
use rocket::request;
use std::{env, fmt};

use crate::error::Error;
use super::github;
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
    fn new() -> Result<Authenticator, Error> {
        super::db::establish_connection()
            .map(|conn| Authenticator { conn })
    }

    ///
    /// Attempt to authenticate a Clowder user using a MAC'ed cookie, or if that fails, whatever
    /// fallback authentication methods are permitted by local policy (e.g., fake/test auth data).
    ///
    fn authenticate(self, cookies: &mut Cookies) -> Result<AuthContext, Error> {
        let user = cookies
            .get_private(AUTH_COOKIE_NAME)
            .ok_or(Error::AuthRequired)
            .and_then(|ref username| self.lookup_user(username.value()))
            .or_else(|_| self.try_fake_auth())?;

        Ok(AuthContext {
            conn: self.conn,
            user: user,
        })
    }

    ///
    /// Look up a known GitHub user in the Clowder user database.
    ///
    fn github_user(&self, gh_username: &str) -> Result<User, Error> {
        GithubAccount::get(&gh_username, &self.conn)
            .map(|(_, user)| user)
            .map_err(|_| Error::AuthError(format!["Unknown GitHub user '{}'", gh_username]))
    }

    ///
    /// Attempt to look up a user (by Clowder username) in the user database.
    ///
    fn lookup_user(&self, clowder_username: &str) -> Result<User, Error> {
        User::with_username(&clowder_username, &self.conn)
            .map_err(|e| match e {
                diesel::result::Error::NotFound => {
                    Error::AuthError(format!["No such user ('{}')", clowder_username])
                },
                e => Error::DatabaseError(e)
            })
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
            .and_then(|username| self.github_user(&username))
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

impl fmt::Debug for AuthContext {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write![f, "AuthContext for User '{:?}'", self.user]
    }
}

impl<'a, 'r> request::FromRequest<'a, 'r> for AuthContext {
    type Error = Error;

    fn from_request(req: &'a request::Request<'r>) -> request::Outcome<AuthContext, Self::Error> {
        let auth_context = Authenticator::new()
            .and_then(|a| a.authenticate(&mut req.cookies()));

        match auth_context {
            Ok(ctx) => rocket::outcome::Outcome::Success(ctx),
            Err(e) => {
                warn!["Authentication failed: {:?}", e];

                let failure = match e {
                    Error::AuthError(_) => (rocket::http::Status::Forbidden, e),
                    Error::AuthRequired => (rocket::http::Status::Unauthorized, e),
                    Error::DatabaseError(DieselError::NotFound) => {
                        (rocket::http::Status::Forbidden, e)
                    }
                    _ => (rocket::http::Status::InternalServerError, e),
                };

                rocket::outcome::Outcome::Failure(failure)
            }
        }
    }
}

/// Handle a GitHub OAuth callback.
pub fn github_callback(code: String, cookies: rocket::http::Cookies) -> Result<(), Error> {
    let username = github::auth_callback(code)?;

    Authenticator::new()?
        .github_user(&username)
        .map(|user| set_user_cookie(cookies, user.username))
}

/// Log the user out by clearing their auth cookie.
pub fn logout<'c>(mut jar: Cookies) {
    jar.get_private(AUTH_COOKIE_NAME)
        .map(|c| jar.remove_private(c));
}

/// Generate a cookie that attests to a logged-in user's username.
pub fn set_user_cookie<'c, S: Into<String>>(mut jar: Cookies, username: S) {
    jar.add_private(rocket::http::Cookie::new(
        String::from(AUTH_COOKIE_NAME),
        username.into(),
    ))
}
