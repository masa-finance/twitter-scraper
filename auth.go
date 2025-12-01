package twitterscraper

import (
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

const (
	// Bearer token for X.com API - requires x-client-transaction-id header for authenticated requests
	bearerToken2 = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
)

// IsLoggedIn checks if scraper has valid auth cookies.
// This does NOT hit Twitter endpoints - it only checks for the presence of auth cookies.
func (s *Scraper) IsLoggedIn() bool {
	for _, cookie := range s.client.Jar.Cookies(twURL) {
		if cookie.Name == "auth_token" && cookie.Value != "" {
			s.isLogged = true
			return true
		}
	}
	s.isLogged = false
	return false
}

// Logout resets the session
func (s *Scraper) Logout() error {
	s.isLogged = false
	s.guestToken = ""
	s.client.Jar, _ = cookiejar.New(nil)
	s.setBearerToken(bearerToken)
	return nil
}

// GetCookies returns the current session cookies
func (s *Scraper) GetCookies() []*http.Cookie {
	var cookies []*http.Cookie
	for _, cookie := range s.client.Jar.Cookies(twURL) {
		if strings.Contains(cookie.Name, "guest") {
			continue
		}
		cookie.Domain = twURL.Host
		cookies = append(cookies, cookie)
	}
	return cookies
}

// SetCookies sets cookies for authentication
func (s *Scraper) SetCookies(cookies []*http.Cookie) {
	s.client.Jar.SetCookies(twURL, cookies)
}

// ClearCookies clears all cookies
func (s *Scraper) ClearCookies() {
	s.client.Jar, _ = cookiejar.New(nil)
}

// AuthToken represents auth_token and ct0 cookies for authentication
type AuthToken struct {
	Token     string
	CSRFToken string
}

// SetAuthToken sets authentication using auth_token and ct0 cookies
func (s *Scraper) SetAuthToken(token AuthToken) {
	expires := time.Date(2030, time.January, 1, 0, 0, 0, 0, time.UTC)
	cookies := []*http.Cookie{
		{
			Name:    "auth_token",
			Value:   token.Token,
			Domain:  "x.com",
			Path:    "/",
			Expires: expires,
			Secure:  true,
		},
		{
			Name:    "ct0",
			Value:   token.CSRFToken,
			Domain:  "x.com",
			Path:    "/",
			Expires: expires,
			Secure:  true,
		},
	}
	s.SetCookies(cookies)
}
