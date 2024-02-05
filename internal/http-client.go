package internal

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
)

// CustomHeaderTransport is used to add our custom http headers to http requests
type CustomHeaderTransport struct {
	Transport http.RoundTripper
	Headers   *map[string]string
}

// RoundTrip adds our custom headers and then uses the default http transport to RoundTrip
func (cht *CustomHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	if cht.Headers != nil {
		for key, value := range *cht.Headers {
			req.Header.Add(key, value)
		}
	}
	return cht.Transport.RoundTrip(req)
}

// NewHttpClient creates a new HTTP client and adds serverName to the TLS configuration.
// this part is important to have accurate and valid TLS sessions (solved an old issue with DTLS).
func NewHttpClient(serverName string) *http.Client {
	defaultTransport := http.DefaultTransport.(*http.Transport)
	systemCerts, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal("Could not load system cert pool.")
	}

	defaultTransport.TLSClientConfig = &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: false,
		RootCAs:            systemCerts,
	}

	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			log.Println("Redirected from", via[len(via)-1].URL.String(), "to", req.URL.String())
			return nil
		},
		Transport: &CustomHeaderTransport{
			Transport: defaultTransport,
			Headers: &map[string]string{
				"User-Agent":          "AnyConnect Linux_64 4.7.00136",
				"Accept":              "*/*",
				"Accept-Encoding":     "identity",
				"X-Transcend-Version": "1",
				"X-Aggregate-Auth":    "1",
				"X-Support-HTTP-Auth": "true",
				"Content-Type":        "application/x-www-form-urlencoded",
				"Connection":          "keep-alive",
			},
		},
	}
}
