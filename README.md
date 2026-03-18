# Twitter Scraper

[![Go Reference](https://pkg.go.dev/badge/github.com/imperatrona/twitter-scraper.svg)](https://pkg.go.dev/github.com/imperatrona/twitter-scraper)

Production-focused scraper for X web endpoints.

This fork is maintained around the endpoint surface the codebase actually implements and that we actively test. It is intended for authenticated scraping with real browser cookies.

## Installation

```shell
go get -u github.com/imperatrona/twitter-scraper
```

## What This Supports

The production-maintained surface documented in this README is:

- Authentication via `SetCookies` or `SetAuthToken`
- `GetTweet`
- `GetTweetReplies`
- `GetTweetRetweeters`
- `GetTweets`
- `GetTweetsAndReplies`
- `FetchTweets`
- `FetchTweetsAndReplies`
- `FetchTweetsByUserID`
- `FetchTweetsAndRepliesByUserID`
- `GetHomeTweets`
- `FetchHomeTweets`
- `SearchTweets`
- `FetchSearchTweets`
- `SearchProfiles`
- `FetchSearchProfiles`
- `GetProfile`
- `GetProfileByID`
- `GetUserIDByScreenName`
- `GetTrends`
- `GetCookies`
- `Logout`
- Proxy support via `SetProxy`
- User-Agent overrides via `SetUserAgent`

The older README documented a much larger feature set; unsupported or unmaintained sections have been removed here so production docs match the code we actually keep current.

## Quick Start

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	twitterscraper "github.com/imperatrona/twitter-scraper"
)

func main() {
	var cookies []*http.Cookie

	f, err := os.Open("cookies.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cookies); err != nil {
		panic(err)
	}

	scraper := twitterscraper.New()
	scraper.SetCookies(cookies)

	if !scraper.IsLoggedIn() {
		panic("missing auth cookies")
	}

	for tweet := range scraper.SearchTweets(context.Background(), "subnet signal", 20) {
		if tweet.Error != nil {
			panic(tweet.Error)
		}
		fmt.Println(tweet.ID, tweet.Text)
	}
}
```

## Authentication

Most useful endpoints in this library require authentication.

The recommended approach is to load browser cookies containing at least:

- `auth_token`
- `ct0`

### Using cookies

```go
var cookies []*http.Cookie

f, err := os.Open("cookies.json")
if err != nil {
	panic(err)
}
defer f.Close()

if err := json.NewDecoder(f).Decode(&cookies); err != nil {
	panic(err)
}

scraper.SetCookies(cookies)
if !scraper.IsLoggedIn() {
	panic("invalid cookies")
}
```

### Using `AuthToken`

```go
scraper.SetAuthToken(twitterscraper.AuthToken{
	Token:     "auth_token",
	CSRFToken: "ct0",
})

if !scraper.IsLoggedIn() {
	panic("invalid auth token")
}
```

### Saving cookies

```go
cookies := scraper.GetCookies()

data, err := json.Marshal(cookies)
if err != nil {
	panic(err)
}

if err := os.WriteFile("cookies.json", data, 0o600); err != nil {
	panic(err)
}
```

### Logging out

```go
if err := scraper.Logout(); err != nil {
	panic(err)
}
```

## Channel-Based Methods

Some methods stream results over a channel and paginate internally. Always check `Error` before using the item.

```go
for tweet := range scraper.GetTweets(context.Background(), "x", 50) {
	if tweet.Error != nil {
		panic(tweet.Error)
	}
	fmt.Println(tweet.Text)
}
```

These channel helpers use the same underlying endpoints and rate limits as their corresponding `Fetch*` methods.

## Supported Methods

### Get a tweet

```go
tweet, err := scraper.GetTweet("1328684389388185600")
```

### Get tweet replies

```go
replies, cursors, err := scraper.GetTweetReplies("1328684389388185600", "")
```

`GetTweetReplies` can return multiple cursors, one per thread branch.

### Get tweet retweeters

```go
profiles, cursor, err := scraper.GetTweetRetweeters("1328684389388185600", 20, "")
```

### Get user tweets

```go
for tweet := range scraper.GetTweets(context.Background(), "taylorswift13", 50) {
	if tweet.Error != nil {
		panic(tweet.Error)
	}
	fmt.Println(tweet.Text)
}
```

### Fetch user tweets

```go
tweets, cursor, err := scraper.FetchTweets("taylorswift13", 20, "")
```

### Get user tweets and replies

```go
for tweet := range scraper.GetTweetsAndReplies(context.Background(), "taylorswift13", 50) {
	if tweet.Error != nil {
		panic(tweet.Error)
	}
	fmt.Println(tweet.Text)
}
```

### Fetch tweets by numeric user ID

```go
tweets, cursor, err := scraper.FetchTweetsByUserID("783214", 20, "")
```

### Fetch tweets and replies by numeric user ID

```go
tweets, cursor, err := scraper.FetchTweetsAndRepliesByUserID("783214", 20, "")
```

### Get home timeline

Requires authentication.

```go
for tweet := range scraper.GetHomeTweets(context.Background(), 50) {
	if tweet.Error != nil {
		panic(tweet.Error)
	}
	fmt.Println(tweet.Text)
}
```

### Fetch home timeline page

Requires authentication.

```go
tweets, cursor, err := scraper.FetchHomeTweets(20, "")
```

### Search tweets

Requires authentication.

```go
for tweet := range scraper.SearchTweets(context.Background(), "twitter scraper -filter:retweets", 50) {
	if tweet.Error != nil {
		panic(tweet.Error)
	}
	fmt.Println(tweet.Text)
}
```

Use `SetSearchMode` to switch modes. Supported modes are:

- `SearchTop`
- `SearchLatest`
- `SearchPhotos`
- `SearchVideos`
- `SearchUsers`

```go
scraper.SetSearchMode(twitterscraper.SearchLatest)
```

### Fetch tweet search page

```go
tweets, cursor, err := scraper.FetchSearchTweets("subnet signal", 20, "")
```

### Search profiles

Requires authentication.

```go
for profile := range scraper.SearchProfiles(context.Background(), "Twitter", 20) {
	if profile.Error != nil {
		panic(profile.Error)
	}
	fmt.Println(profile.Username, profile.Name)
}
```

### Fetch profile search page

```go
profiles, cursor, err := scraper.FetchSearchProfiles("Twitter", 20, "")
```

### Get profile by username

```go
profile, err := scraper.GetProfile("taylorswift13")
```

### Get profile by numeric user ID

```go
profile, err := scraper.GetProfileByID("783214")
```

### Resolve numeric user ID from screen name

```go
userID, err := scraper.GetUserIDByScreenName("twitter")
```

### Get trends

```go
trends, err := scraper.GetTrends()
```

## Connection Settings

### User-Agent

```go
scraper.SetUserAgent("Mozilla/5.0 ...")
agent := scraper.GetUserAgent()
_ = agent
```

### HTTP proxy

```go
err := scraper.SetProxy("http://localhost:3128")
```

### SOCKS5 proxy

```go
err := scraper.SetProxy("socks5://localhost:1080")
```

SOCKS5 with auth is also supported:

```go
err := scraper.SetProxy("socks5://user:pass@localhost:1080")
```

## Operational Notes

- X rotates GraphQL query IDs and feature payloads over time. This repo tracks the web client, not the public API.
- Authenticated GraphQL requests depend on `x-client-transaction-id`; the library generates that header automatically.
- Rate limits change without notice. Add pacing, retries, and account rotation in production.
- Prefer cookies from a real logged-in browser session over password automation.

## Testing

Authenticated tests need valid session credentials. Keep secrets out of git.

In practice, the easiest path is to provide a local `cookies.json` with a current session before running:

```shell
go test ./...
```
