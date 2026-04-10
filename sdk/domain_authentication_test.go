package sendgrid_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
)

func Test_parseDomainAuthentication(t *testing.T) { //nolint:funlen
	type args struct {
		respBody string
	}

	tests := []struct {
		name  string
		args  args
		want  *sendgrid.DomainAuthentication
		want1 sendgrid.RequestError
	}{
		{
			name: "not automatic security",
			args: args{respBody: `{"id":123,"user_id":234,"subdomain":"aaa","domain":"example.com","username":"user","ips":[],"custom_spf":false,"default":false,"legacy":false,"automatic_security":false,"valid":false,"dns":{"mail_server":{"valid":false,"type":"mx","host":"aaa.example.com","data":"mx.sendgrid.net."},"subdomain_spf":{"valid":false,"type":"txt","host":"bbb.example.com","data":"v=spf1 include:sendgrid.net ~all"},"dkim":{"valid":false,"type":"txt","host":"m1._domainkey.aaa.example.com","data":"k=rsa; t=s; p=ccc"}}}`}, //nolint:lll
			want: &sendgrid.DomainAuthentication{
				ID:                 123,
				UserID:             234,
				Domain:             "example.com",
				Subdomain:          "aaa",
				Username:           "user",
				IPs:                []string{},
				CustomSPF:          false,
				IsDefault:          false,
				AutomaticSecurity:  false,
				CustomDKIMSelector: "",
				Legacy:             false,
				Valid:              false,
				DNS: sendgrid.DomainAuthenticationDNS{
					MailServer: sendgrid.DomainAuthenticationDNSValue{
						Valid: false,
						Type:  "mx",
						Host:  "aaa.example.com",
						Data:  "mx.sendgrid.net.",
					},
					DKIM: sendgrid.DomainAuthenticationDNSValue{
						Valid: false,
						Type:  "txt",
						Host:  "m1._domainkey.aaa.example.com",
						Data:  "k=rsa; t=s; p=ccc",
					},
					SubDomainSPF: sendgrid.DomainAuthenticationDNSValue{
						Valid: false,
						Type:  "txt",
						Host:  "bbb.example.com",
						Data:  "v=spf1 include:sendgrid.net ~all",
					},
				},
			},
			want1: sendgrid.RequestError{StatusCode: 200, Err: nil},
		},
		{
			name: "automatic security",
			args: args{respBody: `{"id":123,"user_id":234,"subdomain":"aaa","domain":"example.com","username":"user","ips":[],"custom_spf":false,"default":false,"legacy":false,"automatic_security":true,"valid":false,"dns":{"mail_cname":{"valid":false,"type":"cname","host":"aaa.example.com","data":"u234.abc.sendgrid.net"},"dkim1":{"valid":false,"type":"cname","host":"s1._domainkey.example.com","data":"s1.domainkey.u234.abc.sendgrid.net"},"dkim2":{"valid":false,"type":"cname","host":"s2._domainkey.example.com","data":"s2.domainkey.u234.abc.sendgrid.net"}}}`}, //nolint:lll
			want: &sendgrid.DomainAuthentication{
				ID:                 123,
				UserID:             234,
				Domain:             "example.com",
				Subdomain:          "aaa",
				Username:           "user",
				IPs:                []string{},
				CustomSPF:          false,
				IsDefault:          false,
				AutomaticSecurity:  true,
				CustomDKIMSelector: "",
				Legacy:             false,
				Valid:              false,
				DNS: sendgrid.DomainAuthenticationDNS{
					MailCNAME: sendgrid.DomainAuthenticationDNSValue{
						Valid: false,
						Type:  "cname",
						Host:  "aaa.example.com",
						Data:  "u234.abc.sendgrid.net",
					},
					DKIM1: sendgrid.DomainAuthenticationDNSValue{
						Valid: false,
						Type:  "cname",
						Host:  "s1._domainkey.example.com",
						Data:  "s1.domainkey.u234.abc.sendgrid.net",
					},
					DKIM2: sendgrid.DomainAuthenticationDNSValue{
						Valid: false,
						Type:  "cname",
						Host:  "s2._domainkey.example.com",
						Data:  "s2.domainkey.u234.abc.sendgrid.net",
					},
				},
			},
			want1: sendgrid.RequestError{StatusCode: 200, Err: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := sendgrid.ParseDomainAuthentication(tt.args.respBody)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDomainAuthentication() got = %v, want %v", got, tt.want)
			}
			if !errors.Is(got1.Err, tt.want1.Err) {
				t.Errorf("parseDomainAuthentication() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

var (
	mux    *http.ServeMux
	server *httptest.Server
)

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	return func() {
		server.Close()
	}
}
func TestClient_ValidateDomainAuthentication_ShouldNotErrorForOkResponseCode(t *testing.T) {
	teardown := setup()
	defer teardown()
	mux.HandleFunc("/whitelabel/domains/123/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"valid":true}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "on-behalf")

	requestError := client.ValidateDomainAuthentication(context.TODO(), "123")

	if requestError.Err != nil {
		t.Errorf("ValidateDomainAuthentication() got = %v, want %v", requestError, "error")
	}
}

func TestClient_ValidateDomainAuthentication_ShouldErrorForErrorResponseCode(t *testing.T) {
	teardown := setup()
	defer teardown()
	mux.HandleFunc("/whitelabel/domains/123/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"valid":true}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "on-behalf")

	requestError := client.ValidateDomainAuthentication(context.TODO(), "123")

	if requestError.Err == nil {
		t.Errorf("ValidateDomainAuthentication() got = %v, want %v", requestError, "error")
	}
}

func TestClient_ValidateDomainAuthentication_ShouldErrorForResponseBodySayingValidFalse(t *testing.T) {
	teardown := setup()
	defer teardown()
	mux.HandleFunc("/whitelabel/domains/123/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"valid":false}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "on-behalf")

	requestError := client.ValidateDomainAuthentication(context.Background(), "123")

	if requestError.Err == nil {
		t.Errorf("ValidateDomainAuthentication() got = %v, want %v", requestError, "error")
	}
	errMsg := sendgrid.ErrDomainAuthenticationValidationFailed.Error()
	if !errors.Is(requestError.Err, sendgrid.ErrDomainAuthenticationValidationFailed) {
		t.Errorf("ValidateDomainAuthentication() got = %v, want wrapped validation failed error", requestError.Err)
	}
	if !strings.Contains(requestError.Err.Error(), errMsg) {
		t.Errorf("ValidateDomainAuthentication() got = %v, want message containing %v", requestError.Err.Error(),
			errMsg)
	}
}
