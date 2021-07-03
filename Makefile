protoGen:
	cd api; \
	protoc ./api.proto --go-grpc_out=../.. --go_out=../.. --objc_out=.;

all: build

build: deps
	go build app/main/main.go;

deps:
	go get ./...;

install:
	mv main /bin/worker