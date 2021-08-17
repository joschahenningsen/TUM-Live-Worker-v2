all: build

protoGen:
	cd api; \
	protoc ./api.proto --go-grpc_out=../.. --go_out=../..

build: deps
	go build app/main/main.go;

deps:
	go get ./...;

install:
	mv main /bin/worker