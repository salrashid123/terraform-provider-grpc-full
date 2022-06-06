package provider

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	echo "github.com/salrashid123/grpc_wireformat/grpc_services/src/echo"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (

	// X509v3 extensions:
	// X509v3 Key Usage: critical
	// 	Certificate Sign, CRL Sign
	// X509v3 Basic Constraints: critical
	// 	CA:TRUE, pathlen:0
	// X509v3 Subject Key Identifier:
	// 	B7:BA:B0:02:A1:E7:BE:34:C6:C1:05:5C:66:78:E5:BB:53:5D:A1:54
	caCert = `-----BEGIN CERTIFICATE-----\nMIIDfjCCAmagAwIBAgIBATANBgkqhkiG9w0BAQsFADBQMQswCQYDVQQGEwJVUzEP\nMA0GA1UECgwGR29vZ2xlMRMwEQYDVQQLDApFbnRlcnByaXNlMRswGQYDVQQDDBJF\nbnRlcnByaXNlIFJvb3QgQ0EwHhcNMjIwNTI2MjI1NjI1WhcNMzIwNTI1MjI1NjI1\nWjBQMQswCQYDVQQGEwJVUzEPMA0GA1UECgwGR29vZ2xlMRMwEQYDVQQLDApFbnRl\ncnByaXNlMRswGQYDVQQDDBJFbnRlcnByaXNlIFJvb3QgQ0EwggEiMA0GCSqGSIb3\nDQEBAQUAA4IBDwAwggEKAoIBAQDQ+bpQHaJQWggUoPXVf/7xqLsOPH5D83MDU8l1\ndamAGe7yhZp4leU5hC6KUs8hqA9NQ67WUEOmzS00D01DfsKHsJo9mbufaHN3ij4l\nIDMqJJOgOTvdz3cEfAFhq2syEjqk1ghEwGJhZ2tdh0LORwLUYfoXgYs0w6m6++z2\nkvLZ4G0EgraqsmpjfFXBRDN/OsBdy68jmZBS9LFo/KZu0KH3/ZKAih39SFNOtKNx\n9gXvF7PJ+KOnWEAjuXpQJDNBF7S9WBDEBaIR+qdY5B5oGzzkcGuOlWbqUWfAXMyb\n7WrWODMf8FS8JHVTAN0eLVmnP0Ibqzvtk48oc7NgTg24O5ZzAgMBAAGjYzBhMA4G\nA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBS7tGUlTrMJ\nafcmmZwFqGq5ktD4ZTAfBgNVHSMEGDAWgBS7tGUlTrMJafcmmZwFqGq5ktD4ZTAN\nBgkqhkiG9w0BAQsFAAOCAQEAnE5jWnXIa6hGJKrIUVHhCxdJ4CDpayKiULhjPipR\nTZxzOlbhJHM/eYfH8VtbHRLkZrG/u3uiGWinLliXWHR9cB+BRgdVOMeehDMKP6o0\nWoACUpyLsbiPKdTUEXzXg4MwLwv23vf2xWvp4TousLA8++rIk1qeFW0NSAUGzYfs\nsKpBP2BdJVXcveAEpfwmbnQTZ0OzceA4RFdu4hMZhOwXgK2WZh4fMhyRBh67ueFh\nkVEGN4UUVAP4r/pJEtf4lLE468yPdD+w0yM0xDVAb9DrMyr3h4FwxHalZdgOeRSq\nATCK3GKv5lwmr/NPdg/cPdG5p/lfWQACwi47XgGi59nYIw==\n-----END CERTIFICATE-----`
	// 	caKey = `-----BEGIN PRIVATE KEY-----
	// MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDQ+bpQHaJQWggU
	// oPXVf/7xqLsOPH5D83MDU8l1damAGe7yhZp4leU5hC6KUs8hqA9NQ67WUEOmzS00
	// D01DfsKHsJo9mbufaHN3ij4lIDMqJJOgOTvdz3cEfAFhq2syEjqk1ghEwGJhZ2td
	// h0LORwLUYfoXgYs0w6m6++z2kvLZ4G0EgraqsmpjfFXBRDN/OsBdy68jmZBS9LFo
	// /KZu0KH3/ZKAih39SFNOtKNx9gXvF7PJ+KOnWEAjuXpQJDNBF7S9WBDEBaIR+qdY
	// 5B5oGzzkcGuOlWbqUWfAXMyb7WrWODMf8FS8JHVTAN0eLVmnP0Ibqzvtk48oc7Ng
	// Tg24O5ZzAgMBAAECggEAMR5BeIs+l3xR4edjYOdQ2SQ7s0DsvLQAGIwdEgqx6HYv
	// /7j/cdBprHcxKToFjXefAR4jfiQngpE/Srk+A9tLhfEwj8IOo409dp97s+Y5oHIw
	// cLyDIcOdyeQLvxU3gPFf71aPYvmFJjfUuIsOXMW8GIde7R95xNEol9aW/+3SPvtf
	// 3b86gVugWvWbUhGWKWBTW2VQnQVo+MZEy4R3OybFuwxnMasOeLUYQb5RceOEWO3Z
	// m6UVLPd+vrt5uKtCJIPhU+39Vw8WYiSEysvmIT/p9yaC0Nlydaa5sBQ7CwqRLIqu
	// Fw+I6SriwoABXCwRoRbOql+HP6+uXJSBm/1j4IGIoQKBgQD/qA11sonVxCanLgXG
	// GOGFbkvhtJ3IYucTkXh1pI+ZHHcK/frvtL+dpukEOFMNTEmOfJnBA/VVpDDuMmIk
	// pF+5qSYqnAxXn7oOynYVsTc6QC+F59UjA+wdwag92GHH7brWJjsa9vyg8rDtEAOk
	// jVZhbf8lBHMjl3B2PVBZqz/k1QKBgQDRQZ3d1/LW2zJ6ZY9CNR1jsKsDCoEZZrnU
	// lq5beWK5MFBQXeS7v6qptYNgM8VNYsCK2mN1YjqA1JYI0sGtj4XTumWYsaOUNuhE
	// Rp++LOsi5eySkRNOOAJAHB9VRiT+U+rwUwZkxMKWZbekmhJYJe9DcUZ8oeHD+btG
	// b2OMESXSJwKBgCEyczz7SAaoB9ThlwJYLMCkx9mxGGPy48qYsymjirn5BkQ5IqKJ
	// t/ACwnM31SD+7PZBm72ChBLw1SG5DSFw7rUvD7Osu7WNGh3dkGPUtTUtLH6Y0gZP
	// 9hMPGIefV2McrYwtPrOLqtZDbVH7KF3vtG3GWME3yLOwcHwKDir2n79ZAoGAR1pT
	// hVDcekz2EmxNBCtuYQ7d0USkrs+rcAUNYR2r/y+tQyoxE6AQhpvhN02P6opQ00gS
	// f/VFs6ZJnqqW5iK5ZG/7sqxn9eMfIiDe2Y8hgp3aJEQZzCMnCUtNl9s6RArDYr08
	// weGh5Hy8uQDcXnhY9KtMeLUOca/XHvZegGVceyMCgYBPW0VHLEULeHijCOI1qcmg
	// 01z6fU3IEBNle80VO5e3+QRjqZnCdOAj0PHqRhpceFkr+QDqlC+9o2eGlicVSE/s
	// 7Mc467clydBvvkyo7DVtQgivcGdOuVd0kEJVMbC9mcrF4pSfiuhxTqQVmJ+B2QO6
	// hiAA1onD1iT72j0DOayXbw==
	// -----END PRIVATE KEY-----`

	// X509v3 extensions:
	// X509v3 Extended Key Usage:
	// 	TLS Web Server Authentication
	// X509v3 Authority Key Identifier:
	// 	keyid:B7:BA:B0:02:A1:E7:BE:34:C6:C1:05:5C:66:78:E5:BB:53:5D:A1:54
	// X509v3 Subject Alternative Name:
	// 	DNS:localhost
	localhostCert = `-----BEGIN CERTIFICATE-----
MIIELzCCAxegAwIBAgIBAjANBgkqhkiG9w0BAQsFADBQMQswCQYDVQQGEwJVUzEP
MA0GA1UECgwGR29vZ2xlMRMwEQYDVQQLDApFbnRlcnByaXNlMRswGQYDVQQDDBJF
bnRlcnByaXNlIFJvb3QgQ0EwHhcNMjIwNTI2MjMwNjIzWhcNMzIwNTI1MjMwNjIz
WjBPMQswCQYDVQQGEwJVUzEPMA0GA1UECgwGR29vZ2xlMRMwEQYDVQQLDApFbnRl
cnByaXNlMRowGAYDVQQDDBFzZXJ2ZXIuZG9tYWluLmNvbTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBALk1Zr85ztqUagPPBJl/m7g+GBcend+JEdmVa9J3
zP7/MBV+kJymdZ1DWeKdXK3CEOqrH3/vHTsCMyDX6H671LlTnBls6ZdDP10ujCds
AHbTrFUfD9U4QPtYkL0J0PIHjYGHnHdOkeRQuE8tBx1bRgVJMSsYSFiaZDVI5B3A
050I41YlEZc6Fq8NIcLig3j2ycqC9eLaDmrKNayRhXBm+N31S26ni3uJUH3sFn7l
Vt63BGv1o3xbcRv8TRCrLzZb18GbpAG3x5hSbQBn5GJhXDXzeNhdVUE1NPWKtlMs
sJ6XBiARqrgoHtanae9f1xCbMkMn+wdjhviIuk7S4t0yGcsCAwEAAaOCARMwggEP
MA4GA1UdDwEB/wQEAwIHgDAJBgNVHRMEAjAAMBMGA1UdJQQMMAoGCCsGAQUFBwMB
MB0GA1UdDgQWBBRdGJsQ0CJAliuiEka8wJlg+hj5CjAfBgNVHSMEGDAWgBS7tGUl
TrMJafcmmZwFqGq5ktD4ZTBFBggrBgEFBQcBAQQ5MDcwNQYIKwYBBQUHMAKGKWh0
dHA6Ly9wa2kuZXNvZGVtb2FwcDIuY29tL2NhL3Jvb3QtY2EuY2VyMDoGA1UdHwQz
MDEwL6AtoCuGKWh0dHA6Ly9wa2kuZXNvZGVtb2FwcDIuY29tL2NhL3Jvb3QtY2Eu
Y3JsMBoGA1UdEQQTMBGHBH8AAAGCCWxvY2FsaG9zdDANBgkqhkiG9w0BAQsFAAOC
AQEAqzmJ1CnygTMKvxkIbQKiYtBLlDAA7tJO+55mcivtk9RxT+PBYEnEV1e+IqHC
4x+YLESVCHQk1Eia3dJy3fEAylRxzICxMrg+4EA2nuDSgVH8CeD74kUEsEzSw8eY
SQH2RoOxly+32+lkw2oF1a38+elMvU/0Z2w1F2CW5sVj1kieG4vzn0rqvmNauU04
r1m6rnN1yq6rtuQ16Y9SQb1VGXs9ijNKMICGcBONqebYCV7nGPCitH5yQsIXy8ns
DoHHMPWPLdj8n6w9drtKeBN4IHooizAuv43HbWapVgVAKsLxffo1B7DcgQDB0MYn
Jh1J6CU1KiOziz1rQrambz66Jw==
-----END CERTIFICATE-----`

	localhostKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAuTVmvznO2pRqA88EmX+buD4YFx6d34kR2ZVr0nfM/v8wFX6Q
nKZ1nUNZ4p1crcIQ6qsff+8dOwIzINfofrvUuVOcGWzpl0M/XS6MJ2wAdtOsVR8P
1ThA+1iQvQnQ8geNgYecd06R5FC4Ty0HHVtGBUkxKxhIWJpkNUjkHcDTnQjjViUR
lzoWrw0hwuKDePbJyoL14toOaso1rJGFcGb43fVLbqeLe4lQfewWfuVW3rcEa/Wj
fFtxG/xNEKsvNlvXwZukAbfHmFJtAGfkYmFcNfN42F1VQTU09Yq2UyywnpcGIBGq
uCge1qdp71/XEJsyQyf7B2OG+Ii6TtLi3TIZywIDAQABAoIBAD05Cd3snhRjOyhH
Jp4XMMKWxB/gXw+ln+DtI9dPAtTIRnzUeblOzVJPEUd3/Ury++SW7LK9uEvpTj1t
Ic3DCW651MAS4KS/9hI3cN0XNpARKMZ6niE9lz1+6VmUBR38oSpQScimkFOI22RQ
3ik2Is9cgoRcYo3ne3ihv8aWF12xIbcP9MjEGh/uTaRZP/PMv3EmDL/5VCN91laB
9ZtTgeaMjTt53fNrz/PKLFA4qTuXkHlnbxMUp1uW+kHpwH8XY/lhYjjlSjOPQg5y
QsrObOqOExnfHa7Rds0FT05XsbV5TmRszF2I9fIcnUp0uwIyQ6iTTQzJIufkADQ7
Jg7HgwECgYEA8a6yG66GvA40JAFp+UK5cl7qKo7YJ2zLPxtRbFW7nZ22jW26u3Fq
FPhBxSHtn4UfsQTqqpXp04KASS7UyUt6ri2raEzUdSaX/uCMtB9fm+5ixpUJOLO9
Q9HCkJNFchktCeYlet89MJ/zkl2JSyTGez32JqHSfNRPQ5ds6tIr4QUCgYEAxC48
Xlx+wjDRpW3SG6QgwPImPjYQCK8hcltEfhltX9veFUJCWYZqEGtD10phS1dQF9WO
FhTpLFgWCaNlrywD6jsDMS+xv8+VfsooIAlThuKpTZWMLQasDwfYc2n7qcHJOwYs
pnX4sQNvN4iMvrL0CADDiSVmH99SU0NdVtauSI8CgYEAjf7vBFaZMNpDhjgSdHHg
lTLw8Ao3M4q3K5+4Sidg8O0duaCTyteK1UE7G0Cg5U2I3i+eVJV56VxOVTEfshkX
vkh04fXqCd6gBQ8XfCjGus3n2PbtkRQBilwurVTpw2zJSnye3r9Uq0H/EKrGJJE5
0GUKP45qJg9zdqn8Q0cyoqUCgYA8Yi7arIWnp/cfgCoHsAEU4nO6+lD9G0qkNEtk
tNbhhn9Y88gQXjsPSrTa8133HqzcaTMOwOj0aTh/RvfpbxbVZcyZuyBu9aoCGJ85
HSXEgsexxbIbuc4D4lpRS/HWUntp24Cqy+z8Lx5wbWtE1zgdrn6BHC3O6aIhVr7I
F9QVKQKBgQDBAbYkyc41baRJm6oKIfDUfjNiYdggsKm3OmL4qs+snqKUHv0WF/aI
U2gGbAvI7b/PKnMs4zct0JoFkZ2MN369PEhcpKnQ6iM8Le+42r8OPhpOs/M0uCeP
x1mpmqYktTV/Y7QcHQsrVaZ+4WWYO0+Erp1mYrm7EUOhvGOGrp/6mw==
-----END RSA PRIVATE KEY-----`

	// package echo;
	// cat src/echo/echo.pb |base64
	echopb = `CuQCChNzcmMvZWNoby9lY2hvLnByb3RvEgRlY2hvIhwKBk1pZGRsZRISCgRuYW1lGAEgASgJUgRuYW1lIngKC0VjaG9SZXF1ZXN0Eh0KCmZpcnN0X25hbWUYASABKAlSCWZpcnN0TmFtZRIbCglsYXN0X25hbWUYAiABKAlSCGxhc3ROYW1lEi0KC21pZGRsZV9uYW1lGAMgASgLMgwuZWNoby5NaWRkbGVSCm1pZGRsZU5hbWUiJQoJRWNob1JlcGx5EhgKB21lc3NhZ2UYASABKAlSB21lc3NhZ2UyPgoKRWNob1NlcnZlchIwCghTYXlIZWxsbxIRLmVjaG8uRWNob1JlcXVlc3QaDy5lY2hvLkVjaG9SZXBseSIAQkBaPmdpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dycGNfd2lyZWZvcm1hdC9ncnBjX3NlcnZpY2VzL3NyYy9lY2hvYgZwcm90bzM=`
)

type Server struct {
	echo.UnimplementedEchoServerServer
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

type TestGrpcMock struct {
	server  *grpc.Server
	Address string
}

const testDataSourceConfig_basic = `
data "grpc" "example" {

  url                = "https://%s/echo.EchoServer/SayHello"
  ca                 = "%s"
  sni                = "localhost"

  registry_files = [
    "%s",
  ]


  request_type  = "echo.EchoRequest"
  response_type = "echo.EchoReply"
  request_body = jsonencode({
    "@type"    = "echo.EchoRequest",
    first_name = "sal",
    last_name  = "mander",
    middle_name = {
		name = "a"
	} 	
  })

}

output "data" {
  value = jsondecode(data.grpc.example.payload).message
}
`

func TestDataSource_test_basic(t *testing.T) {
	testHttpMock, err := setUpMockGRPCServer([]byte(localhostCert), []byte(localhostKey))
	if err != nil {
		t.Fatal(err)
	}
	defer testHttpMock.server.GracefulStop()
	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceConfig_basic, testHttpMock.Address, caCert, echopb),
				// ExpectError: regexp.MustCompile("x509: certificate signed by unknown authority"),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["data.grpc.example"]
					if !ok {
						return fmt.Errorf("missing data resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["data"].Value != "Hello sal a mander" {
						return fmt.Errorf(
							`'data' output is %s; want 'Hello sal a mander'`,
							outputs["data"].Value,
						)
					}

					return nil
				},
			},
		},
	})
}

const testDataSourceConfig_error = `
data "grpc" "example" {

  url                = "https://%s/echo.EchoServer/SayHello"
  ca                 = "%s"
  sni                = "localhost"

  registry_files = [
    "%s",
  ]

  request_type  = "foo.EchoRequest"
  response_type = "foo.EchoReply"
  request_body = jsonencode({
    "@type"    = "foo.EchoRequest",
    first_name = "sal",
    last_name  = "mander",
    middle_name = {
		name = "a"
	} 	
  })

}

output "data" {
  value = jsondecode(data.grpc.example.payload).message
}
`

func TestDataSource_test_error(t *testing.T) {
	testHttpMock, err := setUpMockGRPCServer([]byte(localhostCert), []byte(localhostKey))
	if err != nil {
		t.Fatal(err)
	}
	defer testHttpMock.server.GracefulStop()
	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testDataSourceConfig_error, testHttpMock.Address, caCert, echopb),
				ExpectError: regexp.MustCompile("Error finding request message type"),
			},
		},
	})
}

func (s *Server) SayHello(ctx context.Context, in *echo.EchoRequest) (*echo.EchoReply, error) {
	mname := ""
	m := in.MiddleName
	if m != nil {
		mname = m.Name
	}
	return &echo.EchoReply{Message: "Hello " + in.FirstName + " " + mname + " " + in.LastName}, nil
}

func setUpMockGRPCServer(tlsCert []byte, tlsKey []byte) (*TestGrpcMock, error) {

	formatCaCert := strings.Replace(caCert, `\n`, "\n", -1)
	clientCaCertPool := x509.NewCertPool()
	ok := clientCaCertPool.AppendCertsFromPEM([]byte(formatCaCert))
	if !ok {
		return nil, errors.New("Could not load cacert")
	}
	privBlock, _ := pem.Decode([]byte(localhostKey))
	key, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, err
	}

	pubBlock, _ := pem.Decode([]byte(localhostCert))
	cert, err := x509.ParseCertificate(pubBlock.Bytes)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				PrivateKey:  key,
				Certificate: [][]byte{cert.Raw},
			},
		},
	}

	tlsConfig.BuildNameToCertificate()

	creds := credentials.NewTLS(tlsConfig)

	l, err := net.Listen("tcp", "localhost:0") // IIRC 0 == "first available port"
	if err != nil {
		return nil, err
	}

	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	sopts = append(sopts, grpc.Creds(creds))

	s := grpc.NewServer(sopts...)
	srv := NewServer()
	echo.RegisterEchoServerServer(s, srv)

	fakeGreeterAddr := l.Addr().String()
	go func() {
		if err := s.Serve(l); err != nil {
			panic(err)
		}
	}()

	return &TestGrpcMock{
		server:  s,
		Address: fakeGreeterAddr,
	}, nil
}
