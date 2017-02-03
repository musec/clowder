use chrono;
use maud::*;
use rocket::request::FlashMessage;

use html::link::Link;
use html::Context;


pub fn callout<S1, S2>(kind: S1, title: S2, closeable: bool, content: Markup) -> Markup
    where S1: Into<String>, S2: Into<String>
{
    let k = kind.into();
    let t = title.into();
    let same = &k == &t;

    html! {
        div#flash class={ "mb-3 bs-callout bs-callout-" (k) } {
            @if closeable {
                button type="botton" class="close" aria-label="Close"
                       span aria-hidden="true"
                       onclick="document.getElementById('flash').style['display'] = 'none';"
                        (::maud::PreEscaped("&times;"))
            }

            @if !same { h4 (t) }

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
        html {
            (DOCTYPE)

            head {
                meta charset="utf-8"
                meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no"

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
                        span.navbar-toggler-icon /
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

                        img src="https://allendale.engr.mun.ca/musec.png" 
                            width="40" /
                }

                div.container {
                    div.row {
                        // Check for a flash (one-time) message:
                        (match &flash {
                            &Some(ref f) => callout(f.name(), f.name(), true,
                                                    ::maud::PreEscaped(f.msg().to_string())),
                            &None => html![],
                        })
                    }

                    div.row {
                        div.container (content)
                    }
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

                script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js" /
                script src="https://code.jquery.com/jquery-3.1.1.slim.min.js"
                       integrity="sha384-A7FZj7v+d/sdmMqp/nOQwliLvUsJfDHW+k9Omg/a/EheAdgtzNs3hpfag6Ed950n"
                       crossorigin="anonymous"
                script src="https://cdnjs.cloudflare.com/ajax/libs/tether/1.4.0/js/tether.min.js"
                       integrity="sha384-DztdAPBWPRXSA/3eYEEUWrWCy7G5KFbe8fFjk5JAIxUYHKkDx6Qin1DkWx51bBrb"
                       crossorigin="anonymous"

                script src="/static/js/bootstrap.min.js" /
            }
        }
    }
}
