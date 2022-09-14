package main_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func TestMain(m *testing.M) {
	log.Print("docker-compose pull ...")
	cmd := exec.Command("docker-compose", "-f", "test/docker-compose.yml", "pull")
	cmd.Stdout = io.Discard
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("docker-compose up ...")
	cmd = exec.Command("docker-compose", "-f", "test/docker-compose.yml", "up", "-d")
	cmd.Stdout = io.Discard
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	log.Print("docker-compose down ...")
	cmd = exec.Command("docker-compose", "-f", "test/docker-compose.yml", "down")
	cmd.Stdout = io.Discard
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func TestOpenIDConnectFlow(t *testing.T) {
	// create new remote chrome context (tab)
	allocCtx, rcancel := chromedp.NewRemoteAllocator(context.Background(), "ws://127.0.0.1:9222/")
	defer rcancel()

	chromeLog := &bytes.Buffer{}
	chromeLogWriter := bufio.NewWriter(chromeLog)
	log.SetOutput(chromeLogWriter)
	defer log.SetOutput(os.Stdout)

	defer func() {
		if t.Failed() {
			names := []string{"test_couper-test-openid-provider_1", "test_couper-oidc-gateway_1"}
			for _, name := range names {
				logCmd := exec.Command("docker", "logs", name)
				logCmd.Stdout = os.Stdout
				t.Log("\n" + name + " log output:\n")
				_ = logCmd.Run()
			}
			t.Log("\nChrome log output:\n")
			println(chromeLog.String())
		}
	}()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithDebugf(log.Printf), chromedp.WithLogf(log.Printf))
	defer cancel()

	url := "http://couper:8080/"

	expectedEvents := []testEvent{
		{url: url + "en/docs/?foo=oidc-test", statusCode: http.StatusForbidden, headers: nil},
		{url: url + "_couper/oidc/start?url=%2Fen%2Fdocs%2F%3Ffoo%3Doidc-test", statusCode: http.StatusSeeOther,
			headers: network.Headers{
				"Cache-Control": "no-cache,no-store",
				"Location":      "http://testop:8080/auth?client_id=foo&code_challenge=",
				"Set-Cookie":    "_couper_authvv="},
		},
		{url: "http://testop:8080/auth?client_id=foo&code_challenge=", statusCode: http.StatusSeeOther,
			headers: network.Headers{
				"Cache-Control": "no-cache,no-store",
				"Location":      url + "_couper/oidc/callback?code=asdf&state=%2Fen%2Fdocs%2F%3Ffoo%3Doidc-test",
			},
		},
		{url: url + "_couper/oidc/callback?code=asdf&state=%2Fen%2Fdocs%2F%3Ffoo%3Doidc-test", statusCode: http.StatusSeeOther,
			headers: network.Headers{
				"Cache-Control": "no-cache,no-store",
				"Set-Cookie":    "_couper_access_token=ey",
			}},
		{url: url + "en/docs/?foo=oidc-test", statusCode: http.StatusOK, headers: nil},
	}

	// register event listener
	var testEvents []*testEvent
	rmu := sync.Mutex{}

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch event := ev.(type) {
		case *fetch.EventRequestPaused:
			// catch outgoing reqs and register them by ID and order
			go func(c context.Context, e *fetch.EventRequestPaused) {

				rmu.Lock()
				testEvents = append(testEvents, &testEvent{
					url: e.Request.URL,
				})
				rmu.Unlock()

				cc := chromedp.FromContext(c)
				ec := cdp.WithExecutor(ctx, cc.Target)

				if err := fetch.ContinueRequest(e.RequestID).Do(ec); err != nil && err != context.Canceled {
					t.Error(err)
				}
			}(ctx, event)
		case *network.EventResponseReceivedExtraInfo:
			rmu.Lock()
			defer rmu.Unlock()

			i := len(testEvents) - 1

			testEvents[i].statusCode = event.StatusCode
			testEvents[i].headers = event.Headers
		}
	})

	const tokenName = "_couper_access_token"
	const verifierName = "_couper_authvv"

	// run chrome tab, clear cookies and navigate to url, verify cookie set
	if err := chromedp.Run(ctx,
		network.Enable(),
		fetch.Enable(),
		network.DeleteCookies(tokenName).WithURL(url),                            // TOKEN_COOKIE_NAME
		network.DeleteCookies(verifierName).WithURL(url+"_couper/oidc/callback"), // VERIFIER_COOKIE_NAME
		chromedp.ActionFunc(func(c context.Context) error {
			cookies, err := network.GetAllCookies().Do(c)
			if err != nil {
				return err
			}

			if len(cookies) > 0 {
				for _, cookie := range cookies {
					if cookie.Name == tokenName || cookie.Name == verifierName {
						t.Log(cookie.Name, cookie.Value)
						return fmt.Errorf("expected cleared _couper cookies")
					}
				}
			}

			return nil
		}),
		chromedp.Navigate(url+"en/docs/?foo=oidc-test"),
		chromedp.ActionFunc(func(c context.Context) error {
			cookies, err := network.GetAllCookies().Do(c)
			if err != nil {
				return err
			}

			var acTokenSeen, verifierSeen bool

			for _, cookie := range cookies {
				if cookie.Name == tokenName {
					r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(cookie.Value))
					tokenBytes, _ := ioutil.ReadAll(r)
					token := map[string]interface{}{}
					if err = json.Unmarshal(tokenBytes, &token); err != nil { // valid json?
						return err
					}
					acTokenSeen = true
				}

				if cookie.Name == verifierName {
					verifierSeen = true
				}
			}

			if !acTokenSeen {
				return fmt.Errorf("expected cookie: %q", tokenName)
			}

			if verifierSeen {
				return fmt.Errorf("unexpected cookie: %q", verifierName)
			}

			return nil
		}),
	); err != nil {
		t.Fatal(err)
	}

	rmu.Lock()
	defer rmu.Unlock()

	// just differ first events
	for i, e := range expectedEvents {
		r := testEvents[i]

		headers := true

		if e.headers != nil {
			for k, v := range e.headers {
				if rv, exist := r.headers[k]; !exist || !strings.HasPrefix(rv.(string), v.(string)) {
					headers = false
					break
				}
			}
		}

		if r.statusCode != e.statusCode ||
			!strings.HasPrefix(r.url, e.url) ||
			!headers {
			t.Fatalf("event #%02d:\nwant: %#v\ngot:  %#v\n", i+1, e, *r)
		}

		t.Logf("url: %q, status: %d", r.url, r.statusCode)
	}
}

type testEvent struct {
	url        string
	statusCode int64
	headers    network.Headers
}
