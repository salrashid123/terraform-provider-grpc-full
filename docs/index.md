---
page_title: "gRPC Full Provider"
description: |-
  This provider allowing simple gRPC unary requests to external server.
---

# gRPC Full Provider

The gRPC-FULL provider is a utility for interacting with **generic** gRPC
servers as part of a Terraform configuration. 

Use this provider to construct a gRPC request protobuf message in JSON, send it to a remote system and deserialize the binary response into JSON for use in the terraform module as a datasoruce


## Example Usage

You can find a full end-to-end example [here](https://github.com/salrashid123/terraform-provider-grpc-full/tree/main/example).


given the sample  `.proto`:

```proto
syntax = "proto3";
package echo;

service EchoServer {
  rpc SayHello (EchoRequest) returns (EchoReply) {}
}

message Middle {
  string name = 1;
}

message EchoRequest {
  string first_name = 1;
  string last_name = 2;
  Middle middle_name = 3;
}

message EchoReply {
  string message = 1;
}
```

Construct the JSON format of the messages as the `request_body`, register the compiled `.pb`  and invoke the url.

Finally, use the grpc response data as a datasource.

```terraform
terraform {
  required_providers {
    grpc-full = {
      source = "salrashid123/grpc-full"
    }
  }
}

provider "grpc-full" {}

data "grpc" "example" {
  provider = grpc-full

  url                = "https://localhost:50051/echo.EchoServer/SayHello"
  ca                 = file("${path.module}/certs/root-ca.crt")
  sni                = "localhost"
  request_timeout_ms = 1000

  registry_files = [
    filebase64("${path.module}/src/echo/echo.pb"),
  ]

  request_headers = {
    authorization = "bearer foo"
  }

  request_type  = "echo.EchoRequest"
  response_type = "echo.EchoReply"
  request_body = jsonencode({
    "@type"    = "echo.EchoRequest",
    first_name = "sal",
    last_name  = "mander"
    middle_name = {
      name = "a"
    }
  })

}

output "data" {
  value = jsondecode(data.grpc.example.payload).message
}
```

---

Also see: - [using protorefelect, dynamicpb and wire-encoding to send messages](https://blog.salrashid.dev/articles/2022/grpc_wireformat/)