package twitterscraper_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	twitterscraper "github.com/imperatrona/twitter-scraper"
)

var (
	proxy         = os.Getenv("PROXY")
	proxyRequired = os.Getenv("PROXY_REQUIRED") != ""
	authToken     = os.Getenv("AUTH_TOKEN")
	csrfToken     = os.Getenv("CSRF_TOKEN")
	cookies       = os.Getenv("COOKIES")
	skipAuthTest  = os.Getenv("SKIP_AUTH_TEST") != ""
	testScraper   = newTestScraper(false)
)

func init() {
	if skipAuthTest {
		return
	}

	if authToken != "" && csrfToken != "" {
		testScraper.SetAuthToken(twitterscraper.AuthToken{Token: authToken, CSRFToken: csrfToken})
		if !testScraper.IsLoggedIn() {
			panic("Invalid AuthToken")
		}
		return
	}

	if cookies != "" {
		var parsedCookies []*http.Cookie
		json.NewDecoder(strings.NewReader(cookies)).Decode(&parsedCookies)
		testScraper.SetCookies(parsedCookies)
		if !testScraper.IsLoggedIn() {
			panic("Invalid Cookies")
		}
		return
	}

	skipAuthTest = true
	fmt.Println("No cookie authentication data provided, skipping all tests that require auth")
}

func newTestScraper(skip_auth bool) *twitterscraper.Scraper {
	s := twitterscraper.New()

	if proxy != "" && proxyRequired {
		err := s.SetProxy(proxy)
		if err != nil {
			panic(fmt.Sprintf("SetProxy() error = %v", err))
		}
	}

	return s
}

func TestLoginToken(t *testing.T) {
	if skipAuthTest || authToken == "" || csrfToken == "" {
		t.Skip("Skipping test due to environment variable")
	}

	scraper := newTestScraper(true)

	scraper.SetAuthToken(twitterscraper.AuthToken{Token: authToken, CSRFToken: csrfToken})
	if !scraper.IsLoggedIn() {
		t.Error("Expected IsLoggedIn() = true")
	}
}

func TestLoginCookie(t *testing.T) {
	if skipAuthTest || cookies == "" {
		t.Skip("Skipping test due to environment variable")
	}

	scraper := newTestScraper(true)

	var c []*http.Cookie

	json.NewDecoder(strings.NewReader(cookies)).Decode(&c)

	scraper.SetCookies(c)
	if !scraper.IsLoggedIn() {
		t.Error("Expected IsLoggedIn() = true")
	}
}
