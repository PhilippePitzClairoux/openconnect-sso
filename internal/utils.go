package internal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
)

type CustomHeaderTransport struct {
	Transport http.RoundTripper
	Headers   *map[string]string
}

func (cht *CustomHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	if cht.Headers != nil {
		for key, value := range *cht.Headers {
			req.Header.Add(key, value)
		}
	}
	return cht.Transport.RoundTrip(req)
}

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

func CookieFinder(ctx context.Context, cookies chan string, name string) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSentExtraInfo:
			for _, cookie := range ev.AssociatedCookies {
				if cookie.Cookie.Name == name {
					cookies <- cookie.Cookie.Value
				}
			}
		}
	})
}
