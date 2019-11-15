/*
 * Copyright 2016-2017, 2019 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use diesel::pg::PgConnection;
use diesel::prelude::*;
use error::Error;

pub mod models;
pub mod schema;

pub fn establish_connection() -> Result<PgConnection, Error> {
    PgConnection::establish(&super::getenv("DATABASE_URL")?)
        .map_err(Error::DatabaseConnectionError)
}
