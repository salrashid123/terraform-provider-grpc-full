module github.com/salrashid123/terraform-provider-grpc-full

require (
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/hashicorp/go-hclog v1.2.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.0
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/psanford/lencode v0.3.0
	github.com/salrashid123/grpc_wireformat/grpc_services/src/echo v0.0.0
	golang.org/x/net v0.0.0-20210326060303-6b1517762897
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
)

go 1.13

replace (
  github.com/salrashid123/grpc_wireformat/grpc_services/src/echo => "./example/src/echo"
)