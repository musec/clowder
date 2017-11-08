#![feature(custom_derive)]
#![feature(plugin)]
#![feature(proc_macro)]
#![plugin(rocket_codegen)]
#![recursion_limit="128"]

extern crate chrono;
extern crate chrono_humanize;
extern crate crypto;
#[macro_use] extern crate diesel;
#[macro_use] extern crate diesel_codegen;
extern crate dotenv;
extern crate hyper;
extern crate hyper_native_tls;
extern crate itertools;
extern crate maud;
extern crate marksman_escape;
extern crate native_tls;
extern crate rand;
extern crate rocket;
extern crate rustc_serialize;
extern crate url;

use std::env;

mod db;
mod html;


fn main() {
    dotenv::dotenv().ok();
    let route_prefix = env::var("CLOWDER_PREFIX").unwrap_or(String::from("/"));

    rocket::ignite()
           .catch(html::error_catchers())
           .mount(&route_prefix, html::all_routes())
           .launch()
           ;
}
