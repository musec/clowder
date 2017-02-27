#![feature(custom_derive)]
#![feature(plugin)]
#![plugin(maud_macros)]
#![plugin(rocket_codegen)]

extern crate chrono;
extern crate chrono_humanize;
extern crate crypto;
#[macro_use] extern crate diesel;
#[macro_use] extern crate diesel_codegen;
extern crate dotenv;
extern crate hyper;
extern crate hyper_native_tls;
#[macro_use] extern crate lazy_static;
extern crate maud;
extern crate marksman_escape;
extern crate native_tls;
extern crate rand;
extern crate rocket;
extern crate rustc_serialize;
extern crate url;

mod db;
mod html;


fn main() {
    dotenv::dotenv().ok();
    rocket::ignite()
           .catch(html::error_catchers())
           .mount("/", html::all_routes())
           .launch()
           ;
}
