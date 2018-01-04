package oauth

import (
	"strings"
	"testing"
)


const json_file = `{
	"installed":{
		"auth_uri":"https://accounts.google.com/o/oauth2/auth",
		"client_secret":"not telling",
		"token_uri":"https://accounts.google.com/o/oauth2/token",
		"client_email":"",
		"redirect_uris":["urn:ietf:wg:oauth:2.0:oob","oob"],
		"client_x509_cert_url":"",
		"client_id":"also not telling",
		"auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs"
	}
}
`

func TestConfigExtractor(t *testing.T) {
	reader := strings.NewReader(json_file)
	wanted := map[string]string{"client_secret":""}
	FindProperties(reader, wanted)
	if got, expected := wanted["client_secret"], "not telling"; got != expected {
		t.Fatalf("got %v, wanted %v", got, wanted)
	}
}
