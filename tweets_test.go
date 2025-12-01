package twitterscraper_test

import (
	"context"
	"testing"
)

func TestGetTweets(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	count := 0
	maxTweetsNbr := 5
	dupcheck := make(map[string]bool)

	for tweet := range testScraper.GetTweets(context.Background(), "X", maxTweetsNbr) {
		if tweet.Error != nil {
			t.Error(tweet.Error)
			continue
		}
		count++
		if tweet.ID == "" {
			t.Error("Expected tweet ID is empty")
		} else {
			if dupcheck[tweet.ID] {
				t.Errorf("Detect duplicated tweet ID: %s", tweet.ID)
			} else {
				dupcheck[tweet.ID] = true
			}
		}
		if tweet.UserID == "" {
			t.Error("Expected tweet UserID is empty")
		}
		if tweet.PermanentURL == "" {
			t.Error("Expected tweet PermanentURL is empty")
		}
	}

	if count == 0 {
		t.Error("Expected at least some tweets")
	}
}

func TestGetTweet(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	// Use a well-known tweet that should exist
	tweet, err := testScraper.GetTweet("1697304622749086011")
	if err != nil {
		t.Fatal(err)
	}

	if tweet.ID == "" {
		t.Error("Expected tweet ID to be non-empty")
	}
	if tweet.Text == "" {
		t.Error("Expected tweet Text to be non-empty")
	}
}

func TestGetTweetsAndReplies(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	// Use X's user ID
	tweets, _, err := testScraper.FetchTweetsAndRepliesByUserID("783214", 5, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(tweets) == 0 {
		t.Error("Expected at least some tweets")
	}
}
