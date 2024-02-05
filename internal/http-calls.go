package internal

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

// Data extractors used instead of creating a struct to parse the content of an XML in AuthenticationConfirmation
var sessionToken = regexp.MustCompile("<session-token>(.*)</session-token>")
var serverCert = regexp.MustCompile("<server-cert-hash>(.*)</server-cert-hash>")

const postAuthConfirmLoginPayload = `<?xml version="1.0" encoding="UTF-8"?>
				  <config-auth client="vpn" type="auth-reply" aggregate-auth-version="2">
					<version who="vpn">%s</version>
					<device-id>linux-64</device-id>
					<session-token/>
					<session-id/>
					<opaque is-for="sg">
						<tunnel-group>%s</tunnel-group>
						<aggauth-handle>%s</aggauth-handle>
						<auth-method>single-sign-on-v2</auth-method>
						<config-hash>%s</config-hash>
					</opaque>
					<auth>
						<sso-token>%s</sso-token>
					</auth>
				  </config-auth>`

const postAuthInitRequestPayload = `<?xml version="1.0" encoding="UTF-8"?>
				<config-auth client="vpn" type="init" aggregate-auth-version="2">
					<version who="vpn">%s</version>
					<device-id>linux-64</device-id>
					<group-select></group-select>
					<group-access>%s</group-access>
					<capabilities>
						<auth-method>single-sign-on-v2</auth-method>
					</capabilities>
				</config-auth>`

const VERSION = "4.7.00136"

// AuthenticationInitExpectedResponse is a struct used to parse the XML payload
// we receive during the AuthenticationInit function. It contains valuable
// information that will be used throughout various parts of the program
type AuthenticationInitExpectedResponse struct {
	XMLName xml.Name `xml:"config-auth"`
	Opaque  struct {
		TunnelGroup   string `xml:"tunnel-group"`
		AggauthHandle string `xml:"aggauth-handle"`
		AuthMethod    string `xml:"auth-method"`
		ConfigHash    string `xml:"config-hash"`
	} `xml:"opaque"`

	Auth struct {
		Id                   string `xml:"id,attr"`
		Title                string `xml:"title"`
		Message              string `xml:"message"`
		Banner               string `xml:"banner"`
		SsoV2Login           string `xml:"sso-v2-login"`
		SsoV2LoginFinal      string `xml:"sso-v2-login-final"`
		SsoV2Logout          string `xml:"sso-v2-logout"`
		SsoV2LogoutFinal     string `xml:"sso-v2-logout-final"`
		SsoV2TokenCookieName string `xml:"sso-v2-token-cookie-name"`
		SsoV2ErrorCookieName string `xml:"sso-v2-error-cookie-name"`
		Form                 struct {
			Input struct {
				Value string `xml:",chardata"`
				Type  string `xml:"type,attr"`
				Name  string `xml:"name,attr"`
			} `xml:"input"`
		} `xml:"form"`
	} `xml:"auth"`

	Client               string `xml:"client,attr"`
	Type                 string `xml:"type,attr"`
	AggregateAuthVersion string `xml:"aggregate-auth-version,attr"`
}

// AuthenticationInit sends a http request to _url to get the actual URL and initiate SAML login request
func AuthenticationInit(client *http.Client, _url string) *AuthenticationInitExpectedResponse {
	payload := fmt.Sprintf(postAuthInitRequestPayload, VERSION, _url)

	post, err := client.Post(_url, `application/x-www-form-urlencoded`, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(post.Body)
	defer post.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var response AuthenticationInitExpectedResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}

	return &response
}

// AuthenticationConfirmation sends a http request to _url confirming the authentication was successfull
// (means we got the cookie and we're ready to start the next phase)
func AuthenticationConfirmation(client *http.Client, auth *AuthenticationInitExpectedResponse, ssoToken, _url string) (string, string) {
	payload := fmt.Sprintf(
		postAuthConfirmLoginPayload,
		VERSION,
		auth.Opaque.TunnelGroup,
		auth.Opaque.AggauthHandle,
		auth.Opaque.ConfigHash,
		ssoToken,
	)

	post, err := client.Post(_url, `application/x-www-form-urlencoded`, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(post.Body)
	defer post.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	token := sessionToken.FindStringSubmatch(string(body))
	cert := serverCert.FindStringSubmatch(string(body))

	if len(token) != 2 || len(cert) != 2 {
		log.Fatal("There was an issue while trying to extract token/cert...")
	}

	return token[1], cert[1]
}
