package internal

import (
	"log"
	"net/http"
	"net/url"
	"os"
)

// thanks chatGPT for the refactoring of this function.
// i'm way too high to be doing this right now...!
func getActualUrl(client *http.Client, targetUrl string) string {
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

func (oc *OpenConnectCtx) handleExit() {
	sig := <-oc.exitChan

	log.Println("Closing Browser...")
	oc.closeBrowser()

	if oc.process != nil {
		err := oc.process.Cancel()
		log.Printf("Closing openconnect process (error : %s)\n", err)
	}

	log.Printf("Got an exit signal (%s)! Cya!", sig.String())
	os.Exit(0)
}
