#[allow(warnings)]
mod bindings;

use bindings::exports::wasi::http::incoming_handler::Guest;
use bindings::wasi::http::types::{
    Fields, IncomingRequest, Method, OutgoingBody, OutgoingResponse, ResponseOutparam,
};

struct Component;

impl Guest for Component {
    fn handle(req: IncomingRequest, out: ResponseOutparam) {
        let method = req.method();
        let path = req.path_with_query().unwrap_or_else(|| "/".to_string());

        // 単純なルーティング。テストはこの分岐を網羅する想定。
        let (status, body) = match (&method, path.as_str()) {
            (Method::Get, "/hello") => (200u16, b"Hello, world!".to_vec()),
            (Method::Get, "/empty") => (200u16, Vec::new()),
            (Method::Get, p) if p.starts_with("/echo-path") => (200u16, p.as_bytes().to_vec()),
            (Method::Get, "/teapot") => (418u16, b"I'm a teapot".to_vec()),
            _ => (404u16, b"not found".to_vec()),
        };

        let headers = Fields::new();
        headers
            .set(&"content-type".to_string(), &[b"text/plain; charset=utf-8".to_vec()])
            .ok();

        let resp = OutgoingResponse::new(headers);
        resp.set_status_code(status).ok();

        let body_resource = resp.body().expect("response body unavailable");

        // ResponseOutparam::set を先に呼ぶのが wasi-http の規約
        // (set 前に body 書き込みは UB)
        ResponseOutparam::set(out, Ok(resp));

        if !body.is_empty() {
            let stream = body_resource.write().expect("body write stream");
            stream.blocking_write_and_flush(&body).ok();
            drop(stream);
        }
        OutgoingBody::finish(body_resource, None).ok();
    }
}

bindings::export!(Component with_types_in bindings);
