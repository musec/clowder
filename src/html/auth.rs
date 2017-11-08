use crypto::hmac::Hmac;
use crypto::mac::Mac;
use crypto::sha2::Sha512;
use db::models::*;
use diesel::pg::PgConnection as Connection;
use rand::Rng;
use rand::os::OsRng;
use rocket::http::{Cookie,Cookies};
use rocket::request;
use rustc_serialize::hex::ToHex;
use std::env;

use html::error::Error;
use super::rocket;

lazy_static! {
    /// A random HMAC key that is regenerated on every server restart.
    static ref HMAC_KEY: [u8; 32] = OsRng::new().expect("Unable to open system RNG").gen();
}


///
/// A struct that authenticates users given a MAC'ed cookie or a debug auth bypass
/// (e.g., `CLOWDER_FAKE_GITHUB_USERNAME`).
///
struct Authenticator {
    conn: Connection,
}

impl Authenticator {
    fn new() -> Authenticator {
        Authenticator {
            conn: super::db::establish_connection(),
        }
    }

    ///
    /// Attempt to authenticate a Clowder user using a MAC'ed cookie, or if that fails, whatever
    /// fallback authentication methods are permitted by local policy (e.g., fake/test auth data).
    ///
    fn authenticate(self, cookies: &mut Cookies) -> Result<AuthContext, Error> {
        let user = cookies.get_private("clowder_username")
                          .ok_or(Error::AuthRequired)
                          .and_then(check_mac)
                          .and_then(|username| self.lookup_user(&username))
                          .or_else(|_| self.try_fake_auth())
                          ?;

        Ok(AuthContext {
            conn: self.conn,
            user: user,
        })
    }

    ///
    /// Attempt to look up a user (by Clowder username) in the user database.
    ///
    fn lookup_user(&self, clowder_username: &str) -> Result<User, Error> {
        User::with_username(&clowder_username, &self.conn)
             .map_err(Error::DatabaseError)
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
            .and_then(|username| GithubAccount::get(&username, &self.conn)
                                               .map_err(Error::DatabaseError)
                                               .map(|(_, user)| user))
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
    jar.get_private("user").map(|c| jar.remove_private(c));
}

/// Generate a cookie that attests to a logged-in user's username.
pub fn set_user_cookie<'c, S: Into<String>>(mut jar: Cookies, username: S)
    -> Result<(), Error>
{
    let name = username.into();
    let value = format!["{}-{}", name, hmac(&name)?];
    jar.add_private(Cookie::new(String::from("user"), value));

    Ok(())
}

fn check_mac(cookie: Cookie) -> Result<String, Error> {
    let s = cookie.value();

    let idx = s.rfind("-").ok_or(Error::AuthError(format!["no MAC in cookie: {}", s]))?;
    let (uname, mac) = s.split_at(idx);
    let my_mac = hmac(uname)?;

    if my_mac == mac[1..] {
        Ok(uname.to_string())
    } else {
        println!["Cookie authentication failure: {} != {}", my_mac, &mac[1..]];
        Err(Error::AuthRequired)
    }
}

fn hmac(input: &str) -> Result<String, Error> {
    let mut hmac = Hmac::new(Sha512::new(), &*HMAC_KEY);
    hmac.input(input.as_bytes());
    Ok(hmac.result().code().to_hex())
}
