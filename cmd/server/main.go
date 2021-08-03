package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func main() {
	port := `:8080`
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	log.Printf("listening on %s", port)

	http.Handle("/", http.FileServer(http.Dir(`.`)))
	logHandler := handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)

	http.ListenAndServeTLS(port, "cert.pem", "key.pem", logHandler)
}
