
Terraform Provider for gRPC Dataources
=======================================

This datasource provider issues simple unary [gRPC](https://grpc.io/) API calls to an external system.

The API calls' response protobuf is converted to JSON and surfaced back to terraform.

As is the case with other terraform datasources, this provider is not meant to perform CRUD operations or mutations on the remote system.

Instead, use this provider to just get some simple data back.

>> Note, this provider is experimental and alpha quality, caveat emptor


- Website: https://www.terraform.io
- [using protorefelect, dynamicpb and wire-encoding to send messages](https://github.com/salrashid123/grpc_wireformat)

- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)

<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/5/5b/HTTP_logo.svg/220px-HTTP_logo.svg.png" width="200px">

<img src="https://grpc.io/img/logos/grpc-logo.png" width="200px">


Maintainers
-----------

This provider plugin is maintained by the sal, just sal for now.

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 0.14.x+
- [Go](https://golang.org/doc/install) 1.17 (to build the provider plugin)


Usage
---------------------

This provider is published here:

*  [https://registry.terraform.io/providers/salrashid123/grpc-full/latest](https://registry.terraform.io/providers/salrashid123/grpc-full/latest)


To use this, you must first provide the compiled descriptor you intend to use. 

You can compile the descriptors easily using `protoc` as shown below:

```
 protoc  \
    --descriptor_set_out=src/echo/echo.pb   src/echo/echo.proto
```

Once compiled, you can use this provider after setting some required values:

* `url`:  this is the endpoint to call.  You must provide the fully qualified URL including the service name and method
* `ca`: this is the certificate authority that signed the server cert for TLS connections
* `registry_files`: this is a list of the compiled descriptors to load.  You must set the `@type` key
* `request_type`: the message type sent to the server
* `response_type`: the message sent by the server

For an end-to-end example, see the repos `example/` folder for instructions

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
  ca                 = file("${path.module}/certs/tls-ca-chain.pem")
  sni                = "localhost"
  request_timeout_ms = 1000

  registry_files = [
    filebase64("${path.module}/src/echo/echo.pb"),
  ]

  request_headers = {
    authorization = "bearer foo"
  }

  request_body = jsonencode({
    "@type"    = "echo.EchoRequest",
    first_name = "sal",
    last_name  = "amander"
  })
  request_type  = "echo.EchoRequest"
  response_type = "echo.EchoReply"
}

output "data" {
  value = jsondecode(data.grpc.example.payload).message
}
```

For other configurations, see [example/index.md](blob/main/docs/index.md)

Building the DEV Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/salrashid123/terraform-provider-grpc-full

```sh
mkdir -p $GOPATH/src/github.com/terraform-providers
cd $GOPATH/src/github.com/terraform-providers
git clone https://github.com/salrashid123/terraform-provider-grpc-full.git
```

Enter the provider directory and build the provider

```sh
cd $GOPATH/src/github.com/terraform-providers/terraform-provider-grpc-full
make fmt
make build
```

Using the DEV provider
----------------------

Copy the provider to your directory

```bash
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/salrashid123/http-grpc/5.0.0/linux_amd64/
cp $GOBIN/terraform-provider-http-full ~/.terraform.d/plugins/registry.terraform.io/salrashid123/grpc-full/5.0.0/linux_amd64/terraform-provider-grpc-full_v5.0.0
```

Then

```bash
cd example
terraform init

terraform apply
```

with

```hcl
terraform {
  required_providers {
    grpc-full = {
      source  = "registry.terraform.io/salrashid123/grpc-full"
      version = "~> 5.0.0"
    }
  }
}

provider "grpc-full" {
}
 
data "http" "example" {
  provider = grpc-full
  
  url = "https://localhost:50051/echo.EchoSever/SayHello"
  ca = file("${path.module}/../certs/tls-ca-chain.pem")
  request_headers = {
    authorization = "bearer foo"
  }

  request_body = jsonencode({
    @type = "echo.EchoRequest",
    first_name = "sal"
    last_name = "amander"
  })
}

output "data" {
  value = jsondecode(data.grpc.example.body)
}
```


...

In order to test the provider, you can simply run `make test`.


### TEST

```sh
$ make test
```

