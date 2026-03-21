package main

import (
	"bytes"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"snippetbox.kita-ikuyoo.miku01/internal/models/mocks"
	"strings"
	"testing"
	"time"
)

func newTestApplication(t *testing.T) *application {
	// Create an instance of the template cache.
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}
	// And a form decoder.
	formDecoder := form.NewDecoder()
	// And a session manager instance. Note that we use the same settings as
	// production, except that we *don't* set a Store for the session manager.
	// If no store is set, the SCS package will default to using a transient
	// in-memory store, which is ideal for testing purposes.
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true
	return &application{
		logger:         slog.New(slog.DiscardHandler),
		snippets:       &mocks.SnippetModel{}, // Use the mock.
		users:          &mocks.UserModel{},    // Use the mock.
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
}

// Define a custom testServer type which embeds an httptest.Server instance.
type testServer struct {
	*httptest.Server
}

// Define a testResponse struct to hold data about responses from the test
// server. Note that this struct includes fields for the HTTP response headers
// and cookies, as well as the status code and body.
type testResponse struct {
	status  int
	headers http.Header
	cookies []*http.Cookie
	body    string
}

// Implement a get() method on our custom testServer type. This makes a GET
// request to a given url path using the test server client and it returns a
// testResponse struct containing the response data.
func (ts *testServer) get(t *testing.T, urlPath string) testResponse {
	req, err := http.NewRequest(http.MethodGet, ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return testResponse{
		status:  res.StatusCode,
		headers: res.Header,
		cookies: res.Cookies(),
		body:    string(bytes.TrimSpace(body)),
	}
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	// Initialize the test server as normal.
	ts := httptest.NewTLSServer(h)
	// Initialize a new cookie jar.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	// Add the cookie jar to the test server client. Any response cookies will
	// now be stored in the jar and sent with subsequent requests when using
	// this client.
	ts.Client().Jar = jar
	// Prevent the test server client from following redirects by setting a
	// custom CheckRedirect function. This function runs whenever a 3xx
	// response is received. By returning http.ErrUseLastResponse, it tells
	// the client to stop and immediately return the received response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &testServer{ts}
}

func (ts *testServer) resetClientCookieJar(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar
}

func extractCSRFToken(t *testing.T, body string) string {
	// Define a regular expression which captures the CSRF token value from the
	// HTML for our user signup page.
	csrfTokenRX := regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)
	// Use the FindStringSubmatch method to extract the token from the HTML body.
	// Note that this returns an slice with the entire matched pattern in the
	// first position, and the values of any captured data in the subsequent
	// positions.
	// matches[0] = "<input type='hidden' name='csrf_token' value='abc123'>"
	// matches[1] = "abc123"
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}
	return html.UnescapeString(matches[1])
}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) testResponse {
	req, err := http.NewRequest(http.MethodPost, ts.URL+urlPath, strings.NewReader(form.Encode()))
	//req_body, err := io.ReadAll(req.Body)
	//log.Print(req)
	//log.Print(string(req_body))
	if err != nil {
		t.Fatal(err)
	}
	// Set the appropriate Content-Type header for form data and the Sec-Fetch-Site
	// header.
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	res, err := ts.Client().Do(req)
	//log.Print(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	return testResponse{
		status:  res.StatusCode,
		headers: res.Header,
		cookies: res.Cookies(),
		body:    string(bytes.TrimSpace(body)),
	}
}
