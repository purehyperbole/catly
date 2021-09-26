FROM golang:1.17.0-bullseye as build-env

ARG gopath=/go

ENV GOPATH=$gopath

RUN apt update && apt install -y git make protobuf-compiler
ADD . /build
WORKDIR /build
RUN go install github.com/golang/protobuf/protoc-gen-go
RUN make generate
RUN CGO_ENABLED=0 go build -o catly-server cmd/server/main.go

FROM gcr.io/distroless/static-debian10

COPY --from=build-env /build/catly-server /
CMD ["/catly-server"]