## How to compile these proto files into golang code

1. Install protoc version: [v3.17.3](https://github.com/protocolbuffers/protobuf/releases/tag/v3.17.3)

my protoc version:

```shell
$ protoc --version
libprotoc 3.17.3
```

2. Install protoc-gen-go and protoc-gen-go-grpc

3. Generate gRPC proto clients

```shell
cd ${your PROJECT path}/spec/proto/runtime/v1
protoc -I. --go_out=. --go-grpc_out=. *.proto
```
