/*
 * Copyright 2017 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

extern crate vergen;

use std::io::{self, Write};

fn main() {
    vergen::vergen(vergen::SEMVER)
        .or_else(|err| writeln![io::stderr(), "unable to generate version string: {:?}", err])
        .expect("failed to print error message for version string generation error");
}
