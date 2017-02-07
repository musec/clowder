use maud::*;


/// An HTML <input> field
pub struct Input {
    name: String,
    value: Option<String>,
    size: Option<i8>,
    writable: bool,
}

impl Input {
    pub fn new<S: Into<String>>(name: S) -> Input {
        Input {
            name: name.into(),
            value: None,
            size: None,
            writable: true,
        }
    }

    pub fn value<S: Into<String>>(self, value: S) -> Input {
        Input {
            name: self.name,
            value: Some(value.into()),
            size: self.size,
            writable: self.writable,
        }
    }

    pub fn size(self, size: i8) -> Input {
        Input {
            name: self.name,
            value: self.value,
            size: Some(size),
            writable: self.writable,
        }
    }

    pub fn writable(self, writable: bool) -> Input {
        Input {
            name: self.name,
            value: self.value,
            size: self.size,
            writable: writable,
        }
    }
}

impl Render for Input {
    fn render(&self) -> Markup {
        let mut s = format!["<input name=\"{}\"", super::escape(&self.name)];

        if let Some(ref value) = self.value {
            s += &format![" value=\"{}\"", super::escape(value)];
        }

        if let Some(size) = self.size {
            s += &format![" size=\"{}\"", size];
        }

        if !self.writable {
            s += &String::from(" readonly");
        }

        s.push('>');

        PreEscaped(s)
    }
}
