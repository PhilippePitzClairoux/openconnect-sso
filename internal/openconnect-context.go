package internal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
)

type OpenConnectCtx struct {
	process         *exec.Cmd
	client          *http.Client
	exit            chan os.Signal
	cookieFoundChan chan string
	exitChan        chan os.Signal
	targetUrl       string
	server          string
	username        string
	password        string
	browserCtx      context.Context
	closeBrowser    context.CancelFunc
	trace           bool
}

func NewOpenconnectCtx(server, username, password string, trace bool) *OpenConnectCtx {
	client := NewHttpClient(server)
	exit := make(chan os.Signal)

	// register exit signals
	signal.Notify(exit, os.Kill, os.Interrupt)

	return &OpenConnectCtx{
		client:          client,
		cookieFoundChan: make(chan string),
		exitChan:        exit,
		targetUrl:       getActualUrl(client, server),
		username:        username,
		password:        password,
		trace:           trace,
	}
}

func (oc *OpenConnectCtx) Run() error {
	samlAuth, err := oc.AuthenticationInit()
	if err != nil {
		log.Println("Could not start authentication process...")
		return err
	}

	tasks, err := oc.startBrowser(samlAuth)
	if err != nil {
		log.Println("Could not start browser properly...")
		return err
	}

	// close browser at the end - no matter what happens
	defer oc.closeBrowser()

	// handle exit signal
	log.Println("Starting goroutine to handle exit signals")
	go oc.handleExit()

	log.Println("Starting goroutine to search for cookie", samlAuth.Auth.SsoV2TokenCookieName)
	go oc.browserCookieFinder(samlAuth.Auth.SsoV2TokenCookieName, samlAuth.Auth.SsoV2ErrorCookieName)

	log.Println("Open browser and navigate to SSO login page : ", samlAuth.Auth.SsoV2Login)
	err = chromedp.Run(oc.browserCtx, tasks)
	if err != nil {
		return err
	}

	// consume cookie and connect to vpn
	return oc.startVpnOnLoginCookie(samlAuth)
}

func (oc *OpenConnectCtx) startBrowser(samlAuth *AuthenticationInitExpectedResponse) (chromedp.Tasks, error) {
	oc.browserCtx, oc.closeBrowser = createBrowserContext()
	tasks := oc.generateDefaultBrowserTasks(samlAuth)

	// setup listener to exit program when browser is closed
	chromedp.ListenTarget(oc.browserCtx, func(ev interface{}) {
		closeBrowserOnRenderProcessGone(ev, oc.exitChan)
	})

	return tasks, nil
}

func (oc *OpenConnectCtx) Post(url, contentType string, buffer *bytes.Buffer) (resp *http.Response, err error) {
	oc.tracef("POST %s (content-type: %s), body len : %d\n", url, contentType, buffer.Len())
	return oc.client.Post(url, contentType, buffer)
}

func (oc *OpenConnectCtx) buildArgs(version, cookie, servercert string) []string {
	args := []string{
		"openconnect",
		"--useragent=\"OpenConnect-SSO\"",
		fmt.Sprintf("--version-string=%s", version),
		fmt.Sprintf("--cookie=%s", cookie),
		fmt.Sprintf("--servercert=%s", servercert),
	}

	if oc.trace {
		args = append(args, "--verbose")
	}

	args = append(args, oc.targetUrl)
	return args
}

// startVpnOnLoginCookie waits to get a cookie from the authenticationCookies channel before confirming
// the authentication process (to get token/cert) and then starting openconnect
func (oc *OpenConnectCtx) startVpnOnLoginCookie(auth *AuthenticationInitExpectedResponse) error {
	log.Println("Starting cookie consumer to find session")
	for cookie := range oc.cookieFoundChan {
		token, cert, err := oc.AuthenticationConfirmation(auth, cookie)
		oc.closeBrowser() // close browser

		if err != nil {
			return err
		}

		oc.process = exec.Command("sudo",
			oc.buildArgs(VERSION, token, cert)...,
		)

		oc.process.Stdout = os.Stdout
		oc.process.Stderr = os.Stdout
		oc.process.Stdin = os.Stdin

		log.Println("Starting openconnect: ", oc.process.String())
		return oc.process.Run()
	}

	return nil
}

func (oc *OpenConnectCtx) tracef(format string, v ...any) {
	if oc.trace {
		log.Printf(format, v...)
	}
}
