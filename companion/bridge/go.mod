module github.com/the40n8/Sims4Access-mac/companion/bridge

go 1.26.0

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/protobuf/cmd/protoc-gen-go
)

require (
	connectrpc.com/connect v1.21.1
	google.golang.org/protobuf v1.36.12
)

require rsc.io/qr v0.2.0
