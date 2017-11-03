use maud::*;


/// An HTML <input> field
pub struct Input {
    name: String,
    class: Option<String>,
    value: Option<String>,
    size: Option<i8>,
    writable: bool,
}

impl Input {
    pub fn new<S: Into<String>>(name: S) -> Input {
        Input {
            name: name.into(),
            class: None,
            value: None,
            size: None,
            writable: true,
        }
    }

    pub fn class<S: Into<String>>(self, class: S) -> Input {
        Input {
            name: self.name,
            class: Some(class.into()),
            value: self.value,
            size: self.size,
            writable: self.writable,
        }
    }

    pub fn value<S: Into<String>>(self, value: S) -> Input {
        Input {
            name: self.name,
            class: self.class,
            value: Some(value.into()),
            size: self.size,
            writable: self.writable,
        }
    }

    pub fn size(self, size: i8) -> Input {
        Input {
            name: self.name,
            class: self.class,
            value: self.value,
            size: Some(size),
            writable: self.writable,
        }
    }

    pub fn writable(self, writable: bool) -> Input {
        Input {
            name: self.name,
            class: self.class,
            value: self.value,
            size: self.size,
            writable: writable,
        }
    }
}

impl Render for Input {
    fn render(&self) -> Markup {
        let mut s = format!["<input name=\"{}\"", super::escape(&self.name)];

        if let Some(ref class) = self.class {
            s += &format![" class=\"{}\"", super::escape(class)];
        }

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



/// An HTML <select> field
pub struct Select {
    name: String,
    options: Vec<SelectOption>,
    multiple: bool,
}

impl Select {
    pub fn new<S: Into<String>>(name: S) -> Select {
        Select {
            name: name.into(),
            options: Vec::new(),
            multiple: false,

        }
    }

    pub fn multiple(mut self, mult: bool) -> Self {
        self.multiple = mult;
        self
    }

    pub fn set_options(mut self, opts: Vec<SelectOption>) -> Self {
        self.options = opts;
        self
    }
}

impl Render for Select {
    fn render(&self) -> Markup {
        html! {
            select name=(self.name) multiple?[self.multiple] {
                @for o in &self.options {
                    (o)
                }
            }
        }
    }
}

pub struct SelectOption {
    name: String,
    label: String,
    selected: bool,
}

impl SelectOption {
    pub fn new<S1, S2>(name: S1, label: S2) -> SelectOption
        where S1: Into<String>, S2: Into<String>
    {
        SelectOption {
            name: name.into(),
            label: label.into(),
            selected: false,
        }
    }

    pub fn selected(self, s: bool) -> SelectOption {
        SelectOption {
            name: self.name,
            label: self.label,
            selected: s,
        }
    }
}

impl<'a> From<(&'a str, &'a str)> for SelectOption {
    fn from(tuple: (&str, &str)) -> SelectOption {
        SelectOption {
            name: tuple.0.to_string(),
            label: tuple.1.to_string(),
            selected: false,
        }
    }
}

impl Render for SelectOption {
    fn render(&self) -> Markup {
        html! {
            option value=(self.name) selected?[self.selected] (self.label)
        }
    }
}


/// An HTML form submission button.
pub struct SubmitButton {
    label: String,
}

impl SubmitButton {
    pub fn new() -> SubmitButton {
        SubmitButton {
            label: String::from("Submit")
        }
    }

    pub fn label<S: Into<String>>(self, l: S) -> SubmitButton {
        SubmitButton {
            label: l.into(),
        }
    }
}

impl Render for SubmitButton {
    fn render(&self) -> Markup {
        html! {
            input type="submit" value=(self.label) (super::route_prefix())
        }
    }
}
