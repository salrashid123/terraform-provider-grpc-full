package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/psanford/lencode"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
)

func dataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRead,

		Schema: map[string]*schema.Schema{

			"registry_files": {
				Type:     schema.TypeList,
				Computed: false,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"url": {
				Type:     schema.TypeString,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"sni": {
				Type:     schema.TypeString,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"request_headers": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"request_body": {
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"payload": {
				Type:     schema.TypeString,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"insecure_skip_verify": {
				Type:     schema.TypeBool,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
				Default: false,
			},
			"request_timeout_ms": {
				Type:     schema.TypeInt,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			"request_type": {
				Type:     schema.TypeString,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"response_type": {
				Type:     schema.TypeString,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"response_headers": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status_code": {
				Type:     schema.TypeInt,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"ca": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	url := d.Get("url").(string)
	sni := d.Get("sni").(string)
	request_type := d.Get("request_type").(string)
	response_type := d.Get("response_type").(string)
	headers := d.Get("request_headers").(map[string]interface{})

	pbFiles := d.Get("registry_files").([]interface{})

	for _, fileContentB64 := range pbFiles {

		fc, ok := fileContentB64.(string)
		if !ok {
			return append(diags, diag.Errorf("Error converting filecontent to string")...)
		}
		fileContent, err := base64.StdEncoding.DecodeString(fc)
		if err != nil {
			return append(diags, diag.Errorf("Error decoding file .pb ")...)
		}
		fileDescriptors := &descriptorpb.FileDescriptorSet{}
		err = proto.Unmarshal(fileContent, fileDescriptors)
		if err != nil {
			return append(diags, diag.Errorf("Error unmarshaling .pb files")...)
		}
		for _, pb := range fileDescriptors.GetFile() {
			var fdr protoreflect.FileDescriptor
			fdr, err = protodesc.NewFile(pb, protoregistry.GlobalFiles)
			if err != nil {
				return append(diags, diag.Errorf("Error getting proto files")...)
			}
			fmt.Printf("Loading package %s\n", fdr.Package().Name())
			dd, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(request_type))
			if err != nil {
				// todo catch the not found and continue
				//	return append(diags, diag.Errorf(fmt.Sprintf("Error  FindDescriptorByName proto file %s error: [%s]", request_type, err.Error()))...)
			}
			if dd == nil {

				err = protoregistry.GlobalFiles.RegisterFile(fdr)
				if err != nil {
					return append(diags, diag.Errorf("Error Registering proto file")...)
				}
				for _, m := range pb.MessageType {
					md := fdr.Messages().ByName(protoreflect.Name(*m.Name))
					mdType := dynamicpb.NewMessageType(md)

					err = protoregistry.GlobalTypes.RegisterMessage(mdType)
					if err != nil {
						return append(diags, diag.Errorf("Error registering message")...)
					}
				}
			}
		}
	}

	requestMessageType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(request_type))
	if err != nil {
		return append(diags, diag.Errorf("Error finding request message type")...)
	}
	if requestMessageType == nil {
		return append(diags, diag.Errorf("Error finding request message type")...)
	}
	//requestMessageDescriptor := requestMessageType.Descriptor()
	// reflectRequest := requestMessageType.New()
	// in, err := proto.Marshal(reflectRequest.Interface())
	// if err != nil {
	// 	return append(diags, diag.Errorf("Error generating reflectRequest")...)
	// }

	var skip_verify bool
	skip_verify_override, ok := d.GetOk("insecure_skip_verify")
	if ok {
		if skip_verify, ok = skip_verify_override.(bool); !ok {
			return append(diags, diag.Errorf("Error overriding skip_verify_override")...)
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: skip_verify,
		ServerName:         sni,
	}
	castr, ok := d.GetOk("ca")
	if ok {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(castr.(string)))
		tlsConfig.RootCAs = caCertPool
	}

	client := http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	timeout_override, ok := d.GetOk("request_timeout_ms")
	if ok {
		var timeout int
		if timeout, ok = timeout_override.(int); !ok {
			return append(diags, diag.Errorf("Error overriding request_timeout_ms")...)
		}
		client.Timeout = time.Duration(timeout) * time.Millisecond
	}

	request_body, ok := d.GetOk("request_body")
	if !ok {
		return append(diags, diag.Errorf("Error reading request_body")...)
	}

	a, err := anypb.New(requestMessageType.New().Interface())
	if err != nil {
		return append(diags, diag.Errorf("Error setting HTTP response body: %s", err)...)
	}

	rbody := bytes.NewReader([]byte(request_body.(string)))

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(rbody)
	if err != nil {
		return append(diags, diag.Errorf("Error reading request buffer: %s", err)...)
	}
	err = protojson.Unmarshal(buf.Bytes(), a)
	if err != nil {
		return append(diags, diag.Errorf("Error setting HTTP response body: %s", err)...)
	}

	var out bytes.Buffer
	enc := lencode.NewEncoder(&out, lencode.SeparatorOpt([]byte{0}))
	err = enc.Encode(a.Value)
	if err != nil {
		return append(diags, diag.Errorf("Error lencoding request: %s", err)...)
	}

	reader := bytes.NewReader(out.Bytes())

	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return append(diags, diag.Errorf("Error creating http client: %s", err)...)
	}
	req.Header.Set("content-type", "application/grpc")

	for name, value := range headers {
		v, ok := value.(string)
		if !ok {
			return append(diags, diag.Errorf("Error converting header [%s] to string: %s", name, err)...)
		}
		req.Header.Set(name, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return append(diags, diag.Errorf("Error creating grpcCall: %s", err)...)
	}
	if resp.StatusCode != http.StatusOK {
		return append(diags, diag.Errorf("Error grpcCall status !=StatusOK  got: %s", resp.Status)...)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return append(diags, diag.Errorf("Error setting HTTP response body: %s", err)...)
	}

	bytesReader := bytes.NewReader(bodyBytes)
	// now unpack the wiremessage to get to the unary response
	respMessage := lencode.NewDecoder(bytesReader, lencode.SeparatorOpt([]byte{0}))
	respMessageBytes, err := respMessage.Decode()
	if err != nil {
		return append(diags, diag.Errorf("Error reading respMessageBytes: %s", err)...)
	}

	replyMessageType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(response_type))
	if err != nil {
		return append(diags, diag.Errorf("Error getting replyMessageType: %s", err)...)
	}

	pmr := replyMessageType.New()

	err = proto.Unmarshal(respMessageBytes, pmr.Interface())
	if err != nil {
		return append(diags, diag.Errorf("Error setting decoding lencode body: %s", err)...)
	}

	s, err := protojson.Marshal(pmr.Interface())
	if err != nil {
		return append(diags, diag.Errorf("Error unmarshalling protojson response: %s", err)...)
	}

	if err = d.Set("status_code", resp.StatusCode); err != nil {
		return append(diags, diag.Errorf("Error setting HTTP status_code: %s", err)...)
	}

	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		// Concatenate according to RFC2616
		// cf. https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2
		responseHeaders[k] = strings.Join(v, ", ")
	}
	if err = d.Set("response_headers", responseHeaders); err != nil {
		return append(diags, diag.Errorf("Error setting HTTP response headers: %s", err)...)
	}

	if err = d.Set("payload", string(s)); err != nil {
		return append(diags, diag.Errorf("Error setting HTTP response body: %s", err)...)
	}

	// set ID as something more stable than time
	d.SetId(url)

	return diags
}
