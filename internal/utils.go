package internal

import (
	"context"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"net/url"
)

func BrowserCookieFinder(ctx context.Context, cookies chan string, name string) {
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

// thanks chatGPT for the refactoring of this function.
// i'm way too high to be doing this right now...!
func GetActualUrl(client *http.Client, targetUrl string) string {
	// Ensure the URL is properly formatted with the HTTPS protocol.
	uri, err := url.ParseRequestURI("https://" + targetUrl)
	if err != nil {
		log.Fatalf("Invalid URL format: %v", err)
	}

	// Create a new HTTP request.
	r, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		log.Fatalf("Could not create HTTP request: %v", err)
	}

	// Correctly set the Host header to the domain part without the protocol.
	r.Host = uri.Host

	do, err := client.Do(r)
	if err != nil {
		log.Println(err)
		log.Fatal("Could not perform request to fetch actual URL")
	}

	// returns the URL after all the redirections
	return do.Request.URL.String()
}

// chrome options
var opts = append(chromedp.DefaultExecAllocatorOptions[:],
	chromedp.Flag("headless", false), // Set headless mode to false
	chromedp.Flag("disable-gpu", false),
)

func CreateBrowserContext() (context.Context, context.CancelFunc) {
	// create context
	ctx, _ := chromedp.NewContext(context.Background())

	// Create an allocator
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)

	return chromedp.NewContext(allocCtx)
}
