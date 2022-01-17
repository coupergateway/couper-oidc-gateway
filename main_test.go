package main_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func TestOpenIDConnectFlow(t *testing.T) {
	// create new chrome context (tab)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	url := "http://localhost:8080/"

	type item struct {
		url        string
		statusCode int64
		headers    network.Headers
		event      string
	}

	expectedEvents := []item{
		{url, 0, nil, reflect.TypeOf(&network.EventRequestWillBeSent{}).String()},
		{"", http.StatusForbidden, nil, reflect.TypeOf(&network.EventResponseReceivedExtraInfo{}).String()},
		{url, http.StatusForbidden, nil, reflect.TypeOf(&network.EventResponseReceived{}).String()},
		{"scriptInitiated: " + url + "_couper/oidc/start?url=%2F", 0, nil, reflect.TypeOf(&page.EventFrameRequestedNavigation{}).String()},
		{url + "_couper/oidc/start?url=%2F", 0, nil, reflect.TypeOf(&network.EventRequestWillBeSent{}).String()},
		{"", http.StatusSeeOther, network.Headers{
			"Cache-Control": "no-cache,no-store",
			"Location":      "http://localhost:8081/auth?client_id=foo&code_challenge=",
			"Set-Cookie":    "_couper_authvv="},
			reflect.TypeOf(&network.EventResponseReceivedExtraInfo{}).String()},
		{"http://localhost:8081/auth?client_id=foo&code_challenge=", 0, nil, reflect.TypeOf(&network.EventRequestWillBeSent{}).String()},
		{"", http.StatusSeeOther, network.Headers{
			"Cache-Control": "no-cache,no-store",
			"Location":      url + "_couper/oidc/callback?code=asdf&state=%2F",
		}, reflect.TypeOf(&network.EventResponseReceivedExtraInfo{}).String()},
		{url + "_couper/oidc/callback?code=asdf&state=%2F", 0, nil, reflect.TypeOf(&network.EventRequestWillBeSent{}).String()},
		{"", http.StatusSeeOther, network.Headers{
			"Cache-Control": "no-cache,no-store",
			"Set-Cookie":    "_couper_access_token=ey",
		}, reflect.TypeOf(&network.EventResponseReceivedExtraInfo{}).String()},
		{url, 0, nil, reflect.TypeOf(&network.EventRequestWillBeSent{}).String()},
		{"", http.StatusOK, nil, reflect.TypeOf(&network.EventResponseReceivedExtraInfo{}).String()},
		{url, http.StatusOK, nil, reflect.TypeOf(&network.EventResponseReceived{}).String()},
	}

	// register event listener
	var resultEvents []item
	rmu := sync.Mutex{}
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		rmu.Lock()
		defer rmu.Unlock()

		switch event := ev.(type) {
		case *network.EventResponseReceivedExtraInfo:
			resultEvents = append(resultEvents, item{
				statusCode: event.StatusCode,
				headers:    event.Headers,
				event:      reflect.TypeOf(event).String(),
			})
		case *network.EventResponseReceived:
			resultEvents = append(resultEvents, item{
				url:        event.Response.URL,
				statusCode: event.Response.Status,
				headers:    event.Response.Headers,
				event:      reflect.TypeOf(event).String(),
			})
		case *network.EventRequestWillBeSent:
			resultEvents = append(resultEvents, item{
				url:   event.Request.URL,
				event: reflect.TypeOf(event).String(),
			})
		case *page.EventFrameRequestedNavigation:
			resultEvents = append(resultEvents, item{
				url:   fmt.Sprintf("%s: %s", event.Reason, event.URL),
				event: reflect.TypeOf(event).String(),
			})
		}
	})

	// run chrome tab, clear cookies and navigate to url, verify cookie set
	if err := chromedp.Run(ctx,
		network.DeleteCookies("_couper_access_token").WithURL(url), // TOKEN_COOKIE_NAME
		network.DeleteCookies("_couper_authvv").WithDomain(url),    // VERIFIER_COOKIE_NAME
		chromedp.ActionFunc(func(c context.Context) error {
			cookies, err := network.GetCookies().
				WithUrls([]string{url}).Do(c)
			if err != nil {
				return err
			}

			if len(cookies) > 0 {
				for _, cookie := range cookies {
					t.Log(cookie.Name, cookie.Value)
				}
				return fmt.Errorf("expected cleared _couper cookies")
			}

			return nil
		}),
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(c context.Context) error {
			cookies, err := network.GetCookies().
				WithUrls([]string{url}).Do(c)
			if err != nil {
				return err
			}

			for _, cookie := range cookies {
				if cookie.Name == "_couper_access_token" {
					r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(cookie.Value))
					tokenBytes, _ := ioutil.ReadAll(r)
					token := map[string]interface{}{}
					return json.Unmarshal(tokenBytes, &token)
				}
			}

			return fmt.Errorf("expected cookie with couper access token")
		}),
	); err != nil {
		t.Fatal(err)
	}

	rmu.Lock()
	defer rmu.Unlock()

	// just differ first events
	for i, e := range expectedEvents {
		r := resultEvents[i]

		headers := true

		if e.headers != nil {
			for k, v := range e.headers {
				if rv, exist := r.headers[k]; !exist || !strings.HasPrefix(rv.(string), v.(string)) {
					headers = false
					break
				}
			}
		}

		if r.statusCode != e.statusCode || r.event != e.event ||
			!strings.HasPrefix(r.url, e.url) || !headers {
			t.Fatalf("event #%02d:\nwant: %#v\ngot:  %#v\n", i+1, e, resultEvents[i])
		}

		if r.url != "" { // just logging for ci
			if r.statusCode != 0 { // response events
				t.Logf("event: %q, url: %q, status: %d", r.event, r.url, r.statusCode)
			} else {
				t.Logf("event: %q, url: %q", r.event, r.url)
			}
		}
	}

}
