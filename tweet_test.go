package twitterscraper_test

import (
	"testing"
)

func TestGetTweetRetweeters(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}
	tweetId := "1792634158977568997"

	retweeters, _, err := testScraper.GetTweetRetweeters(tweetId, 20, "")
	if err != nil {
		t.Error(err)
	}

	if len(retweeters) == 0 {
		t.Error("0 tweet retweeters")
	}
}
