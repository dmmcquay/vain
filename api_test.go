package vain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const window = 5 * time.Minute

func TestAdd(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)
	tok, err := db.addUser("sm@example.org")
	if err != nil {
		t.Error("failure to add user: %v", err)
	}

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("couldn't GET: %v", err)
	}
	resp.Body.Close()

	if got, want := len(db.Pkgs()), 0; got != want {
		t.Fatalf("started with something in it; got %d, want %d", got, want)
	}

	{
		bad := ts.URL
		body := strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`)
		req, err := http.NewRequest("POST", bad, body)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("couldn't POST: %v", err)
		}
		if got, want := resp.StatusCode, http.StatusBadRequest; got != want {
			buf := &bytes.Buffer{}
			io.Copy(buf, resp.Body)
			t.Logf("%s", buf.Bytes())
			t.Fatalf("bad request got incorrect status: got %d, want %d", got, want)
		}
		resp.Body.Close()

		if got, want := len(db.Pkgs()), 0; got != want {
			t.Fatalf("started with something in it; got %d, want %d", got, want)
		}
	}

	{
		u := fmt.Sprintf("%s/%s", ts.URL, prefix["pkgs"])
		resp, err := http.Get(u)
		if err != nil {
			t.Error(err)
		}
		buf := &bytes.Buffer{}
		io.Copy(buf, resp.Body)
		pkgs := []Package{}
		if err := json.NewDecoder(buf).Decode(&pkgs); err != nil {
			t.Fatalf("problem parsing json: %v, \n%q", err, buf)
		}
		if got, want := len(pkgs), 0; got != want {
			t.Fatalf("should have empty pkg list; got %d, want %d", got, want)
		}
	}

	{

		u := fmt.Sprintf("%s/foo", ts.URL)
		body := strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`)
		req, err := http.NewRequest("POST", u, body)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("problem performing request: %v", err)
		}
		buf := &bytes.Buffer{}
		io.Copy(buf, resp.Body)
		t.Logf("%v", buf)
		resp.Body.Close()

		if got, want := len(db.Pkgs()), 1; got != want {
			t.Fatalf("pkgs should have something in it; got %d, want %d", got, want)
		}
		t.Logf("packages: %v", db.Pkgs())

		ur, err := url.Parse(ts.URL)
		if err != nil {
			t.Error(err)
		}

		good := fmt.Sprintf("%s/foo", ur.Host)

		if !db.PackageExists(good) {
			t.Fatalf("did not find package for %s; should have posted a valid package", good)
		}
		p, err := db.Package(good)
		t.Logf("%+v", p)
		if err != nil {
			t.Fatalf("problem getting package: %v", err)
		}
		if got, want := p.Path, good; got != want {
			t.Fatalf("package name did not go through as expected; got %q, want %q", got, want)
		}
		if got, want := p.Repo, "https://s.mcquay.me/sm/vain"; got != want {
			t.Fatalf("repo did not go through as expected; got %q, want %q", got, want)
		}
		if got, want := p.Vcs, "git"; got != want {
			t.Fatalf("Vcs did not go through as expected; got %q, want %q", got, want)
		}
	}

	resp, err = http.Get(ts.URL + "?go-get=1")
	if err != nil {
		t.Fatalf("couldn't GET: %v", err)
	}
	defer resp.Body.Close()
	if want := http.StatusOK; resp.StatusCode != want {
		t.Fatalf("Should have succeeded to fetch /; got %s, want %s", resp.Status, http.StatusText(want))
	}
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, resp.Body); err != nil {
		t.Fatalf("couldn't read content from server: %v", err)
	}
	if got, want := strings.Count(buf.String(), "<meta"), 1; got != want {
		t.Fatalf("did not find all the tags I need; got %d, want %d", got, want)
	}

	{
		u := fmt.Sprintf("%s/%s", ts.URL, prefix["pkgs"])
		resp, err := http.Get(u)
		if err != nil {
			t.Error(err)
		}
		buf := &bytes.Buffer{}
		io.Copy(buf, resp.Body)
		pkgs := []Package{}
		if err := json.NewDecoder(buf).Decode(&pkgs); err != nil {
			t.Fatalf("problem parsing json: %v, \n%q", err, buf)
		}
		if got, want := len(pkgs), 1; got != want {
			t.Fatalf("should (mildly) populated pkg list; got %d, want %d", got, want)
		}
	}
}

func TestInvalidPath(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)
	tok, err := db.addUser("sm@example.org")
	if err != nil {
		t.Error("failure to add user: %v", err)
	}

	bad := ts.URL
	body := strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`)
	req, err := http.NewRequest("POST", bad, body)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("couldn't POST: %v", err)
	}
	if len(db.Pkgs()) != 0 {
		t.Fatalf("should have failed to insert; got %d, want %d", len(db.Pkgs()), 0)
	}
	if got, want := resp.StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("should have failed to post at bad route; got %s, want %s", http.StatusText(got), http.StatusText(want))
	}
}

func TestCannotDuplicateExistingPath(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)

	tok, err := db.addUser("sm@example.org")
	if err != nil {
		t.Error("failure to add user: %v", err)
	}

	u := fmt.Sprintf("%s/foo", ts.URL)
	{
		body := strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`)
		req, err := http.NewRequest("POST", u, body)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("couldn't POST: %v", err)
		}
		if want := http.StatusOK; resp.StatusCode != want {
			t.Fatalf("initial post should have worked; got %s, want %s", resp.Status, http.StatusText(want))
		}
	}

	{
		body := strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`)
		req, err := http.NewRequest("POST", u, body)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("couldn't POST: %v", err)
		}
		if want := http.StatusConflict; resp.StatusCode != want {
			t.Fatalf("initial post should have worked; got %s, want %s", resp.Status, http.StatusText(want))
		}
	}
}

func TestCannotAddExistingSubPath(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)

	tok, err := db.addUser("sm@example.org")
	if err != nil {
		t.Error("failure to add user: %v", err)
	}

	{
		u := fmt.Sprintf("%s/foo/bar", ts.URL)
		t.Logf("url: %v", u)
		body := strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`)
		req, err := http.NewRequest("POST", u, body)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("couldn't POST: %v", err)
		}
		if want := http.StatusOK; resp.StatusCode != want {
			t.Fatalf("initial post should have worked; got %s, want %s", resp.Status, http.StatusText(want))
		}
	}

	{
		u := fmt.Sprintf("%s/foo", ts.URL)
		body := strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`)
		req, err := http.NewRequest("POST", u, body)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("couldn't POST: %v", err)
		}
		if want := http.StatusConflict; resp.StatusCode != want {
			t.Fatalf("initial post should have worked; got %s, want %s", resp.Status, http.StatusText(want))
		}
	}
}

func TestMissingRepo(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)

	tok, err := db.addUser("sm@example.org")
	if err != nil {
		t.Error("failure to add user: %v", err)
	}

	u := fmt.Sprintf("%s/foo", ts.URL)
	body := strings.NewReader(`{}`)
	req, err := http.NewRequest("POST", u, body)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("couldn't POST: %v", err)
	}
	if len(db.Pkgs()) != 0 {
		t.Fatalf("should have failed to insert; got %d, want %d", len(db.Pkgs()), 0)
	}
	if want := http.StatusBadRequest; resp.StatusCode != want {
		t.Fatalf("should have failed to post with bad payload; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestBadJson(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)

	tok, err := db.addUser("sm@example.org")
	if err != nil {
		t.Error("failure to add user: %v", err)
	}

	u := fmt.Sprintf("%s/foo", ts.URL)
	body := strings.NewReader(`{`)
	req, err := http.NewRequest("POST", u, body)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("couldn't POST: %v", err)
	}
	if len(db.Pkgs()) != 0 {
		t.Fatalf("should have failed to insert; got %d, want %d", len(db.Pkgs()), 0)
	}
	if want := http.StatusBadRequest; resp.StatusCode != want {
		t.Fatalf("should have failed to post at bad route; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestNoAuth(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)

	u := fmt.Sprintf("%s/foo", ts.URL)
	body := strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`)
	req, err := http.NewRequest("POST", u, body)
	req.Header.Add("Content-Type", "application/json")

	// here we don't set the Authorization header
	// req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("couldn't POST: %v", err)
	}
	resp.Body.Close()
	if got, want := resp.StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("posted with missing auth; got %v, want %v", http.StatusText(got), http.StatusText(want))
	}
}

func TestBadVcs(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)

	tok, err := db.addUser("sm@example.org")
	if err != nil {
		t.Error("failure to add user: %v", err)
	}

	u := fmt.Sprintf("%s/foo", ts.URL)
	body := strings.NewReader(`{"vcs": "bitbucket", "repo": "https://s.mcquay.me/sm/vain"}`)
	req, err := http.NewRequest("POST", u, body)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("couldn't POST: %v", err)
	}
	resp.Body.Close()
	if got, want := resp.StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("should have reported bad vcs specified; got %v, want %v", http.StatusText(got), http.StatusText(want))
	}
}

func TestUnsupportedMethod(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)

	tok, err := db.addUser("sm@example.org")
	if err != nil {
		t.Error("failure to add user: %v", err)
	}

	url := fmt.Sprintf("%s/foo", ts.URL)
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("couldn't POST: %v", err)
	}
	if len(db.Pkgs()) != 0 {
		t.Fatalf("should have failed to insert; got %d, want %d", len(db.Pkgs()), 0)
	}
	if want := http.StatusMethodNotAllowed; resp.StatusCode != want {
		t.Fatalf("should have failed to post at bad route; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestDelete(t *testing.T) {
	db, done := testDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	sm := http.NewServeMux()
	NewServer(sm, db, "", window)
	ts := httptest.NewServer(sm)

	tok, err := db.addUser("sm@example.org")
	if err != nil {
		t.Error("failure to add user: %v", err)
	}
	t.Logf("%v", tok)
	if len(db.Pkgs()) != 0 {
		t.Fatalf("started with something in it; got %d, want %d", len(db.Pkgs()), 0)
	}

	u := fmt.Sprintf("%s/foo", ts.URL)
	body := strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`)
	req, err := http.NewRequest("POST", u, body)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("couldn't POST: %v", err)
	}

	if got, want := len(db.Pkgs()), 1; got != want {
		t.Fatalf("pkgs should have something in it; got %d, want %d", got, want)
	}

	{
		// test not found
		u := fmt.Sprintf("%s/bar", ts.URL)
		client := &http.Client{}
		req, err := http.NewRequest("DELETE", u, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("couldn't POST: %v", err)
		}
		if got, want := resp.StatusCode, http.StatusNotFound; got != want {
			t.Fatalf("should have not been able to delete unknown package; got %v, want %v", http.StatusText(got), http.StatusText(want))
		}
	}

	{
		client := &http.Client{}
		req, err := http.NewRequest("DELETE", u, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("couldn't POST: %v", err)
		}

		if got, want := len(db.Pkgs()), 0; got != want {
			t.Fatalf("pkgs should be empty; got %d, want %d", got, want)
		}
	}
}
