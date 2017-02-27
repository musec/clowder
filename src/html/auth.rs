use db::models::*;
use db::schema::*;
use diesel::*;
use diesel::pg::PgConnection as Connection;
use rocket::*;
use rocket::http::Cookie;

use html::error::Error;


/// Authenticate a user request, returning either a User or an Error.
pub fn authenticate(req: &Request, conn: &Connection) -> Result<User, Error> {
    req.cookies()
        .find("username")
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
pub fn user_cookie<'c>(username: String) -> Cookie<'c>
{
    Cookie::new(String::from("username"), username)
}
