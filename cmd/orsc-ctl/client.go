package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a thin HTTP client for the OpenRSC Admin API. Every response is
// wrapped in an {ok,data,error} envelope; call() unwraps it so callers see only
// the raw "data" (or an error carrying the server's message + HTTP status).
type Client struct {
	base  string
	token string
	http  *http.Client
}

// NewClient builds a Client against base (the /admin/v1 root) authenticating
// with token via the X-Admin-Token header. timeout bounds each request.
func NewClient(base, token string, timeout time.Duration) *Client {
	return &Client{
		base:  strings.TrimRight(base, "/"),
		token: token,
		http:  &http.Client{Timeout: timeout},
	}
}

// envelope is the universal response wrapper. data/error are kept raw because
// their shapes vary per endpoint; error is decoded lazily by errMessage.
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

// call performs one request and unwraps the envelope. method is the HTTP verb,
// path is appended to the base (e.g. "/server/stats" or "/players/foo/mute"),
// query is applied to the URL if non-nil, and body is JSON-marshaled if non-nil
// (with Content-Type application/json). On ok==false or a non-2xx status it
// returns an error including the server's error message and the status; on
// success it returns the raw "data".
func (c *Client) call(method, path string, query url.Values, body any) (json.RawMessage, error) {
	u := c.base + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		rdr = bytes.NewReader(buf)
	}

	req, err := http.NewRequest(method, u, rdr)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("X-Admin-Token", c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Decode the envelope. If the payload isn't a valid envelope, fall back to
	// surfacing the raw body with the status so the user still sees something.
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("%s %s: %s: %s", method, path, resp.Status, strings.TrimSpace(string(raw)))
		}
		return nil, fmt.Errorf("%s %s: decode response: %w", method, path, err)
	}

	if !env.OK || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := errMessage(env.Error)
		if msg == "" {
			msg = resp.Status
		}
		return nil, fmt.Errorf("%s %s: %s (status %s)", method, path, msg, resp.Status)
	}
	return env.Data, nil
}

// errMessage extracts a human-readable message from the envelope's error field,
// which may be a string, a {message,...} object, or a {code,message} object.
func errMessage(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var obj map[string]any
	if json.Unmarshal(raw, &obj) == nil {
		var msg, code string
		if v, ok := obj["message"].(string); ok {
			msg = v
		}
		if v, ok := obj["code"].(string); ok {
			code = v
		}
		switch {
		case code != "" && msg != "":
			return code + ": " + msg
		case msg != "":
			return msg
		case code != "":
			return code
		}
	}
	return strings.TrimSpace(string(raw))
}

// get issues GET path with the given query.
func (c *Client) get(path string, query url.Values) (json.RawMessage, error) {
	return c.call(http.MethodGet, path, query, nil)
}

// post issues POST path with body marshaled to JSON (nil body sends no payload).
func (c *Client) post(path string, body any) (json.RawMessage, error) {
	return c.call(http.MethodPost, path, nil, body)
}

// patch issues PATCH path with body marshaled to JSON.
func (c *Client) patch(path string, body any) (json.RawMessage, error) {
	return c.call(http.MethodPatch, path, nil, body)
}

// del issues DELETE path with an optional JSON body.
func (c *Client) del(path string, body any) (json.RawMessage, error) {
	return c.call(http.MethodDelete, path, nil, body)
}

// printResult pretty-prints a json.RawMessage result, honoring the -json flag
// (which is just pretty-printed compacted-then-indented JSON either way here;
// both modes emit valid JSON, -json being the "machine" form). Use this to
// render whatever call() returns.
func printResult(raw json.RawMessage) error {
	if len(raw) == 0 || string(raw) == "null" {
		// Endpoints with no data payload (e.g. save-all) still succeeded.
		if *jsonOut {
			fmt.Println("null")
		} else {
			fmt.Println("ok")
		}
		return nil
	}
	if *jsonOut {
		// Raw/compact machine output.
		fmt.Println(strings.TrimSpace(string(raw)))
		return nil
	}
	return printJSON(raw)
}

// printJSON pretty-prints any value (or re-indents a json.RawMessage) to stdout.
func printJSON(v any) error {
	var out []byte
	var err error
	if raw, ok := v.(json.RawMessage); ok {
		var buf bytes.Buffer
		if err = json.Indent(&buf, raw, "", "  "); err == nil {
			out = buf.Bytes()
		}
	} else {
		out, err = json.MarshalIndent(v, "", "  ")
	}
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
