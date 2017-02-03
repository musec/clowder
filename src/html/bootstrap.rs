use chrono;
use maud::*;
use rocket::request::FlashMessage;

use html::link::Link;
use html::Context;


pub fn alert<S1, S2>(kind: S1, msg: S2) -> Markup
    where S1: Into<String>, S2: Into<String>
{
    html! {
        div class={ "alert alert-dismissable alert-" (kind.into()) } role="alert" {
            (PreEscaped(msg.into()))
            button.close type="button" data-dismiss="alert" aria-label="Close"
                span aria-hidden="true" (PreEscaped("&times;".to_string()))
        }
    }
}

pub fn callout<S1, S2>(kind: S1, title: S2, content: Markup) -> Markup
    where S1: Into<String>, S2: Into<String>
{
    html! {
        div#flash class={ "mb-3 bs-callout bs-callout-" (kind.into()) } {
            h4 (title.into())
            (content)
        }
    }
}


/// Render a complete HTML page with Bootstrap styling wrapped around the
/// provided content.
pub fn render<S>(title: S, ctx: &Context, flash: Option<FlashMessage>, content: Markup) -> Markup
    where S: Into<String>
{
    html! {
        (DOCTYPE)

        html {
            head {
                meta charset="utf-8" /
                meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" /

                title (title.into())

                link rel="stylesheet" href="/static/css/bootstrap.min.css" /
                link rel="stylesheet" href="/static/css/musec.css" /
                link rel="stylesheet" href="/static/css/sticky-footer-navbar.css" /
            }

            body {
                nav.navbar.navbar-toggleable-md.navbar-light.bg-faded.fixed-top {
                    button.navbar-toggler.navbar-toggler-right type="button"
                           data-toggle="collapse" data-target="#navbarSupportedContent"
                           aria-controls="navbarSupportedContent" aria-expanded="false"
                           aria-label="Toggle navigation" {
                        span.navbar-toggler-icon {}
                    }

                    a class="navbar-brand" href="/" "Clowder"

                    div.collapse.navbar-collapse#navbarSupportedContent
                        ul.navbar-nav.mr-auto {
                            li.nav-item a.nav-link href="/machines" "Machines"
                            li.nav-item a.nav-link href="/reservations" "Reservations"
                        }

                        span style="margin-right: 1em" {
                            (Link::to("/profile", ctx.user.username.as_str()))
                        }

                        img src="https://allendale.engr.mun.ca/musec.png" alt="MUSEC logo"
                            width="40" /
                }

                div.container {
                    div.row div class="col-md-12"
                        // Check for a flash (one-time) message:
                        (match &flash {
                            &Some(ref f) => alert(f.name(), f.msg()),
                            &None => html![],
                        })

                    div.row div class="col-md-12" div.container (content)
                }

                footer.footer {
                    div.container.text-muted {
                        div.row.text-muted {
                            @let now = chrono::Local::now() {
                                div class="col-md-11" ""
                                div class="col-md-2" (now.format("%e %b %Y"))
                            }
                        }
                    }
                }

                script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js" {}
                script src="https://code.jquery.com/jquery-3.1.1.slim.min.js"
                       integrity="sha384-A7FZj7v+d/sdmMqp/nOQwliLvUsJfDHW+k9Omg/a/EheAdgtzNs3hpfag6Ed950n"
                       crossorigin="anonymous" {}
                script src="https://cdnjs.cloudflare.com/ajax/libs/tether/1.4.0/js/tether.min.js"
                       integrity="sha384-DztdAPBWPRXSA/3eYEEUWrWCy7G5KFbe8fFjk5JAIxUYHKkDx6Qin1DkWx51bBrb"
                       crossorigin="anonymous" {}

                script src="/static/js/bootstrap.min.js" {}
            }
        }
    }
}
