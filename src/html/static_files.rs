use std::fs::File;
use std::io;


#[get("/css/<filename>")]
fn static_css(filename: String) -> io::Result<File> {
    File::open(format!["static/css/{}", filename])
}

#[get("/images/<filename>")]
fn static_images(filename: String) -> io::Result<File> {
    File::open(format!["static/images/{}", filename])
}

#[get("/js/<filename>")]
fn static_js(filename: String) -> io::Result<File> {
    File::open(format!["static/js/{}", filename])
}

