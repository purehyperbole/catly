test:
	go test -v -cover -covermode=atomic -race -coverprofile=coverage.out $(shell go list ./...)

generate:
	protoc --proto_path=protocol --go_out=plugins=grpc:protocol --go_opt=paths=source_relative protocol/catly/object.proto