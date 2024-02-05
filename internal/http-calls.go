package internal

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"regexp"
)

var sessionToken = regexp.MustCompile("<session-token>(.*)</session-token>")
var serverCert = regexp.MustCompile("<server-cert-hash>(.*)</server-cert-hash>")

const VERSION = "4.7.00136"

type AuthInitRequestResponse struct {
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

func PostAuthConfirmLogin(client *http.Client, auth *AuthInitRequestResponse, ssoToken, _url string) (string, string) {
	payload := `<?xml version="1.0" encoding="UTF-8"?>
				  <config-auth client="vpn" type="auth-reply" aggregate-auth-version="2">
					<version who="vpn">` + VERSION + `</version>
					<device-id>linux-64</device-id>
					<session-token/>
					<session-id/>
					<opaque is-for="sg">
						<tunnel-group>` + auth.Opaque.TunnelGroup + `</tunnel-group>
						<aggauth-handle>` + auth.Opaque.AggauthHandle + `</aggauth-handle>
						<auth-method>single-sign-on-v2</auth-method>
						<config-hash>` + auth.Opaque.ConfigHash + `</config-hash>
					</opaque>
					<auth>
						<sso-token>` + ssoToken + `</sso-token>
					</auth>
				  </config-auth>`

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

func PostAuthInitRequest(client *http.Client, _url string) *AuthInitRequestResponse {
	payload := `<?xml version="1.0" encoding="UTF-8"?>
				<config-auth client="vpn" type="init" aggregate-auth-version="2">
					<version who="vpn">` + VERSION + `</version>
					<device-id>linux-64</device-id>
					<group-select>` + `</group-select>
					<group-access>` + _url + `</group-access>
					<capabilities>
						<auth-method>single-sign-on-v2</auth-method>
					</capabilities>
				</config-auth>`

	post, err := client.Post(_url, `application/x-www-form-urlencoded`, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(post.Body)
	defer post.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var response AuthInitRequestResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}

	return &response
}
