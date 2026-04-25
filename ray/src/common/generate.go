package common

//go:generate protoc --go_out=. --go-grpc_out=. --go-grpc_opt=require_unimplemented_servers=false registry.proto task.proto
