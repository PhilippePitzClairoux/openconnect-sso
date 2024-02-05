package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/chromedp/chromedp"
	"go-openconnect-sso/internal"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// flags
var server = flag.String("server", "", "server to connect to via openconnect")

func main() {
	flag.Parse()

	client := internal.NewHttpClient(*server)
	cookieFound := make(chan string)
	targetUrl := internal.GetActualUrl(client, *server)
	samlAuth := internal.AuthenticationInit(client, targetUrl)
	ctx, closeBrowser := internal.CreateBrowserContext()

	log.Println("Starting goroutine that searches for authentication cookie ", samlAuth.Auth.SsoV2TokenCookieName)
	go internal.BrowserCookieFinder(ctx, cookieFound, samlAuth.Auth.SsoV2TokenCookieName)

	log.Println("Open browser and navigate to SSO login page : ", samlAuth.Auth.SsoV2Login)
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(samlAuth.Auth.SsoV2Login),
	})
	if err != nil {
		log.Fatal(err)
	}

	// consume cookie and connect to vpn
	startVpnOnLoginCookie(cookieFound, client, samlAuth, targetUrl, closeBrowser)
}

// startVpnOnLoginCookie waits to get a cookie from the authenticationCookies channel before confirming
// the authentication process (to get token/cert) and then starting openconnect
func startVpnOnLoginCookie(authenticationCookies chan string, client *http.Client, auth *internal.AuthenticationInitExpectedResponse, targetUrl string, closeBrowser context.CancelFunc) {
	for cookie := range authenticationCookies {
		token, cert := internal.AuthenticationConfirmation(client, auth, cookie, targetUrl)
		closeBrowser() // close browser

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
			targetUrl,
		)

		command.Stdout = os.Stdout
		command.Stderr = os.Stdout
		command.Stdin = os.Stdin

		log.Println("Starting openconnect: ", command.String())
		err := command.Run()
		if err != nil {
			log.Fatal("Could not start command : ", err)
		}
	}
}
