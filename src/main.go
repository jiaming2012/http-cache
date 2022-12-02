//package main
//
//import (
//	"github.com/elazarl/goproxy"
//	"log"
//	"net/http"
//)
//
//func main() {
//	proxy := goproxy.NewProxyHttpServer()
//	proxy.Verbose = true
//	proxy.OnResponse()
//	log.Fatal(http.ListenAndServe(":8080", proxy))
//}

package main

import (
	"fmt"
	"github.com/jiaming2012/http-cache/src/cache"
	"net/http"
	"os"
)

var storage cache.Storage

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := &proxy{}

	fmt.Printf("Proxy listening on :%s\n", port)
	http.ListenAndServe(":"+port, handler)
}
