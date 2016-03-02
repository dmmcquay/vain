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
)

func TestAdd(t *testing.T) {
	ms := NewSimpleStore("")
	sm := http.NewServeMux()
	_ = NewServer(sm, ms)
	ts := httptest.NewServer(sm)
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("couldn't GET: %v", err)
	}
	resp.Body.Close()
	if len(ms.p) != 0 {
		t.Errorf("started with something in it; got %d, want %d", len(ms.p), 0)
	}

	bad := ts.URL
	resp, err = http.Post(bad, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	resp.Body.Close()
	if len(ms.p) != 0 {
		t.Errorf("started with something in it; got %d, want %d", len(ms.p), 0)
	}

	{
		u := fmt.Sprintf("%s/db/", ts.URL)
		resp, err := http.Get(u)
		if err != nil {
			t.Error(err)
		}
		buf := &bytes.Buffer{}
		io.Copy(buf, resp.Body)
		pkgs := []Package{}
		if err := json.NewDecoder(buf).Decode(&pkgs); err != nil {
			t.Errorf("problem parsing json: %v, \n%q", err, buf)
		}
		if got, want := len(pkgs), 0; got != want {
			t.Errorf("should have empty pkg list; got %d, want %d", got, want)
		}
	}

	u := fmt.Sprintf("%s/foo", ts.URL)
	resp, err = http.Post(u, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}

	if len(ms.p) != 1 {
		t.Errorf("storage should have something in it; got %d, want %d", len(ms.p), 1)
	}

	ur, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}
	good := fmt.Sprintf("%s/foo", ur.Host)
	p, ok := ms.p[good]
	if !ok {
		t.Fatalf("did not find package for %s; should have posted a valid package", good)
	}
	if p.path != good {
		t.Errorf("package name did not go through as expected; got %q, want %q", p.path, good)
	}
	if want := "https://s.mcquay.me/sm/vain"; p.Repo != want {
		t.Errorf("repo did not go through as expected; got %q, want %q", p.Repo, want)
	}
	if got, want := p.Vcs, "git"; got != want {
		t.Errorf("Vcs did not go through as expected; got %q, want %q", got, want)
	}

	resp, err = http.Get(ts.URL)
	if err != nil {
		t.Errorf("couldn't GET: %v", err)
	}
	defer resp.Body.Close()
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("Should have succeeded to fetch /; got %s, want %s", resp.Status, http.StatusText(want))
	}
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, resp.Body); err != nil {
		t.Errorf("couldn't read content from server: %v", err)
	}
	if got, want := strings.Count(buf.String(), "meta"), 1; got != want {
		t.Errorf("did not find all the tags I need; got %d, want %d", got, want)
	}

	{
		u := fmt.Sprintf("%s/db/", ts.URL)
		resp, err := http.Get(u)
		if err != nil {
			t.Error(err)
		}
		buf := &bytes.Buffer{}
		io.Copy(buf, resp.Body)
		pkgs := []Package{}
		if err := json.NewDecoder(buf).Decode(&pkgs); err != nil {
			t.Errorf("problem parsing json: %v, \n%q", err, buf)
		}
		if got, want := len(pkgs), 1; got != want {
			t.Errorf("should (mildly) populated pkg list; got %d, want %d", got, want)
		}
	}
}

func TestInvalidPath(t *testing.T) {
	ms := NewSimpleStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)

	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if len(ms.p) != 0 {
		t.Errorf("should have failed to insert; got %d, want %d", len(ms.p), 0)
	}
	if want := http.StatusBadRequest; resp.StatusCode != want {
		t.Errorf("should have failed to post at bad route; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestCannotDuplicateExistingPath(t *testing.T) {
	ms := NewSimpleStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)

	url := fmt.Sprintf("%s/foo", ts.URL)
	resp, err := http.Post(url, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("initial post should have worked; got %s, want %s", resp.Status, http.StatusText(want))
	}
	resp, err = http.Post(url, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if want := http.StatusConflict; resp.StatusCode != want {
		t.Errorf("initial post should have worked; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestCannotAddExistingSubPath(t *testing.T) {
	ms := NewSimpleStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)

	url := fmt.Sprintf("%s/foo/bar", ts.URL)
	resp, err := http.Post(url, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("initial post should have worked; got %s, want %s", resp.Status, http.StatusText(want))
	}

	url = fmt.Sprintf("%s/foo", ts.URL)
	resp, err = http.Post(url, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	resp, err = http.Post(url, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if want := http.StatusConflict; resp.StatusCode != want {
		t.Errorf("initial post should have worked; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestMissingRepo(t *testing.T) {
	ms := NewSimpleStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)
	url := fmt.Sprintf("%s/foo", ts.URL)
	resp, err := http.Post(url, "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if len(ms.p) != 0 {
		t.Errorf("should have failed to insert; got %d, want %d", len(ms.p), 0)
	}
	if want := http.StatusBadRequest; resp.StatusCode != want {
		t.Errorf("should have failed to post at bad route; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestBadJson(t *testing.T) {
	ms := NewSimpleStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)
	url := fmt.Sprintf("%s/foo", ts.URL)
	resp, err := http.Post(url, "application/json", strings.NewReader(`{`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if len(ms.p) != 0 {
		t.Errorf("should have failed to insert; got %d, want %d", len(ms.p), 0)
	}
	if want := http.StatusBadRequest; resp.StatusCode != want {
		t.Errorf("should have failed to post at bad route; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestUnsupportedMethod(t *testing.T) {
	ms := NewSimpleStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)
	url := fmt.Sprintf("%s/foo", ts.URL)
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if len(ms.p) != 0 {
		t.Errorf("should have failed to insert; got %d, want %d", len(ms.p), 0)
	}
	if want := http.StatusMethodNotAllowed; resp.StatusCode != want {
		t.Errorf("should have failed to post at bad route; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestNewServer(t *testing.T) {
	ms := NewSimpleStore("")
	sm := http.NewServeMux()
	s := NewServer(sm, ms)
	ts := httptest.NewServer(s)
	url := fmt.Sprintf("%s/foo", ts.URL)
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if len(ms.p) != 0 {
		t.Errorf("should have failed to insert; got %d, want %d", len(ms.p), 0)
	}
	if want := http.StatusMethodNotAllowed; resp.StatusCode != want {
		t.Errorf("should have failed to post at bad route; got %s, want %s", resp.Status, http.StatusText(want))
	}
}

func TestDelete(t *testing.T) {
	ms := NewSimpleStore("")
	sm := http.NewServeMux()
	_ = NewServer(sm, ms)
	ts := httptest.NewServer(sm)
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("couldn't GET: %v", err)
	}
	resp.Body.Close()
	if len(ms.p) != 0 {
		t.Errorf("started with something in it; got %d, want %d", len(ms.p), 0)
	}

	u := fmt.Sprintf("%s/foo", ts.URL)
	resp, err = http.Post(u, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}

	if got, want := len(ms.p), 1; got != want {
		t.Errorf("storage should have something in it; got %d, want %d", got, want)
	}

	{
		// test not found
		u := fmt.Sprintf("%s/bar", ts.URL)
		client := &http.Client{}
		req, err := http.NewRequest("DELETE", u, nil)
		resp, err = client.Do(req)
		if err != nil {
			t.Errorf("couldn't POST: %v", err)
		}
		if got, want := resp.StatusCode, http.StatusNotFound; got != want {
			t.Errorf("should have not been able to delete unknown package; got %v, want %v", http.StatusText(got), http.StatusText(want))
		}
	}
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", u, nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}

	if got, want := len(ms.p), 0; got != want {
		t.Errorf("storage should be empty; got %d, want %d", got, want)
	}
}
