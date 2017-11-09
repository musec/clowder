/*
 * Copyright 2016-2017 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use super::hyper;
use super::rustc_serialize;
use super::url;

use hyper::header;
use hyper_native_tls::NativeTlsClient;
use rustc_serialize::Decodable;
use url::Url;

use super::Error;
use std::env;
use std::io::Read;

const ACCESS_TOKEN_URL: &'static str = "https://github.com/login/oauth/access_token";


///
/// Retrieve a GitHub user's username based on an OAuth code passed to a GitHub callback.
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
/// This function handles steps 4-6, taking an auth code and (if all goes well) returning a
/// GitHub username.
///
pub fn auth_callback(auth_code: String) -> Result<String, Error> {
    OAuthClient::new(env::var("CLOWDER_GH_CLIENT_ID")?)
        ?
        .set_secret(env::var("CLOWDER_GH_CLIENT_SECRET")?)
        .set_oauth_code(auth_code)
        .user()
        .map(|u| u.username().to_string())
}


/// A GitHub OAuth client.
struct OAuthClient {
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
}

impl UserData {
    pub fn username(&self) -> &str {
        &self.login
    }
}


impl OAuthClient {
    pub fn new<S: Into<String>>(id: S) -> Result<OAuthClient, Error> {
        let tls_connector = hyper::net::HttpsConnector::new(NativeTlsClient::new()?);

        Ok(OAuthClient {
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
        where T: Decodable,
              U: hyper::client::IntoUrl
    {
        let access_token = self.access_token()?.to_string();

        let mut response = self.http
            .get(url)
            .header(header::Authorization(header::Bearer { token: access_token }))
            .header(header::UserAgent(String::from("musec/clowder")))
            .send()?;

        let body = response_str(&mut response)?;

        rustc_serialize::json::decode(&body)
            .map_err(|err| format!["failed to parse {}: {}", body, err])
            .map_err(Error::InvalidData)
    }

    ///
    /// Retrieve a GitHub user access token based on an OAuth code.
    /// This implements step five in the overall process (described in `auth_callback`).
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
            }

            "error_description" => {
                return Err(Error::InvalidData(value.to_string()));
            }

            _ => {}
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
