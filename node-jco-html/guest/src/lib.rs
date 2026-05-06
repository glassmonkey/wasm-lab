// cargo-component が初回ビルド時に bindings.rs を自動生成する。
// 生成元は wit/world.wit の fetcher world。
#[allow(warnings)]
mod bindings;

use bindings::wasi::http::outgoing_handler;
use bindings::wasi::http::types::{
    Fields, Method, OutgoingRequest, Scheme,
};
use bindings::wasi::io::streams::StreamError;
use bindings::Guest;
use bindings::PageInfo;

struct Component;

impl Guest for Component {
    fn extract(url_str: String) -> Result<PageInfo, String> {
        let url = url::Url::parse(&url_str).map_err(|e| format!("invalid url: {e}"))?;
        let html = http_get(&url)?;
        Ok(parse_html(&html))
    }
}

fn http_get(url: &url::Url) -> Result<String, String> {
    let scheme = match url.scheme() {
        "https" => Scheme::Https,
        "http" => Scheme::Http,
        other => return Err(format!("unsupported scheme: {other}")),
    };
    let authority = url
        .host_str()
        .map(|h| match url.port() {
            Some(p) => format!("{h}:{p}"),
            None => h.to_string(),
        })
        .ok_or_else(|| "missing host".to_string())?;
    let path = if url.path().is_empty() { "/" } else { url.path() };
    let path_query = match url.query() {
        Some(q) => format!("{path}?{q}"),
        None => path.to_string(),
    };

    let headers = Fields::new();
    let req = OutgoingRequest::new(headers);
    req.set_method(&Method::Get).map_err(|_| "set method".to_string())?;
    req.set_scheme(Some(&scheme)).map_err(|_| "set scheme".to_string())?;
    req.set_authority(Some(&authority)).map_err(|_| "set authority".to_string())?;
    req.set_path_with_query(Some(&path_query)).map_err(|_| "set path".to_string())?;

    let resp_future =
        outgoing_handler::handle(req, None).map_err(|e| format!("handle failed: {e:?}"))?;

    // ブロッキング待ち
    resp_future.subscribe().block();

    let resp = resp_future
        .get()
        .ok_or_else(|| "no response".to_string())?
        .map_err(|_| "response future already taken".to_string())?
        .map_err(|e| format!("error response: {e:?}"))?;

    let status = resp.status();
    if status != 200 {
        return Err(format!("HTTP {status}"));
    }

    let body = resp.consume().map_err(|_| "consume body".to_string())?;
    let stream = body.stream().map_err(|_| "open stream".to_string())?;

    let mut buf = Vec::new();
    loop {
        match stream.blocking_read(8 * 1024) {
            Ok(chunk) if chunk.is_empty() => break,
            Ok(chunk) => buf.extend_from_slice(&chunk),
            Err(StreamError::Closed) => break,
            Err(e) => return Err(format!("stream error: {e:?}")),
        }
    }

    String::from_utf8(buf).map_err(|e| format!("utf-8 decode: {e}"))
}

fn parse_html(html: &str) -> PageInfo {
    use scraper::{Html, Selector};

    let doc = Html::parse_document(html);

    let title_sel = Selector::parse("title").expect("static selector");
    let title = doc
        .select(&title_sel)
        .next()
        .map(|el| el.text().collect::<String>().trim().to_string())
        .unwrap_or_default();

    let link_sel = Selector::parse("a[href]").expect("static selector");
    let links: Vec<String> = doc
        .select(&link_sel)
        .filter_map(|el| el.value().attr("href"))
        .map(|s| s.to_string())
        .collect();

    PageInfo { title, links }
}

bindings::export!(Component with_types_in bindings);
