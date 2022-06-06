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

  // note, the following a live gRPC server of the same proto running in cloud run
  // you can invoke this endpoint but note that the requests are logged (not that i'd care to do anything with them but thats just what cloud run does..)
  # url                = "https://grpc-server-6w42z6vi3q-uc.a.run.app:443/echo.EchoServer/SayHello"
  # sni                = "grpc-server-6w42z6vi3q-uc.a.run.app"

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

