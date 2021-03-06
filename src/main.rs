/*
 * Copyright 2016-2018 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

#![feature(decl_macro)]
#![feature(plugin)]
#![feature(proc_macro_hygiene)]
#![recursion_limit = "128"]

extern crate chrono;
extern crate chrono_humanize;
extern crate crypto;
#[macro_use]
extern crate diesel;
extern crate dotenv;
extern crate hyper;
extern crate hyper_native_tls;
extern crate itertools;
#[macro_use]
extern crate log;
extern crate marksman_escape;
extern crate maud;
extern crate native_tls;
extern crate rand;
#[macro_use]
extern crate rocket;
extern crate rustc_serialize;
extern crate url;

use std::env;

mod db;
mod error;
mod html;

fn getenv(name: &str) -> Result<String, error::Error> {
    use error::Error;

    match env::var(name) {
        Ok(url) => Ok(url),
        Err(env::VarError::NotPresent) => {
            Err(Error::ConfigError(format!["{} not set", name]))
        },
        Err(env::VarError::NotUnicode(s)) => {
            Err(Error::ConfigError(format!["Invalid value for {}: {:?}", name, s]))
        },
    }
}

fn main() {
    dotenv::dotenv().expect("Failed to parse .env");
    let route_prefix = env::var("CLOWDER_PREFIX").unwrap_or(String::from("/"));

    rocket::ignite()
        .register(html::error_catchers())
        .mount(&route_prefix, html::all_routes())
        .launch();
}
