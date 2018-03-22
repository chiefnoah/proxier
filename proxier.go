package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "proxier"
	app.Usage = "A commandline development path proxier"
	app.Action = func(c *cli.Context) error {
		port := c.String("port")
		paths := c.StringSlice("path")
		dests := c.StringSlice("dest")
		rewriteHosts := c.Bool("rewritehost")
		//Check if we have an equal number of path -> destination mappings, if we don't we die.
		if len(paths) != len(dests) {
			log.Fatal("Same number of path -> destination mappings is required")
		}

		//Generate ReverseProxies
		for i, p := range paths {
			log.Printf("%s -> %s", p, dests[i])
			base := makeTargetURL(p, dests[i])
			http.HandleFunc(p, newRedirectProxy(base, rewriteHosts).ServeHTTP)
		}
		log.Printf("Listening on port %s", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
		return nil
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "port, P", Value: "8080"},
		cli.StringSliceFlag{Name: "path, p"},
		cli.StringSliceFlag{Name: "dest, d"},
		cli.BoolFlag{Name: "rewritehost, r"},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

//Adapted from https://golang.org/src/net/http/httputil/reverseproxy.go

// NewSingleHostReverseProxy returns a new ReverseProxy that routes
// URLs to the scheme, host, and base path provided in target. If the
// target's path is "/base" and the incoming request was for "/dir",
// the target request will be for /base/dir.
// NewSingleHostReverseProxy does not rewrite the Host header.
// To rewrite Host headers, use ReverseProxy directly with a custom
// Director policy.
func newRedirectProxy(target *url.URL, rewriteHost bool) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		if strings.HasSuffix(target.String(), "/") {
			log.Printf("Redirecting %s to %s", req.URL.String(), target.String()+req.URL.Path)
		} else {
			log.Printf("Redirecting %s to %s", req.URL.String(), target.String())
		}
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)

		if rewriteHost {
			req.Host = target.Host
		}
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

//I did this wrong. See here for a reference: https://stackoverflow.com/questions/32542282/how-do-i-rewrite-urls-in-a-proxy-response-in-nginx
func makeTargetURL(p, t string) *url.URL {
	var base *url.URL
	if strings.HasSuffix(t, "/") {
		var err error
		base, err = url.Parse(t)
		if err != nil {
			log.Fatalf("Invalid target URL (%s) parameter: %s", t, err)
		}

	} else {
		pathURL, err := url.Parse(p)
		if err != nil {
			log.Fatalf("Invalid path URL (%s) parameter: %s", t, err)
		}
		log.Printf("TEST: %s", path.Join(t, pathURL.Path))
		base, err = url.Parse(path.Join(t, pathURL.Path))
		if err != nil {
			log.Fatalf("Unable to process paths: %s", err)
		}
		log.Printf("Target without provided basepath: %s", base.String())
	}
	return base
}
