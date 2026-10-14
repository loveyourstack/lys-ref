# refrpcsrv

gRPC server.

## Prerequisites

### Install 'protoc' command for generating protocol buffer Go code from .proto files
```shell
sudo apt install -y protobuf-compiler
```

### Add Go package needed for running gRPC server
```shell
go get google.golang.org/grpc
```