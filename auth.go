package twitterscraper

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	logoutURL = "https://api.x.com/1.1/account/logout.json"
	// Bearer token that requires x-client-transaction-id header
	bearerToken1      = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
	bearerToken2      = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
	appConsumerKey    = "3nVuSoBZnx6U4vzUxf5w"
	appConsumerSecret = "Bcs59EFbbsdF6Sl9Ng71smgStWEGwXXKSjYvPVt7qys"
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
	req, err := http.NewRequest("POST", logoutURL, nil)
	if err != nil {
		return err
	}
	err = s.RequestAPI(req, nil)
	if err != nil {
		return err
	}

	s.isLogged = false
	s.isOpenAccount = false
	s.guestToken = ""
	s.oAuthToken = ""
	s.oAuthSecret = ""
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
	expires := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	cookies := []*http.Cookie{{
		Name:       "auth_token",
		Value:      token.Token,
		Path:       "",
		Domain:     "x.com",
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
		Domain:     "x.com",
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

func (s *Scraper) sign(method string, ref *url.URL) string {
	m := make(map[string]string)
	m["oauth_consumer_key"] = appConsumerKey
	m["oauth_nonce"] = "0"
	m["oauth_signature_method"] = "HMAC-SHA1"
	m["oauth_timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	m["oauth_token"] = s.oAuthToken

	key := []byte(appConsumerSecret + "&" + s.oAuthSecret)
	h := hmac.New(sha1.New, key)

	query := ref.Query()
	for k, v := range m {
		query.Set(k, v)
	}

	req := []string{method, ref.Scheme + "://" + ref.Host + ref.Path, query.Encode()}
	var reqBuf bytes.Buffer
	for _, value := range req {
		if reqBuf.Len() > 0 {
			reqBuf.WriteByte('&')
		}
		reqBuf.WriteString(url.QueryEscape(value))
	}
	h.Write(reqBuf.Bytes())

	m["oauth_signature"] = base64.StdEncoding.EncodeToString(h.Sum(nil))

	var b bytes.Buffer
	for k, v := range m {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(url.QueryEscape(v))
	}

	return "OAuth " + b.String()
}
