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

	// Try to load cookies from cookies.json if COOKIES env var not set
	if cookies == "" {
		if cookieData, err := os.ReadFile("cookies.json"); err == nil {
			var parsedCookies []*http.Cookie
			if err := json.Unmarshal(cookieData, &parsedCookies); err == nil {
				// Normalize cookie domains to x.com
				for _, cookie := range parsedCookies {
					if cookie.Domain == "twitter.com" || cookie.Domain == "" {
						cookie.Domain = "x.com"
					}
				}
				testScraper.SetCookies(parsedCookies)

				// Check for auth_token presence (does not hit Twitter endpoints)
				if testScraper.IsLoggedIn() {
					fmt.Println("Successfully loaded cookies from cookies.json")
					return
				} else {
					fmt.Println("Warning: cookies.json does not contain auth_token")
				}
			} else {
				fmt.Printf("Warning: failed to parse cookies.json: %v\n", err)
			}
		} else {
			fmt.Printf("Warning: failed to read cookies.json: %v\n", err)
		}
	}

	skipAuthTest = true
	fmt.Println("No auth data provided (cookies.json, AUTH_TOKEN/CSRF_TOKEN, or COOKIES env var), skipping auth tests")
}

func newTestScraper(skip_auth bool) *twitterscraper.Scraper {
	s := twitterscraper.New()

	if proxy != "" && proxyRequired {
		err := s.SetProxy(proxy)
		if err != nil {
			panic(fmt.Sprintf("SetProxy() error = %v", err))
		}
	}

	if skip_auth == true || !skipAuthTest {
		s.ClearGuestToken()
		return s
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
