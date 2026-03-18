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
				RestID       string     `json:"rest_id"`
				Legacy       legacyUser `json:"legacy"`
				CoreUserInfo struct {
					CreatedAt  string `json:"created_at"`
					Name       string `json:"name"`
					ScreenName string `json:"screen_name"`
				} `json:"core"`
				Message        string `json:"message"`
				IsBlueVerified bool   `json:"is_blue_verified"`
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
	// Hash + payload from live x.com UserByScreenName (Mar 2026).
	req, err := http.NewRequest("GET", "https://x.com/i/api/graphql/IGgvgiOx4QZndDHuD3x9TQ/UserByScreenName", nil)
	if err != nil {
		return Profile{}, err
	}

	variables := map[string]interface{}{
		"screen_name":           username,
		"withGrokTranslatedBio": true,
	}

	features := map[string]interface{}{
		"hidden_profile_subscriptions_enabled":                              true,
		"profile_label_improvements_pcf_label_in_post_enabled":              true,
		"responsive_web_profile_redirect_enabled":                           false,
		"rweb_tipjar_consumption_enabled":                                   false,
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
	fieldToggles := map[string]interface{}{
		"withPayments":            false,
		"withAuxiliaryUserLabels": true,
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	query.Set("fieldToggles", mapToJSONString(fieldToggles))
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
	if jsn.Data.User.Result.Legacy.ScreenName == "" && jsn.Data.User.Result.CoreUserInfo.ScreenName != "" {
		jsn.Data.User.Result.Legacy.ScreenName = jsn.Data.User.Result.CoreUserInfo.ScreenName
	}
	if jsn.Data.User.Result.Legacy.Name == "" && jsn.Data.User.Result.CoreUserInfo.Name != "" {
		jsn.Data.User.Result.Legacy.Name = jsn.Data.User.Result.CoreUserInfo.Name
	}
	if jsn.Data.User.Result.Legacy.CreatedAt == "" && jsn.Data.User.Result.CoreUserInfo.CreatedAt != "" {
		jsn.Data.User.Result.Legacy.CreatedAt = jsn.Data.User.Result.CoreUserInfo.CreatedAt
	}

	if jsn.Data.User.Result.Legacy.ScreenName == "" {
		return Profile{}, fmt.Errorf("either @%s does not exist or is private", username)
	}

	profile := parseProfile(jsn.Data.User.Result.Legacy)
	profile.IsBlueVerified = jsn.Data.User.Result.IsBlueVerified
	return profile, nil
}

func (s *Scraper) GetProfileByID(userID string) (Profile, error) {
	username, err := s.resolveScreenNameByUserID(userID)
	if err != nil {
		return Profile{}, err
	}

	profile, err := s.GetProfile(username)
	if err != nil {
		return Profile{}, err
	}

	// Keep the caller-supplied ID authoritative even if X resolves via username.
	profile.UserID = userID
	return profile, nil
}

func (s *Scraper) resolveScreenNameByUserID(userID string) (string, error) {
	// Current x.com profile pages resolve the username client-side after /i/user/<id>.
	// Resolve by reading the username from the user's timeline APIs first.
	if tweets, _, err := s.FetchTweetsByUserID(userID, 1, ""); err == nil {
		for _, tweet := range tweets {
			if tweet != nil && tweet.Username != "" {
				return tweet.Username, nil
			}
		}
	}

	if tweets, _, err := s.FetchTweetsAndRepliesByUserID(userID, 1, ""); err == nil {
		for _, tweet := range tweets {
			if tweet != nil && tweet.Username != "" {
				return tweet.Username, nil
			}
		}
	}

	return "", fmt.Errorf("could not resolve user id %s", userID)
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
