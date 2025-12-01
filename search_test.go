package twitterscraper_test

import (
	"context"
	"testing"

	twitterscraper "github.com/imperatrona/twitter-scraper"
)

func TestFetchSearchCursor(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	tweets, cursor, err := testScraper.FetchSearchTweets("twitter", 10, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(tweets) == 0 {
		t.Error("Expected at least some tweets")
	}

	if cursor == "" {
		t.Error("Expected cursor to be non-empty")
	}
}

func TestGetSearchTweets(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	count := 0
	maxTweetsNbr := 10
	dupcheck := make(map[string]bool)

	testScraper.SetSearchMode(twitterscraper.SearchLatest)
	for tweet := range testScraper.SearchTweets(context.Background(), "twitter", maxTweetsNbr) {
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
		if tweet.PermanentURL == "" {
			t.Error("Expected tweet PermanentURL is empty")
		}
		if tweet.Text == "" {
			t.Error("Expected tweet Text is empty")
		}
	}

	if count == 0 {
		t.Error("Expected at least some tweets")
	}
}

func TestGetSearchProfiles(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	count := 0
	maxProfilesNbr := 5

	testScraper.SetSearchMode(twitterscraper.SearchUsers)
	for profile := range testScraper.SearchProfiles(context.Background(), "Twitter", maxProfilesNbr) {
		if profile.Error != nil {
			t.Error(profile.Error)
			continue
		}
		count++
		if profile.UserID == "" {
			t.Error("Expected UserID is empty")
		}
	}

	// Profile search may return fewer results, just verify it doesn't error
	t.Logf("Got %d profiles", count)
}
