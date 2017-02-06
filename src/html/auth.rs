use std::error::Error as StdError;

use db::models::*;
use db::schema::*;
use diesel::*;
use diesel::pg::PgConnection as Connection;
use rocket::*;
use rocket::http::hyper::header::{Authorization, Basic, Header};

use html::error::Error;


/// Authenticate a user request, returning either a User or an Error.
pub fn authenticate(req: &Request, conn: &Connection) -> Result<User, Error> {
    let cookies = req.cookies();
    let headers = req.headers();

    cookies.find("username")
        .map(|cookie| cookie.value)
        .ok_or(Error::AuthRequired)
        .or(
            headers.get_one("Authorization")
                .ok_or(Error::AuthRequired)
                .map(|s| s.as_bytes().to_vec())
                .and_then(|raw: Vec<u8>| {
                    Header::parse_header(&[raw])
                        .map_err(|e| e.description().to_string())
                        .map_err(Error::BadRequest)
                })
                .and_then(authenticate_header)
        )
        .and_then(|ref uname| {
            use self::users::dsl::*;

            users.filter(username.eq(uname))
                .first(conn)
                .map_err(Error::DatabaseError)
        })
}



/// Check a user's credentials as embodied in an HTTP Basic Authorization header.
fn authenticate_header(auth_header: Authorization<Basic>) -> Result<String, Error> {
    // TODO: actual authentication!
    println!["auth_header: {:?}", auth_header];
    Ok(auth_header.username.clone())
}
