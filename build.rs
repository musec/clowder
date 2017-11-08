extern crate vergen;

use std::io::{self, Write};

fn main() {
    vergen::vergen(vergen::SEMVER)
           .or_else(|err| writeln![io::stderr(), "unable to generate version string: {:?}", err])
           .expect("failed to print error message for version string generation error");
}
