---
page_title: "gRPC Full Provider"
description: |-
  This provider allowing simple gRPC unary requests to external server.
---

# gRPC Full Provider

The gRPC-FULL provider is a utility provider for interacting with generic gRPC
servers as part of a Terraform configuration. 

The gRPC service can return arbitrary protobuf messages that are deserialized to JSON for use in terraform.


## Example Usage

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
    last_name  = "amander"
  })

}

output "data" {
  value = jsondecode(data.grpc.example.payload).message
}
```