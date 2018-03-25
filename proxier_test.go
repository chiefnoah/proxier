package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func testHandler(w http.ResponseWriter, r *http.Request) {

}

func TestTrailingTargetSlashRemovesPath(t *testing.T) {
	requestPath := "/test/test"

	ws1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
		if r.URL.Path == requestPath {
			t.Error("Request URL contains proxy route path.")
		}
	}))
	defer ws1.Close()
	requestURL, _ := url.Parse(ws1.URL)
	requestURL.Path = requestURL.Path + "/"
	proxy := httptest.NewServer(newRedirectProxy(requestURL, false))
	defer proxy.Close()
	_, err := http.Get(proxy.URL + requestPath)
	if err != nil {
		t.Errorf("Web request error: %s", err)
	}

}

func TestTrailingNoTargetKeepsFullPath(t *testing.T) {
	requestPath := "/test/test"

	ws1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
		if r.URL.Path != requestPath {
			t.Error("Request URL contains proxy route path.")
		}
	}))
	defer ws1.Close()
	requestURL, _ := url.Parse(ws1.URL)
	requestURL.Path = requestURL.Path
	proxy := httptest.NewServer(newRedirectProxy(requestURL, false))
	defer proxy.Close()
	_, err := http.Get(proxy.URL + requestPath)
	if err != nil {
		t.Errorf("Web request error: %s", err)
	}

}

func TestHostHeaderRewritten(t *testing.T) {
	requestPath := "/test/test"
	targetURL := &url.URL{Host: "test", Path: requestPath}
	proxy := newRedirectProxy(targetURL, true)

	req, _ := http.NewRequest("GET", requestPath, nil)
	req.Host = "changed"
	proxy.Director(req)
	if req.Host != "test" {
		t.Error("Host header was not changed")
	}
}

func TestHostHeaderNotRewritten(t *testing.T) {
	requestPath := "/test/test"
	targetURL := &url.URL{Host: "test", Path: requestPath}
	proxy := newRedirectProxy(targetURL, false)

	req, _ := http.NewRequest("GET", requestPath, nil)
	req.Host = "changed"
	proxy.Director(req)
	if req.Host != "changed" {
		t.Error("Host header was not changed")
	}
}
