package internal

import (
	"context"
	"github.com/chromedp/cdproto/inspector"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"os"
	"strings"
)

// opts are chrome options
var opts = append(chromedp.DefaultExecAllocatorOptions[:],
	chromedp.Flag("headless", false), // Set headless mode to false
	chromedp.Flag("disable-gpu", false),
)

// createBrowserContext creates new chromedp context and exec allocator
func createBrowserContext() (context.Context, context.CancelFunc) {
	ctx, _ := chromedp.NewContext(context.Background())
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)

	return chromedp.NewContext(allocCtx)
}

// closeBrowserOnRenderProcessGone sends an exit signal when the
// "Render process gone" is found (user manually closes the browser).
func closeBrowserOnRenderProcessGone(ev interface{}, exit chan os.Signal) {
	ins, ok := ev.(*inspector.EventDetached)
	if ok {
		if strings.Contains(ins.Reason.String(), "Render process gone.") {
			exit <- os.Kill
		}
	}
}

// browserCookieFinder setup's a chromedp listener in order to look
// through the cookies channel for a cookie that matches name
func (oc *OpenconnectCtx) browserCookieFinder(name string) {
	chromedp.ListenTarget(oc.browserCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSentExtraInfo:
			for _, cookie := range ev.AssociatedCookies {
				if cookie.Cookie.Name == name {
					oc.cookieFoundChan <- cookie.Cookie.Value
				}
			}
		}
	})
}

// generateDefaultBrowserTasks adds a task to inject username and/or password if the argument is present.
// also adds the initial Navigate command to open the browser on the right window
func (oc *OpenconnectCtx) generateDefaultBrowserTasks(samlAuth *AuthenticationInitExpectedResponse) chromedp.Tasks {
	var tasks chromedp.Tasks

	// create list of tasks to be executed by browser
	tasks = append(tasks, chromedp.Navigate(samlAuth.Auth.SsoV2Login))
	addAutofillTaskOnValue(&tasks, oc.password, "#passwordInput")
	addAutofillTaskOnValue(&tasks, oc.username, "#userNameInput")

	return tasks
}

// addAutofillTaskOnValue adds a task if value is not empty
func addAutofillTaskOnValue(actions *chromedp.Tasks, value, selector string) {
	if value != "" {
		*actions = append(
			*actions,
			chromedp.WaitVisible(selector, chromedp.ByID),
			chromedp.SendKeys(selector, value, chromedp.ByID),
		)
	}
}
