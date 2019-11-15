/*
 * Copyright 2016-2017, 2019 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use std::env;

use diesel::pg::PgConnection;
use diesel::prelude::*;
use dotenv::dotenv;
use error::Error;

pub mod models;
pub mod schema;

pub fn establish_connection() -> Result<PgConnection, Error> {
    dotenv()?;

    let database_url = match env::var("DATABASE_URL") {
        Ok(url) => Ok(url),
        Err(env::VarError::NotPresent) => Err(Error::config("DATABASE_URL not set")),
        Err(env::VarError::NotUnicode(s)) => {
            Err(Error::config(format!["Invalid DATABASE_URL: {:?}", s]))
        },
    };

    PgConnection::establish(&database_url?)
        .map_err(Error::DatabaseConnectionError)
}
