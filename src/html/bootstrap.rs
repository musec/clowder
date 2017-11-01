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


///
/// A Bootstrap modal dialog.
///
/// # Examples
///
/// ```
/// let markup = ModalDialog::new("login")
///     .title("Login required")
///     .body(html! {
///         a href={ OAUTH_URL) "?client_id" = (id) } "Sign in with XXXXX"
///     })
///     .closeable(true)
///     .start_open(true)
///     .render()
///     ;
/// ```
///
pub struct ModalDialog {
    /// The HTML ID used to identify this dialog.
    id: String,

    /// Title to display (if not specified, defaults to HTML id).
    title: Option<String>,

    /// Whether or not the dialog should have an "X" to close it.
    closeable: bool,

    /// Whether or not the dialog should be open from inception.
    start_open: bool,

    /// The body of the dialog to display (if any): a good place for text.
    body: Option<Markup>,

    /// The dialog footer (if any): a good place for response buttons (e.g., "Login", "Cancel").
    footer: Option<Markup>,
}

impl ModalDialog {
    pub fn new<S: Into<String>>(id: S) -> ModalDialog {
        ModalDialog {
            id: id.into(),
            title: None,
            closeable: false,
            start_open: false,
            body: None,
            footer: None,
        }
    }

    pub fn title<S: Into<String>>(mut self, title: S) -> Self {
        self.title = Some(title.into());
        self
    }

    pub fn closeable(mut self, closeable: bool) -> Self {
        self.closeable = closeable;
        self
    }

    pub fn start_open(mut self, start_open: bool) -> Self {
        self.start_open = start_open;
        self
    }

    pub fn body(mut self, body: Markup) -> Self {
        self.body = Some(body);
        self
    }

    pub fn footer(mut self, footer: Markup) -> Self {
        self.footer = Some(footer);
        self
    }
}

impl Render for ModalDialog {
    fn render(&self) -> Markup {
        let label_id = format!["{}_label", self.id];
        let title = if let Some(ref t) = self.title { t } else { &self.id };

        html! {
            div.modal.fade
                id=(self.id)
                role="dialog"
                aria-labelledby=(label_id)
                aria-hidden="true" {

                div.modal-dialog role="document"
                    div.modal-content {
                        div.modal-header {
                            h5.modal-title id=(label_id) (title)

                            @if self.closeable {
                                button.close type="button"
                                    data-dismiss="modal"
                                    aria-label="Close"

                                    span aria-hidden="true" (PreEscaped("&times;"))
                            }
                        }

                        @if let Some(ref body) = self.body.as_ref() {
                            div.modal-body
                                (body)
                        }

                        @if let Some(ref footer) = self.footer.as_ref() {
                            div.modal-footer
                                (footer)
                        }
                    }
            }

            @if self.start_open {
                script type="text/javascript" {
                    "$(document).ready(function () { $('#login').modal('show'); });"
                }
            }
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

    pub fn content(mut self, c: Markup) -> Self {
        self.content = Some(c);
        self
    }

    pub fn flash(mut self, f: Option<FlashMessage>) -> Self {
        self.flash = f;
        self
    }

    pub fn nav(mut self, m: Vec<NavItem>) -> Self {
        self.nav = m;
        self
    }

    pub fn user(mut self, username: &str, display_name: &str) -> Self {
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

                    link rel="stylesheet" href="/css/bootstrap.min.css" /
                    link rel="stylesheet" href="/css/musec.css" /
                    link rel="stylesheet" href="/css/sticky-footer-navbar.css" /
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
                                div class="col-md-11" ""
                                div class="col-md-2" (chrono::Local::now().format("%e %b %Y"))
                            }
                        }
                    }

                    script src="https://cdnjs.cloudflare.com/ajax/libs/tether/1.4.0/js/tether.min.js"
                           integrity="sha384-DztdAPBWPRXSA/3eYEEUWrWCy7G5KFbe8fFjk5JAIxUYHKkDx6Qin1DkWx51bBrb"
                           crossorigin="anonymous" {}

                    script src="//cdn.jsdelivr.net/momentjs/latest/moment.min.js" {}

                    script src="/js/bootstrap.min.js" {}

                    script src="//cdn.jsdelivr.net/bootstrap.daterangepicker/2/daterangepicker.js"
                        {}

                    script src="https://use.fontawesome.com/ff559252db.js" {}

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
