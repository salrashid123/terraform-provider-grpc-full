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
    last_name  = "amander"
  })

}

output "data" {
  value = jsondecode(data.grpc.example.payload).message
}

