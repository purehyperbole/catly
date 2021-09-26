package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/purehyperbole/catly/protocol/catly"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server", "127.0.0.1:8000", "Specifies the address of the gRPC server. Defaults to 127.0.0.1:8000")
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("command must specify a valid file path")
		os.Exit(1)
	}

	// open the specified file
	fd, err := os.Open(os.Args[1])
	check(err, "failed to open specified file")
	defer fd.Close()

	// read it's data into memory
	data, err := io.ReadAll(fd)
	check(err, "failed to read specified file")

	// open the gRPC client and prepare to send the request
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	check(err, "failed to connect to server")

	client := catly.NewObjectClient(conn)

	// send the data
	req := &catly.UploadObjectRequest{
		Name: filepath.Base(os.Args[1]),
		Data: data,
	}

	resp, err := client.Upload(context.Background(), req)
	check(err, "failed to upload file")

	if resp.Status != catly.ObjectStatus_ObjectOK {
		check(errors.New(resp.Error), "file upload failed with")
	}

	fmt.Printf("your image is now available at: %s\n", resp.Url)
}

func check(err error, pfx string) {
	if err != nil {
		fmt.Printf("%s: %s\n", pfx, err.Error())
		os.Exit(1)
	}
}
