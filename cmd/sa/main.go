package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	serviceAccountEmulator "github.com/navikt/nada-backend/pkg/sa/emulator"
	"github.com/rs/zerolog"
)

const (
	readHeaderTimeout = 10 * time.Second
)

// nolint: gochecknoglobals
var port = flag.String("port", "8080", "Port to run the HTTP server on")

func main() {
	flag.Parse()

	log := zerolog.New(os.Stdout)

	saEmulator := serviceAccountEmulator.New(log)
	router := saEmulator.GetRouter()

	log.Printf("Server starting on port %s...", *port)

	server := &http.Server{
		Addr:              ":" + *port,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("starting server")
	}
}
