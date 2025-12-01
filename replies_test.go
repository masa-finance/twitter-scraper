package twitterscraper_test

import (
	"testing"
)

func TestGetReplies(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	// Use a tweet that likely has replies
	tweetId := "1697304622749086011"

	tweets, _, err := testScraper.GetTweetReplies(tweetId, "")
	if err != nil {
		t.Fatal(err)
	}

	// Just verify we got some data back - the original tweet should at least be present
	if len(tweets) < 1 {
		t.Error("Expected at least 1 tweet returned")
	}
}
