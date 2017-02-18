#![feature(custom_derive)]
#![feature(plugin)]
#![plugin(maud_macros)]
#![plugin(rocket_codegen)]

extern crate chrono;
extern crate chrono_humanize;
#[macro_use] extern crate diesel;
#[macro_use] extern crate diesel_codegen;
extern crate dotenv;
extern crate maud;
extern crate marksman_escape;
extern crate rocket;

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
