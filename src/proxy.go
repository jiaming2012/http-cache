package main

import (
	"fmt"
	"github.com/jiaming2012/http-cache/src/cache"
	"github.com/jiaming2012/http-cache/src/cache/eventstore"
	"github.com/jiaming2012/http-cache/src/cache/memory"
	"github.com/jiaming2012/http-cache/src/constants"
	log "github.com/jiaming2012/http-cache/src/logger"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

var storage cache.Storage

// refactor this
func init() {
	if constants.AppMode == "memory" {
		storage = memory.NewStorage()
	} else if constants.AppMode == "eventstoredb" {
		var err error
		storage, err = eventstore.NewStorage()
		if err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("unknown value for env APP_MODE=%v. Expected memory or eventstoredb", constants.AppMode))
	}
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

func deleteHopHeaders(header http.Header) {
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

func buildCacheKey(r *http.Request) string {
	return r.Host + r.URL.String()
}

func cacheResponse(cacheKey string, body []byte) error {
	// todo: handle in viper
	duration, err := time.ParseDuration(constants.CacheDuration)
	if err != nil {
		return err
	}

	storage.Set(cacheKey, body, duration)
	log.Logger.Infof("Successfully cached: %s for %s\n", cacheKey, constants.CacheDuration)
	return nil
}

type proxy struct {
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cacheKey := buildCacheKey(r)
	content := storage.Get(cacheKey)
	if content != nil {
		log.Logger.Debug("Cache Hit!")
		w.Write(content)
	} else {
		log.Logger.Debug("Cache Miss!")
		log.Logger.Infof("Log Request: %s %s %s", r.RemoteAddr, r.Method, r.URL)

		client := &http.Client{}

		//http: Request.RequestURI can't be set in client requests.
		//http://golang.org/src/pkg/net/http/client.go
		r.RequestURI = ""
		r.URL.Scheme = "http"
		r.URL.Host = r.Host

		deleteHopHeaders(r.Header)

		if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			appendHostToXForwardHeader(r.Header, clientIP)
		}

		resp, err := client.Do(r)
		if err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			log.Logger.Fatalf("ServeHTTP: %v", err)
		}
		defer resp.Body.Close()

		// todo: read in chunks as opposed to whole body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Logger.Errorf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		log.Logger.Infof("Log Response: %s %s", r.RemoteAddr, resp.Status)

		deleteHopHeaders(resp.Header)

		// cache our response before sending back to user
		// todo: make this an event
		cacheResponse(cacheKey, body)

		copyHeader(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		fmt.Fprintf(w, "%s", body)
	}
}
