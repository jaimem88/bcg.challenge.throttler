package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jaimemartinez88/bcg.challenge.throttler"
	log "github.com/sirupsen/logrus"
)

const (
	httpServerReadTimeout  = 3 * time.Second
	httpServerWriteTimeout = 120 * time.Second
)

var (
	version            string
	corsAllowedHeaders = handlers.AllowedHeaders([]string{"*"})
	corsAllowedDomains = handlers.AllowedOrigins([]string{
		"*",
	})
	corsAllowedMethods = handlers.AllowedMethods([]string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions})
)

func main() {
	// flag handling
	defaultLocation := flag.String("default", "", "location to write a default configuration to (this will overwrite an existing file at this location)")
	configLocation := flag.String("config", "", "JSON config file to load")

	flag.Parse()
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})
	log.SetLevel(log.DebugLevel)
	if *defaultLocation == "" && *configLocation == "" {
		log.Println("Using default config:")
		data, _ := json.MarshalIndent(config, "", "  ")
		io.Copy(os.Stdout, bytes.NewReader(data))
		fmt.Printf("\n")
	} else if *defaultLocation != "" {
		writeDefaultConfig(*defaultLocation)
		os.Exit(0)
	} else if *configLocation != "" {
		loadConfig(*configLocation)
	}

	service := throttler.NewService(
		config.Environment,
		config.Throttling.N,
		config.Throttling.M,
	)

	r := mux.NewRouter()
	r.Use(service.LoggingMiddleware)
	r.NotFoundHandler = http.HandlerFunc(service.HandleNotFound)

	r.HandleFunc("/healthcheck", service.HandleHealthcheck).Methods(http.MethodGet)

	v1 := r.PathPrefix("/v1/").Subrouter()
	v1.Use(service.ValidateAccessToken)
	v1.Use(service.CheckLimitsMiddlware)

	v1.HandleFunc("/users", service.HandleGetUsers).Methods(http.MethodGet)

	handler := handlers.CORS(corsAllowedHeaders, corsAllowedDomains, corsAllowedMethods)(r)
	port := config.HTTP.ListenPort
	if port == "" { // default to port 8080
		port = "8080"
	}
	httpServer := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  httpServerReadTimeout,
		WriteTimeout: httpServerWriteTimeout,
		Handler:      handler,
	}
	log.Infof("Server listening on port :%s", port)
	if err := httpServer.ListenAndServe(); nil != err {
		log.Fatalln("Failed to start server", err)
	}

}
