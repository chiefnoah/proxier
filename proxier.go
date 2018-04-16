package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/koding/websocketproxy"
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
		registerProxyPaths(paths, dests, rewriteHosts)
		startProxyServer(port)
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

func startProxyServer(port string) {
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func registerProxyPaths(paths, dests []string, rewriteHosts bool) {
	//Generate ReverseProxies
	for i, p := range paths {
		log.Printf("%s -> %s", p, dests[i])
		base := makeTargetURL(p, dests[i])
		if base.Scheme == "http" {
			http.HandleFunc(p, newRedirectProxy(base, rewriteHosts).ServeHTTP)
		} else if base.Scheme == "ws" {
			log.Println("Found a websocket proxy.")
			http.HandleFunc(p, websocketproxy.NewProxy(base).ServeHTTP)
		}
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
		originalURL := req.URL.String()
		//If the target URL ends with a "/", we want to remove the basepath from the route, the same way nginx does it.
		if strings.HasSuffix(target.String(), "/") {
			req.URL.Path = strings.Join(strings.Split(strings.Trim(req.URL.Path, "/"), "/")[1:], "/") //This is really ugly
			log.Printf("Removing path from target: %s", req.URL.Path)
		} else {
			req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		}
		log.Printf("Redirecting %s to %s", originalURL, req.URL.String())
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = path.Join(target.Path, req.URL.Path) //singleJoiningSlash(target.Path, req.URL.Path)

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
	base, err := url.Parse(t)
	if err != nil {
		log.Fatalf("Unable to process target URL: %s", err)
	}
	return base
}
