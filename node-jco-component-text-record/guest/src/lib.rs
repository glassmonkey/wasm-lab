#[allow(warnings)]
mod bindings;

use bindings::{Guest, Person, Point};

struct Component;

impl Guest for Component {
    fn greet(name: String) -> String {
        format!("Hello, {}!", name)
    }

    fn upper(s: String) -> String {
        s.to_uppercase()
    }

    fn words(s: String) -> Vec<String> {
        s.split_whitespace().map(String::from).collect()
    }

    fn translate(p: Point, dx: f64, dy: f64) -> Point {
        Point {
            x: p.x + dx,
            y: p.y + dy,
        }
    }

    fn distance(a: Point, b: Point) -> f64 {
        ((a.x - b.x).powi(2) + (a.y - b.y).powi(2)).sqrt()
    }

    fn format_person(p: Person) -> String {
        format!("{} (age {}) [{}]", p.name, p.age, p.tags.join(", "))
    }

    fn make_person(name: String, age: u32, tags: Vec<String>) -> Person {
        Person { name, age, tags }
    }
}

bindings::export!(Component with_types_in bindings);
