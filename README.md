# Catly [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Go Reference](https://pkg.go.dev/badge/github.com/joinself/catly.svg)](https://pkg.go.dev/github.com/joinself/catly) [![Go Report Card](https://goreportcard.com/badge/github.com/purehyperbole/catly)](https://goreportcard.com/report/github.com/purehyperbole/catly) [![Build Status](https://github.com/purehyperbole/catly/workflows/main/badge.svg)](https://github.com/purehyperbole/catly/actions)

A file uploading service for your finest cat pictures 🐱

It currently supports the `jpeg`, `png` and `gif` image formats and has configurable storage backends.

## Requirements

To run the client and server, you will first need to ensure that you have installed:

1. [Golang 1.17](https://golang.org/doc/install)
2. [Docker](https://docs.docker.com/get-docker/)
3. [Protobuf Compiler](https://grpc.io/docs/protoc-installation/)

## Usage

To get started you will need to build and run the client and server.

### Server

To run the docker version of this server, run these commands in the root of this repo:

```sh
λ docker build -t catly-server .
λ docker run -it -p "8000:8000" -p "8080:8080" catly-server
```

### Client

To compile the client, you will need to run:
```sh
λ go build -o grpc-upload cmd/client/main.go
```

You can then test an upload with:

```sh
λ ./grpc-upload ./cat.jpg
```

## Configuration

There a number of different options that can be supplied when running the client and the server

### Server

When starting the container via docker, any of the following environment variables can be passed in:

| Name                   | Description                                                                                                            | Default            |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------ |
| CATLY_DOMAIN           | The domain that you are running the service under                                                                      | `http://127.0.0.1` |
| CATLY_HTTP_PORT        | The port the HTTP service will run on                                                                                  | `8080`             |
| CATLY_GRPC_PORT        | The port the gRPC upload service will run on                                                                           | `8000`             |
| CATLY_STORAGE_PATH     | The storage path in the container you wish to use. By default, only in memory storage will be used                     | `:memory:`         |
| CATLY_MAX_REQUEST_SIZE | The maximum request size in bytes the server will accept. This can be used to restrict large files from being uploaded | `8388608` (~ 8MB)  |

### Client

The client supports the following flags:

| Name    | Description                                | Default          |
| ------- | ------------------------------------------ | ---------------- |
| -server | Specifies the address for the catly server | `127.0.0.1:8000` |

## Structure

The project is broken into the following components

| Package    | Description                                                                                                                       |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------- |
| api        | Contains an implementation of an HTTP server for serving files and a gRPC server for handling uploads                             |
| cmd/server | Contains the main setup logic for the gRPC/HTTP server                                                                            |
| cmd/client | Contains the main setup logic for the gRPC upload client                                                                          |
| protocol   | Contains the protobuf bindings and definitions for the object service                                                             |
| storage    | Contains different storage implementations for catly server. Currently there is an in memory store, as well as a filesystem store |

## Roadmap
- [ ] Integration testing for client and server
- [ ] CI stage for linting
- [ ] Improved error messages in responses
- [ ] Support for HTTP and secure gRPC
- [ ] LRU cache for improving performance when serving popular images from file storage
- [ ] Generate prebuilt server and client for github releases
