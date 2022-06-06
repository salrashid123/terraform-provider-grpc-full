## Test cases for gRPC services

Included in this provider is a sample gRPC client-server to test with

### Start Server

```bash
cd example/
go run src/grpc_server.go
```

```
# test standalone client
go run src/client/main.go
```

Then apply the terraform provider using the configuration in the repo

```
terraform init
terraform apply

$ terraform apply

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

data = "Hello sal a amander"

```

The output will show a sample response from the gRPC server

