#![feature(plugin)]
#![plugin(maud_macros)]
#![plugin(rocket_codegen)]

extern crate chrono;
extern crate chrono_humanize;
#[macro_use] extern crate diesel;
#[macro_use] extern crate diesel_codegen;
extern crate dotenv;
extern crate maud;
extern crate rocket;

mod db;
mod html;


fn main() {
    rocket::ignite()
           .mount("/", html::all_routes())
           .launch()
           ;
}
