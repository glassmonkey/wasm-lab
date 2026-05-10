#[allow(warnings)]
mod bindings;

use bindings::Guest;

struct Component;

impl Guest for Component {
    fn add_s32(a: i32, b: i32) -> i32 {
        a.wrapping_add(b)
    }
    fn mul_s64(a: i64, b: i64) -> i64 {
        a.wrapping_mul(b)
    }
    fn div_f64(a: f64, b: f64) -> f64 {
        a / b
    }
    fn negate_s32(x: i32) -> i32 {
        x.wrapping_neg()
    }
    fn sum_u32(a: u32, b: u32) -> u32 {
        a.wrapping_add(b)
    }
    fn is_positive(x: i32) -> bool {
        x > 0
    }
}

bindings::export!(Component with_types_in bindings);
