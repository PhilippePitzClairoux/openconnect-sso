package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/chromedp/chromedp"
	"go-openconnect-sso/internal"
	"log"
	"net/http"
	url2 "net/url"
	"os"
	"os/exec"
)

// flags
var server = flag.String("server", "", "server to connect to via openconnect")

func main() {
	flag.Parse()
	url := "https://" + *server
	ctx, cancel := chromedp.NewContext(context.Background())

	client := internal.NewHttpClient(*server)
	defer cancel()
	channel := make(chan string)

	url = getActualUrl(client, *server)
	samlAuth := internal.PostAuthInitRequest(client, url)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Set headless mode to false
		chromedp.Flag("disable-gpu", false),
	)

	// Create an allocator
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create a new browser context
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	//go listenBrowser(&ctx)
	go internal.CookieFinder(ctx, channel, samlAuth.Auth.SsoV2TokenCookieName)
	fmt.Println(samlAuth.Auth.SsoV2TokenCookieName)
	// initialize login and wait for user to finish auth
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(samlAuth.Auth.SsoV2Login),
	})
	if err != nil {
		log.Fatal(err)
	}

	// consume cookie and connect to vpn
	consumeCookies(channel, client, samlAuth, url, cancel)
}

// thanks chatGPT for the refactoring of this function.
// i'm way too high to be doing this right now...!
func getActualUrl(client *http.Client, url string) string {
	// Ensure the URL is properly formatted with the HTTPS protocol.
	uri, err := url2.ParseRequestURI("https://" + url)
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

	return do.Request.URL.String()
}

func consumeCookies(channel chan string, client *http.Client, auth *internal.AuthInitRequestResponse, _url string, cancel context.CancelFunc) {
	for cookie := range channel {
		token, cert := internal.PostAuthConfirmLogin(client, auth, cookie, _url)
		cancel() // close browser

		command := exec.Command("sudo",
			"openconnect",
			"--useragent",
			fmt.Sprintf("AnyConnect Linux_64 %s", internal.VERSION),
			fmt.Sprintf("--version-string"),
			internal.VERSION,
			"--cookie",
			token,
			"--servercert",
			cert,
			_url,
		)

		log.Println("Command : ", command.String())
		command.Stdout = os.Stdout
		command.Stderr = os.Stdout

		err := command.Run()
		if err != nil {
			log.Fatal("Could not start command : ", err)
		}
	}
}
