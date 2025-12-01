package twitterscraper_test

import (
	"testing"
)

func TestGetProfile(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	// Test with a well-known account that should always exist
	profile, err := testScraper.GetProfile("X")
	if err != nil {
		t.Fatal(err)
	}

	if profile.Username == "" {
		t.Error("Expected Username to be non-empty")
	}
	if profile.UserID == "" {
		t.Error("Expected UserID to be non-empty")
	}
	if profile.Name == "" {
		t.Error("Expected Name to be non-empty")
	}
}

func TestGetProfileByID(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	// X's user ID
	profile, err := testScraper.GetProfileByID("783214")
	if err != nil {
		t.Fatal(err)
	}

	if profile.Username == "" {
		t.Error("Expected Username to be non-empty")
	}
}

func TestGetProfileErrorNotFound(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	_, err := testScraper.GetProfile("sample3123131nonexistent")
	if err == nil {
		t.Error("Expected Error for non-existent user, got success")
	}
}

func TestGetUserIDByScreenName(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	userID, err := testScraper.GetUserIDByScreenName("X")
	if err != nil {
		t.Fatal(err)
	}
	if userID == "" {
		t.Error("Expected non-empty user ID")
	}
}
