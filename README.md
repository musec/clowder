## Getting started

### GitHub

Clowder uses OAuth for authentication, with the current provider being GitHub.
At your local site, you need to create a
[GitHub OAuth application](https://github.com/settings/developers)
with your own Client ID and Client Secret.
Set up environment variables containing these values, or put them in a `.env`
file in your source directory:

```sh
echo "CLOWDER_GH_CLIENT_ID=aaaaaaaaaa" > .env
echo "CLOWDER_GH_CLIENT_SECRET=aaaaaaaaaa" >> .env
```


### Rust

For the moment, we depend on crates that depend on Rust nightly.
You will likely want to use [Rustup](https://www.rustup.rs) to install the
nightly version of Rust.


### Database

Clowder requires a database to be created that is accessible to the user running
the service. Using Postgres (at least on FreeBSD), this looks like:

```sh
# service postgresql initdb
# service postgresql start
# su - postgres
$ createuser ${username}    # with a username like, e.g., clowder
$ psql postgres
postgres=# create database clowder;
postgres=# grant all on database clowder to ${username};
```

You should set the database URL in an environment variable, or in a
`.env` file within your source directory:

```sh
$ echo "export DATABASE_URL=postgres://localhost/clowder" >> .env
```

Once the database has been created, we use the
[Diesel](https://crates.io/crates/diesel) ORM to initialize it:

```sh
$ cargo install diesel_cli
$ cd path/to/clowder/source
$ diesel migration run
```


### Clowder

Once Rust and the Clowder database have been set up, you can build and run
Clowder!

```sh
$ cargo build
$ cargo run
```

For development purposes, I like to use
[cargo-watch](https://crates.io/crates/cargo-watch) to rebuild whenever I change
a source file:

```sh
$ cargo watch --ignore '*.swp' --exec run
```

This goes quite nicely with LiveReload.
