#[allow(warnings)]
mod bindings;

use bindings::Guest;
use wit_bindgen_rt::async_support::{error_context_new, ErrorContext};

struct Component;

impl Guest for Component {
    fn try_divide(a: i32, b: i32) -> Result<i32, ErrorContext> {
        if b == 0 {
            Err(error_context_new("division by zero"))
        } else {
            Ok(a / b)
        }
    }
}

bindings::export!(Component with_types_in bindings);
