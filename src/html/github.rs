use super::hyper;
use super::native_tls;
use super::rustc_serialize;
use super::url;

use hyper::header;
use hyper_native_tls::NativeTlsClient;
use rustc_serialize::Decodable;
use url::Url;

use std::error::Error as StdError;
use std::fmt;
use std::io::Read;

const ACCESS_TOKEN_URL: &'static str = "https://github.com/login/oauth/access_token";




/// A GitHub OAuth client.
pub struct Client {
    /// Client ID: identifies the OAuth application.
    id: String,

    /// HTTP client.
    http: hyper::Client,

    /// Client secret: used to prove that we are the client we claim to be.
    secret: Option<String>,

    /// Capabilitity representing delegation of some rights to this client.
    oauth_code: Option<String>,

    /// GitHub access token for a particular user, derived from an OAuth authentication session.
    token: Option<String>,

    /// Information about the user we are operating on behalf of.
    user: Option<UserData>,
}

/// Information about a GitHub user, retreived from the GitHub API.
#[derive(Clone, Debug, RustcDecodable)]
pub struct UserData {
    /// GitHub username (no `@` symbol).
    login: String,

    /// URL for user avatar, if set.
    avatar_url: Option<String>,

    /// User email address: not an `Option` because we always request the `email` OAuth scope.
    email: String,
}

impl UserData {
    pub fn email(&self) -> &str {
        &self.email
    }

    pub fn username(&self) -> &str {
        &self.login
    }
}


impl Client {
    pub fn new<S: Into<String>>(id: S) -> Result<Client, Error> {
        let tls_connector = hyper::net::HttpsConnector::new(NativeTlsClient::new()?);

        Ok(Client {
            id: id.into(),
            secret: None,
            http: hyper::Client::with_connector(tls_connector),
            oauth_code: None,
            token: None,
            user: None,
        })
    }

    pub fn set_secret<S: Into<String>>(mut self, secret: S) -> Self {
        self.secret = Some(secret.into());
        self
    }

    pub fn set_oauth_code<S: Into<String>>(mut self, code: S) -> Self {
        self.oauth_code = Some(code.into());
        self
    }

    pub fn user(&mut self) -> Result<UserData, Error> {
        if self.user.is_none() {
            let url = Url::parse("https://api.github.com/user").expect("bad GitHub URI");
            let user: UserData = self.query(url)?;

            self.user = Some(user);
        }

        self.user
            .as_ref()
            .map(Clone::clone)
            .map(Ok)
            .expect("UNREACHABLE: self.user was just set; should not be None")
    }

    fn access_token(&mut self) -> Result<&str, Error> {
        if let Some(ref token) = self.token {
            Ok(token)
        } else {
            self.retrieve_access_token()
        }
    }

    fn query<T, U>(&mut self, url: U) -> Result<T, Error>
        where T: Decodable, U: hyper::client::IntoUrl
    {
        let access_token = self.access_token()?.to_string();

        let mut response = self.http.get(url)
                .header(header::Authorization(header::Bearer { token: access_token }))
                .header(header::UserAgent(String::from("musec/clowder")))
                .send()?
                ;

        let body = response_str(&mut response)?;

        rustc_serialize::json::decode(&body).map_err(Error::from)
    }

    ///
    /// Retrieve a user access token based on an OAuth code passed to a GitHub callback.
    ///
    /// The flow of control in a GitHub OAuth authentication session is:
    ///
    /// 1. user visits website,
    /// 2. website redirects user to GitHub auth page (with application/client ID),
    /// 3. user authorizes application/client's requested access scope,
    /// 4. GitHub redirects user to an application callback URL, passing a code in the GET query,
    /// 5. application exchanges code, client ID and client secret for a GitHub access token, then
    /// 6. application uses access token in API queries (e.g., for user details).
    ///
    /// This method implements step five in this process.
    ///
    fn retrieve_access_token(&mut self) -> Result<&str, Error> {
        let secret = self.secret
                .as_ref()
                .ok_or(Error::AuthError(String::from("no GitHub client secret has been set")))?;

        let code = self.oauth_code
                .as_ref()
                .ok_or(Error::AuthError(String::from("no OAuth code has been set")))?;

        let form_data = url::form_urlencoded::Serializer::new(String::new())
            .append_pair("client_id", &self.id)
            .append_pair("client_secret", secret)
            .append_pair("code", code)
            .finish();

        let mut response = try! {
            self.http.post(url::Url::parse(ACCESS_TOKEN_URL).expect("malformed GitHub URI"))
                .body(&form_data)
                .send()
        };

        let token = response_str(&mut response).and_then(access_token)?;

        self.token = Some(token);
        Ok(self.token.as_ref().expect("Token was just set to Some(token), should not ever be None"))
    }
}


fn access_token(form_body: String) -> Result<String, Error> {
    for (key, value) in url::form_urlencoded::parse(form_body.as_bytes()) {
        match &*key {
            "access_token" => {
                return Ok(value.to_string());
            },

            "error_description" => {
                return Err(Error::InvalidData(value.to_string()));
            },

            _ => {},
        };
    }

    Err(Error::InvalidData(format!["no access token in GitHub response: '{}'", form_body]))
}

fn response_str(response: &mut hyper::client::response::Response) -> Result<String, Error> {
    let mut body = String::new();

    response.read_to_string(&mut body)
            .map(|_| body)
            .map_err(|e| Error::InvalidData(format!["invalid response from GitHub: {}", e]))
}


#[derive(Debug)]
pub enum Error {
    /// Error authenticating user.
    AuthError(String),

    /// We received invalid date from somewhere.
    InvalidData(String),

    /// There was a problem communicating with a remote host.
    NetError(hyper::Error),

    /// Error instantiating or using the platform-native TLS client.
    TLSError(native_tls::Error),
}

impl StdError for Error {
    fn description(&self) -> &str {
        match self {
            &Error::AuthError(ref msg) => msg,
            &Error::InvalidData(ref msg) => msg,
            &Error::NetError(ref e) => e.description(),
            &Error::TLSError(ref e) => e.description(),
        }
    }

    fn cause(&self) -> Option<&StdError> {
        match self {
            &Error::NetError(ref e) => Some(e),
            &Error::TLSError(ref e) => Some(e),
            _ => None,
        }
    }
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            &Error::AuthError(ref msg) => write![f, "Authorization error: {}", msg],
            &Error::InvalidData(ref msg) => write![f, "Invalid data: {}", msg],
            &Error::NetError(ref e) => write![f, "Network error: {}", e],
            &Error::TLSError(ref e) => write![f, "TLS (secure network communication) error: {}", e],
        }
    }
}

impl From<hyper::Error> for Error {
    fn from(err: hyper::Error) -> Error {
        Error::NetError(err)
    }
}

impl From<native_tls::Error> for Error {
    fn from(err: native_tls::Error) -> Error {
        Error::TLSError(err)
    }
}

impl From<rustc_serialize::json::DecoderError> for Error {
    fn from(err: rustc_serialize::json::DecoderError) -> Error {
        Error::InvalidData(format!["JSON error: {}", err.description()])
    }
}
