package twitterscraper

import (
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

const (
	logoutURL = "https://api.twitter.com/1.1/account/logout.json"
)

type (
	verifyCredentials struct {
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}
)

// IsLoggedIn check if scraper logged in
func (s *Scraper) IsLoggedIn() bool {
	req, err := http.NewRequest("GET", "https://api.twitter.com/1.1/account/verify_credentials.json", nil)
	if err != nil {
		return false
	}
	var verify verifyCredentials
	err = s.RequestAPI(req, &verify)
	if err != nil || verify.Errors != nil {
		s.isLogged = false
		return false
	}
	s.isLogged = true
	return true
}

// Logout is reset session
func (s *Scraper) Logout() error {
	req, err := http.NewRequest("POST", logoutURL, nil)
	if err != nil {
		return err
	}
	err = s.RequestAPI(req, nil)
	if err != nil {
		return err
	}

	s.isLogged = false
	s.client.Jar, _ = cookiejar.New(nil)
	return nil
}

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

func (s *Scraper) SetCookies(cookies []*http.Cookie) {
	s.client.Jar.SetCookies(twURL, cookies)
}

func (s *Scraper) ClearCookies() {
	s.client.Jar, _ = cookiejar.New(nil)
}

// Use auth_token cookie as Token and ct0 cookie as CSRFToken
type AuthToken struct {
	Token     string
	CSRFToken string
}

// Auth using auth_token and ct0 cookies
func (s *Scraper) SetAuthToken(token AuthToken) {
	expires := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	cookies := []*http.Cookie{{
		Name:       "auth_token",
		Value:      token.Token,
		Path:       "",
		Domain:     "twitter.com",
		Expires:    expires,
		RawExpires: "",
		MaxAge:     0,
		Secure:     false,
		HttpOnly:   false,
		SameSite:   0,
		Raw:        "",
		Unparsed:   nil,
	}, {
		Name:       "ct0",
		Value:      token.CSRFToken,
		Path:       "",
		Domain:     "twitter.com",
		Expires:    expires,
		RawExpires: "",
		MaxAge:     0,
		Secure:     false,
		HttpOnly:   false,
		SameSite:   0,
		Raw:        "",
		Unparsed:   nil,
	}}

	s.SetCookies(cookies)
}
