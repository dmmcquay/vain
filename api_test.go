package vain

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdd(t *testing.T) {
	ms := NewMemStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)
	s.hostname = ts.URL
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("couldn't GET: %v", err)
	}
	resp.Body.Close()
	if len(s.storage.p) != 0 {
		t.Errorf("started with something in it; got %d, want %d", len(s.storage.p), 0)
	}

	bad := ts.URL
	resp, err = http.Post(bad, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	resp.Body.Close()
	if len(s.storage.p) != 0 {
		t.Errorf("started with something in it; got %d, want %d", len(s.storage.p), 0)
	}

	good := fmt.Sprintf("%s/foo", ts.URL)
	resp, err = http.Post(good, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}

	if len(s.storage.p) != 1 {
		t.Errorf("storage should have something in it; got %d, want %d", len(s.storage.p), 1)
	}

	p, ok := s.storage.p[good]
	if !ok {
		t.Fatalf("did not find package for %s; should have posted a valid package", good)
	}
	if p.Path != good {
		t.Errorf("package name did not go through as expected; got %q, want %q", p.Path, good)
	}
	if want := "https://s.mcquay.me/sm/vain"; p.Repo != want {
		t.Errorf("repo did not go through as expected; got %q, want %q", p.Repo, want)
	}
	if want := Git; p.Vcs != want {
		t.Errorf("Vcs did not go through as expected; got %q, want %q", p.Vcs, want)
	}
}

func TestInvalidPath(t *testing.T) {
	ms := NewMemStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)
	s.hostname = ts.URL

	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(`{"repo": "https://s.mcquay.me/sm/vain"}`))
	if err != nil {
		t.Errorf("couldn't POST: %v", err)
	}
	if len(s.storage.p) != 0 {
		t.Errorf("should have failed to insert; got %d, want %d", len(s.storage.p), 0)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("should have failed to post at bad route; got %s, want %s", resp.Status, http.StatusText(http.StatusBadRequest))
	}
}

func TestCannotDuplicateExistingPath(t *testing.T) {
	ms := NewMemStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)
	s.hostname = ts.URL

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
	ms := NewMemStore("")
	s := &Server{
		storage: ms,
	}
	ts := httptest.NewServer(s)
	s.hostname = ts.URL

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
