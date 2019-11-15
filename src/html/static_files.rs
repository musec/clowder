/*
 * Copyright 2016-2017 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0, <LICENSE-APACHE or
 * http://apache.org/licenses/LICENSE-2.0> or the MIT license <LICENSE-MIT or
 * http://opensource.org/licenses/MIT>, at your option. This file may not be
 * copied, modified, or distributed except according to those terms.
 */

use std::fs::File;
use std::io;

#[get("/css/<filename>")]
pub fn static_css(filename: String) -> io::Result<File> {
    File::open(format!["static/css/{}", filename])
}

#[get("/images/<filename>")]
pub fn static_images(filename: String) -> io::Result<File> {
    File::open(format!["static/images/{}", filename])
}

#[get("/js/<filename>")]
pub fn static_js(filename: String) -> io::Result<File> {
    File::open(format!["static/js/{}", filename])
}
