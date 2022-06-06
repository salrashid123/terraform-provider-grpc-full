---
page_title: "gRPC-FULL Data Source"
description: |-
  Retrieves the content from an external gRPC service
---

# `grpc` Data Source

The `grpc` data source  facilitates acquiring data from an arbitrary external gRPC service.

## Example Usage

### Unary

```hcl
terraform {
  required_providers {
    grpc-full = {
       source = "salrashid123/grpc-full"
    }
  }
}

provider "grpc-full" {
}

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



## Argument Reference

The following arguments are supported:

* `url` - (Required) The URL to request data from. This URL must respond with
  _must_ include the service and method: 
  (eg `"https://localhost:50051/echo.EchoServer/SayHello"`)


* `registry_files`: this is a list of the compiled descriptors to load.  
  (`protoc --descriptor_set_out=echo.pb  echo.proto`).
  You must set the `@type` key

* `request_type`: the message type sent to the server (eg `"echo.EchoRequest"`)

* `response_type`: the message sent by the server (eg `"echo.EchoReply"`)

* `request_body`: this is json encoded format for the `request_type` being sent.  
   It *must* include an attribute of `@type` that signifies the fully qualified name of the message

* `insecure_skip_verify` - (Optional) Skip server TLS verification (default=`false`).

* `request_timeout_ms` - (Optional) Timeout the request in ms

* `ca`: this is the certificate authority that signed the server cert for TLS connections

* `sni`: the SNI for the server 

## Attributes Reference

The following attributes are exported:

* `status_code` - The status_code of the response if not error

* `payload` - The json format of the gRPC Response.

* `response_headers` - A map of strings representing the response HTTP headers.
  Duplicate headers are concatenated with `, ` according to
  [RFC2616](https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2)



