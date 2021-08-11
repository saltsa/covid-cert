package main

import (
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/handlers"
)

const (
	dir = "./static"
)

func main() {
	port := `:8080`
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	log.Printf("listening on %s", port)

	http.Handle("/", http.FileServer(http.Dir(dir)))
	logHandler := handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)

	http.ListenAndServeTLS(port, "cert.pem", "key.pem", logHandler)
}
