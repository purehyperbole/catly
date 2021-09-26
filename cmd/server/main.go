package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/purehyperbole/catly/api"
	"github.com/purehyperbole/catly/protocol/catly"
	"github.com/purehyperbole/catly/storage"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	// DefaultMaxRequestSize default maximum accepted file size of 8MB
	DefaultMaxRequestSize = 1 << 23
	// DefaultDomain default domain name that the service is running under
	DefaultDomain = "http://127.0.0.1"
	// DefaultHTTPPort default port that the http service will run on
	DefaultHTTPPort = "8080"
	// DefaultGRPCPort default port that the grpc service will run on
	DefaultGRPCPort = "8000"
	// DefaultStoragePath default storage path that will be used. By default,
	// this is will use in-memory storage unless a path is specified
	DefaultStoragePath = ":memory:"
)

// storageProvider defines the interface that storage providers need to implement
type storageProvider interface {
	ReadObject(id string, w io.Writer) error
	WriteObject(id string, r io.Reader) error
}

func main() {
	// get the configuration from the environment
	domain := getEnv("CATLY_DOMAIN", DefaultDomain)
	httpPort := getEnv("CATLY_HTTP_PORT", DefaultHTTPPort)
	grpcPort := getEnv("CATLY_GRPC_PORT", DefaultGRPCPort)
	storagePath := getEnv("CATLY_STORAGE_PATH", DefaultStoragePath)
	maxRequestSize := getEnvInt("CATLY_MAX_REQUEST_SIZE", DefaultMaxRequestSize)

	// setup storage providers based on the different storage options
	log.Info().Msg(fmt.Sprintf("setting up storage in %s", storagePath))

	var sp storageProvider
	var err error

	switch storagePath {
	case DefaultStoragePath:
		sp = storage.NewMemoryStore()
	case "file":
		err = os.MkdirAll(storagePath, 0744)
		check(err, "failed to create storage directory")

		sp, err = storage.NewFileStore(storagePath)
		check(err, "failed to setup file storage")
	}

	// setup the grpc server
	log.Info().Msg(fmt.Sprintf("starting gRPC listener on *:%s", grpcPort))

	address := fmt.Sprintf("%s:%s/", domain, httpPort)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	check(err, "failed to start gRPC listener")

	s := grpc.NewServer(
		grpc.MaxSendMsgSize(maxRequestSize),
		grpc.MaxRecvMsgSize(maxRequestSize),
	)

	catly.RegisterObjectServer(
		s,
		api.NewGRPCResource(address, sp),
	)

	go func() {
		err := s.Serve(listener)
		check(err, "failed to serve gRPC")
	}()

	// start the http server
	log.Info().Msg(fmt.Sprintf("starting HTTP listener on *:%s", httpPort))

	hr := api.NewHTTPResource(sp)

	mux := http.NewServeMux()
	mux.HandleFunc("/", hr.GetObject)

	err = http.ListenAndServe(
		fmt.Sprintf(":%s", httpPort),
		mux,
	)

	check(err, "failed to start HTTP listener")
}

func check(err error, pfx string) {
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("%s: %s", pfx, err.Error()))
	}
}

func getEnv(name, defaultValue string) string {
	e := os.Getenv(name)
	if e == "" {
		return defaultValue
	}

	return e
}

func getEnvInt(name string, defaultValue int) int {
	e := os.Getenv(name)
	if e == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(e)
	check(err, fmt.Sprintf("failed to read integer environment variable '%s'", name))

	return i
}
