package twitterscraper

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Global cache for user IDs
var cacheIDs sync.Map

// Profile of twitter user.
type Profile struct {
	Avatar               string
	Banner               string
	Biography            string
	Birthday             string
	FollowersCount       int
	FollowingCount       int
	FriendsCount         int
	IsPrivate            bool
	IsVerified           bool
	IsBlueVerified       bool
	Joined               *time.Time
	LikesCount           int
	ListedCount          int
	Location             string
	Name                 string
	PinnedTweetIDs       []string
	TweetsCount          int
	URL                  string
	UserID               string
	Username             string
	Website              string
	Sensitive            bool
	Following            bool
	FollowedBy           bool
	MediaCount           int
	FastFollowersCount   int
	NormalFollowersCount int
	ProfileImageShape    string
	HasGraduatedAccess   bool
	CanHighlightTweets   bool
}

type user struct {
	Data struct {
		User struct {
			Result struct {
				RestID         string     `json:"rest_id"`
				Legacy         legacyUser `json:"legacy"`
				Message        string     `json:"message"`
				IsBlueVerified bool       `json:"is_blue_verified"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// GetProfile return parsed user profile.
func (s *Scraper) GetProfile(username string) (Profile, error) {
	var jsn user
	req, err := http.NewRequest("GET", "https://api.x.com/graphql/NimuplG1OB7Fd2btCLdBOw/UserByScreenName", nil)
	if err != nil {
		return Profile{}, err
	}

	variables := map[string]interface{}{
		"screen_name":              username,
		"withSafetyModeUserFields": true,
	}

	features := map[string]interface{}{
		"hidden_profile_subscriptions_enabled":                              true,
		"rweb_tipjar_consumption_enabled":                                   true,
		"responsive_web_graphql_exclude_directive_enabled":                  true,
		"verified_phone_label_enabled":                                      false,
		"subscriptions_verification_info_is_identity_verified_enabled":      true,
		"subscriptions_verification_info_verified_since_enabled":            true,
		"highlights_tweets_tab_ui_enabled":                                  true,
		"responsive_web_twitter_article_notes_tab_enabled":                  true,
		"subscriptions_feature_can_gift_premium":                            true,
		"creator_subscriptions_tweet_preview_api_enabled":                   true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled": false,
		"responsive_web_graphql_timeline_navigation_enabled":                true,
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	req.URL.RawQuery = query.Encode()

	err = s.RequestAPI(req, &jsn)
	if err != nil {
		return Profile{}, err
	}

	if len(jsn.Errors) > 0 && jsn.Data.User.Result.RestID == "" {
		if strings.Contains(jsn.Errors[0].Message, "Missing LdapGroup(visibility-custom-suspension)") {
			return Profile{}, fmt.Errorf("user is suspended")
		}
		return Profile{}, fmt.Errorf("%s", jsn.Errors[0].Message)
	}

	if jsn.Data.User.Result.RestID == "" {
		if jsn.Data.User.Result.Message == "User is suspended" {
			return Profile{}, fmt.Errorf("user is suspended")
		}
		return Profile{}, fmt.Errorf("user not found")
	}
	jsn.Data.User.Result.Legacy.IDStr = jsn.Data.User.Result.RestID

	if jsn.Data.User.Result.Legacy.ScreenName == "" {
		return Profile{}, fmt.Errorf("either @%s does not exist or is private", username)
	}

	profile := parseProfile(jsn.Data.User.Result.Legacy)
	profile.IsBlueVerified = jsn.Data.User.Result.IsBlueVerified
	return profile, nil
}

func (s *Scraper) GetProfileByID(userID string) (Profile, error) {
	var jsn user
	req, err := http.NewRequest("GET", "https://api.x.com/graphql/tD8zKvQzwY3kdx5yz6YmOw/UserByRestId", nil)
	if err != nil {
		return Profile{}, err
	}

	variables := map[string]interface{}{
		"userId":                   userID,
		"withSafetyModeUserFields": true,
	}

	features := map[string]interface{}{
		"hidden_profile_subscriptions_enabled":                              true,
		"rweb_tipjar_consumption_enabled":                                   true,
		"responsive_web_graphql_exclude_directive_enabled":                  true,
		"verified_phone_label_enabled":                                      false,
		"highlights_tweets_tab_ui_enabled":                                  true,
		"responsive_web_twitter_article_notes_tab_enabled":                  true,
		"subscriptions_feature_can_gift_premium":                            true,
		"creator_subscriptions_tweet_preview_api_enabled":                   true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled": false,
		"responsive_web_graphql_timeline_navigation_enabled":                true,
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	req.URL.RawQuery = query.Encode()

	err = s.RequestAPI(req, &jsn)
	if err != nil {
		return Profile{}, err
	}

	if len(jsn.Errors) > 0 && jsn.Data.User.Result.RestID == "" {
		if strings.Contains(jsn.Errors[0].Message, "Missing LdapGroup(visibility-custom-suspension)") {
			return Profile{}, fmt.Errorf("user is suspended")
		}
		return Profile{}, fmt.Errorf("%s", jsn.Errors[0].Message)
	}

	if jsn.Data.User.Result.RestID == "" {
		if jsn.Data.User.Result.Message == "User is suspended" {
			return Profile{}, fmt.Errorf("user is suspended")
		}
		return Profile{}, fmt.Errorf("user not found")
	}
	jsn.Data.User.Result.Legacy.IDStr = jsn.Data.User.Result.RestID

	if jsn.Data.User.Result.Legacy.ScreenName == "" {
		return Profile{}, fmt.Errorf("either @%s does not exist or is private", userID)
	}

	profile := parseProfile(jsn.Data.User.Result.Legacy)
	profile.IsBlueVerified = jsn.Data.User.Result.IsBlueVerified
	return profile, nil
}

// GetUserIDByScreenName from API
func (s *Scraper) GetUserIDByScreenName(screenName string) (string, error) {
	id, ok := cacheIDs.Load(screenName)
	if ok {
		return id.(string), nil
	}

	profile, err := s.GetProfile(screenName)
	if err != nil {
		return "", err
	}

	cacheIDs.Store(screenName, profile.UserID)

	return profile.UserID, nil
}
