use crypto::hmac::Hmac;
use crypto::mac::Mac;
use crypto::sha2::Sha512;
use db::models::*;
use db::schema::*;
use diesel::*;
use diesel::pg::PgConnection as Connection;
use rand::Rng;
use rand::os::OsRng;
use rocket::http::{Cookie,Cookies};
use rustc_serialize::hex::ToHex;

use html::error::Error;

lazy_static! {
    /// A random HMAC key that is regenerated on every server restart.
    static ref HMAC_KEY: [u8; 32] = OsRng::new().expect("Unable to open system RNG").gen();
}

/// Authenticate a user request, returning either a User or an Error.
pub fn authenticate(mut jar: Cookies, conn: &Connection) -> Result<User, Error> {
    jar.get_private("user")
        .map(|cookie| cookie.value().to_string())
        .ok_or(Error::AuthRequired)
        .and_then(|value| {
            let idx = value.rfind("-").ok_or(
                Error::AuthError(format!["no mac found in cookie: {}", value]))?;

            let (uname, mac) = value.split_at(idx);
            let my_mac = hmac(uname)?;

            if my_mac != mac[1..] {
                println!["Cookie authentication failure: {} != {}", my_mac, &mac[1..]];
                return Err(Error::AuthRequired);
            }

            Ok(uname.to_string())
        })
        .and_then(|uname| {
            use self::users::dsl::*;

            users.filter(username.eq(uname))
                .first(conn)
                .map_err(Error::DatabaseError)
        })
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

fn hmac(input: &str) -> Result<String, Error> {
    let mut hmac = Hmac::new(Sha512::new(), &*HMAC_KEY);
    hmac.input(input.as_bytes());
    Ok(hmac.result().code().to_hex())
}
