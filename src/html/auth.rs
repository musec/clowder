use db::models::*;
use db::schema::*;
use diesel::*;
use diesel::pg::PgConnection as Connection;
use rocket::http::{Cookie,CookieJar};

use html::error::Error;


/// Authenticate a user request, returning either a User or an Error.
pub fn authenticate(jar: &CookieJar, conn: &Connection) -> Result<User, Error> {
    jar.find("username")
        .map(|cookie| cookie.value().to_string())
        .ok_or(Error::AuthRequired)
        .and_then(|ref uname| {
            use self::users::dsl::*;

            users.filter(username.eq(uname))
                .first(conn)
                .map_err(Error::DatabaseError)
        })
}

/// Generate a cookie that attests to a logged-in user's username.
pub fn set_user_cookie<'c, S: Into<String>>(jar: &CookieJar, username: S) {
    jar.add(Cookie::new(String::from("username"), username.into()))
}
