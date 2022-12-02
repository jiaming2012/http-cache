package main

import (
	"fmt"
	"github.com/jiaming2012/http-cache/src/constants"
	log "github.com/jiaming2012/http-cache/src/logger"
	"net/http"
)

func main() {
	port := constants.Port

	if port == "" {
		port = "8080"
	}

	handler := &proxy{}

	log.Logger.Infof("Operating in mode %s", constants.AppMode)
	log.Logger.Infof("Proxy listening on :%s\n", port)

	http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
}
