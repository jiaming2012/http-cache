package main

import (
	"fmt"
	"github.com/jiaming2012/http-cache/src/cache/memory"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// refactor this
func init() {
	storage = memory.NewStorage()
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hashKey := r.Host + r.URL.String()
	content := storage.Get(hashKey)
	if content != nil {
		fmt.Print("Cache Hit!\n")
		w.Write(content)
	} else {
		log.Println(r.RemoteAddr, " ", r.Method, " ", r.URL)

		client := &http.Client{}

		//http: Request.RequestURI can't be set in client requests.
		//http://golang.org/src/pkg/net/http/client.go
		r.RequestURI = ""
		r.URL.Scheme = "http"
		r.URL.Host = r.Host

		delHopHeaders(r.Header)

		if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			appendHostToXForwardHeader(r.Header, clientIP)
		}

		resp, err := client.Do(r)
		if err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			log.Fatal("ServeHTTP:", err)
		}
		defer resp.Body.Close()

		log.Println(r.RemoteAddr, " ", resp.Status)

		delHopHeaders(resp.Header)

		// cache
		// todo: make dynamic
		duration := "20s"
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// todo: propagate
			//log.Printf("Error reading body: %v", err)
			//http.Error(w, "can't read body", http.StatusBadRequest)
			//return
			panic(err)
		}

		if d, err := time.ParseDuration(duration); err == nil {
			fmt.Printf("New page cached: %s for %s\n", r.RequestURI, duration)
			storage.Set(hashKey, body, d)
		}

		copyHeader(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		fmt.Fprintf(w, "%s", body)
	}
}
