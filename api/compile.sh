alias protoc="/home/joscha/Downloads/bin/protoc"
# generate messages:
protoc ./api.proto --go_out=../..
# generate services:
protoc ./api.proto --go-grpc_out=../..

