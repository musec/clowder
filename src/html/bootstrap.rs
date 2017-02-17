use chrono;
use maud::*;
use rocket::request::FlashMessage;


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


pub enum NavItem {
    Link {
        href: String,
        text: String,
    },
}

impl NavItem {
    pub fn link<S1, S2>(href: S1, text: S2) -> NavItem
        where S1: Into<String>, S2: Into<String>
    {
        NavItem::Link { href: href.into(), text: text.into() }
    }
}

impl Render for NavItem {
    fn render(&self) -> Markup {
        match self {
            &NavItem::Link { ref href, ref text } => html! {
                li.nav-item a.nav-link href=(href) (text)
            },
        }
    }
}


pub struct Page {
    title: String,
    content: Option<Markup>,
    flash: Option<FlashMessage>,
    nav: Vec<NavItem>,
    user: Option<(String, String)>,
}

impl Page {
    pub fn new<S: Into<String>>(title: S) -> Page {
        Page {
            title: title.into(),
            content: None,
            flash: None,
            nav: vec![],
            user: None,
        }
    }

    pub fn content(&mut self, c: Markup) -> &mut Page {
        self.content = Some(c);
        self
    }

    pub fn flash(&mut self, f: Option<FlashMessage>) -> &mut Page {
        self.flash = f;
        self
    }

    pub fn nav(&mut self, m: Vec<NavItem>) -> &mut Page {
        self.nav = m;
        self
    }

    pub fn user(&mut self, username: &str, display_name: &str) -> &mut Page {
        self.user = Some((username.into(), display_name.into()));
        self
    }
}

impl Render for Page {
    fn render(&self) -> Markup {
        html! {
            (DOCTYPE)

            html {
                head {
                    meta charset="utf-8" /
                    meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" /

                    title (format!["Clowder: {}", self.title])

                    link rel="stylesheet" href="/static/css/bootstrap.min.css" /
                    link rel="stylesheet" href="/static/css/musec.css" /
                    link rel="stylesheet" href="/static/css/sticky-footer-navbar.css" /
                    link rel="stylesheet" href="//cdn.jsdelivr.net/bootstrap.daterangepicker/2/daterangepicker.css" /

                    script src="https://code.jquery.com/jquery-3.1.1.slim.min.js"
                           integrity="sha384-A7FZj7v+d/sdmMqp/nOQwliLvUsJfDHW+k9Omg/a/EheAdgtzNs3hpfag6Ed950n"
                           crossorigin="anonymous" {}
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
                            ul.navbar-nav.mr-auto
                                @for ref m in &self.nav { (m) }

                            @if let Some((ref username, ref display_name)) = self.user {
                                div.dropdown {
                                    a.btn.dropdown-toggle#userDropdown
                                        href="#"
                                        data-toggle="dropdown" data-target="fubar"
                                        aria-haspopup="true" aria-expanded="false"
                                        (display_name)

                                    div.dropdown-menu.dropdown-menu-right#fubar aria-labelledby="userDropdown" {
                                        h6.dropdown-header (username)
                                        a.dropdown-item href={ "/user/" (username) } "Profile"
                                        div.dropdown-divider {}
                                        a.dropdown-item href="/logout" "Log out"
                                    }
                                }
                            }

                            img src="/images/logo.png" alt="Clowder logo"
                                width="40" /
                    }

                    div.container {
                        div.row div class="col-md-12"
                            // Check for a flash (one-time) message:
                            (match self.flash {
                                Some(ref f) => alert(f.name(), f.msg()),
                                None => html![],
                            })

                        @if let Some(ref content) = self.content {
                            div.row div class="col-md-12" div.container (content)
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

                    script src="https://cdnjs.cloudflare.com/ajax/libs/tether/1.4.0/js/tether.min.js"
                           integrity="sha384-DztdAPBWPRXSA/3eYEEUWrWCy7G5KFbe8fFjk5JAIxUYHKkDx6Qin1DkWx51bBrb"
                           crossorigin="anonymous" {}

                    script src="//cdn.jsdelivr.net/momentjs/latest/moment.min.js" {}

                    script src="/static/js/bootstrap.min.js" {}

                    script src="//cdn.jsdelivr.net/bootstrap.daterangepicker/2/daterangepicker.js"
                        {}

                    (PreEscaped("<script>
                    $('input.daterange').daterangepicker({
                        autoApply: true,
                        locale: {
                            format: 'hh:mmZ D MMM YYYY'
                        },
                        showDropdowns: true,
                        showToday: true,
                        startDate: moment(),
                        timePicker: true,
                        timePicker24Hour: true,
                        timePickerIncrement: 15
                    })
                    </script>"))
                }
            }
        }
    }
}
