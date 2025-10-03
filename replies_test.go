package twitterscraper_test

import (
	"testing"
)

func TestGetReplies(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	tweetId := "1697304622749086011"

	tweets, cursors, err := testScraper.GetTweetReplies(tweetId, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(tweets) < 2 {
		t.Fatal("Less than 2 tweets returned")
	}

	if len(cursors) < 1 {
		t.Fatal("No cursors returned")
	}
}
