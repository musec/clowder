/*
 * Copyright 2016-2019 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use super::bootstrap;

use maud;
use maud::{html, Render};
use rocket;

use std::env;

impl Into<bootstrap::Page> for super::Error {
    fn into(self) -> bootstrap::Page {
        bootstrap::Page::new(self.kind())
            .content(html! {
                h1 { (self.kind()) }
                h2 { (self) }
            })
            .link_prefix(super::route_prefix())
    }
}

/// The error catcher for unauthorized accesses prompts for HTTP basic authentication.
#[catch(401)]
pub fn unauthorized(_req: &rocket::Request) -> bootstrap::Page {
    const OAUTH_URL: &'static str = "https://github.com/login/oauth/authorize";

    let content = match env::var("CLOWDER_GH_CLIENT_ID") {
        Ok(ref id) => bootstrap::ModalDialog::new("login")
            .title("Login required")
            .body(html! {
                p {
                    a.btn.btn-secondary.large href={ (OAUTH_URL) "?client_id=" (id) } {
                        i.fa.fa-github aria-hidden="true" {}
                        (maud::PreEscaped("&nbsp;"))
                        "Sign in with GitHub"
                    }
                }
            })
            .footer(html! {
                p.footnote {
                    "Contact "
                    a href="https://www.engr.mun.ca/~anderson/" { "Jonathan Anderson" }
                    " for authorization to use this system."
                }
            })
            .closeable(false)
            .start_open(true)
            .render(),

        Err(ref e) => {
            error!["Error in GitHub client ID: {:?}", e];

            html! {
                h1 { "Login error" }
                p { "Internal server error in GitHub authentication" }
            }
        }
    };

    bootstrap::Page::new("401 Unauthorized")
        .content(content)
        .link_prefix(super::route_prefix())
}

/// 403 forbidden means that (re-)authenticating won't help.
#[catch(403)]
pub fn forbidden() -> bootstrap::Page {
    bootstrap::Page::new("403 Forbidden")
        .content(html! {
            h2 { ("403 Forbidden") }
            p { "You are not authorized to access this resource." }
        })
        .link_prefix(super::route_prefix())
}

/// The 404 handler renders a slightly nicer-looking page than the stock Rocket handler.
#[catch(404)]
pub fn not_found(req: &rocket::Request) -> bootstrap::Page {
    bootstrap::Page::new("404 Not Found")
        .content(html! {
            h2 { ("404 Not Found") }
            p { "The resource " code { (req.uri()) } " could not be found." }
        })
        .link_prefix(super::route_prefix())
}

/// The 500 ISE (Internal Server Error) handler doesn't provide any more information than the
/// stock Rocket handler, but it also looks nicer.
#[catch(500)]
pub fn internal_server_error(_req: &rocket::Request) -> bootstrap::Page {
    bootstrap::Page::new("500 Internal Server Error")
        .content(html! {
            h2 { ("500 Internal Server Error") }
            p { "There is an error in Clowder!" }
        })
        .link_prefix(super::route_prefix())
}
