package twitterscraper

import (
	"context"
	"net/url"
	"strconv"
)

const searchURL = "https://api.x.com/graphql/bshMIjqDk8LTXTq4w91WKw/SearchTimeline"

type searchTimeline struct {
	Data struct {
		SearchByRawQuery struct {
			SearchTimeline struct {
				Timeline struct {
					Instructions []struct {
						Type    string  `json:"type"`
						Entries []entry `json:"entries"`
						Entry   entry   `json:"entry,omitempty"`
					} `json:"instructions"`
				} `json:"timeline"`
			} `json:"search_timeline"`
		} `json:"search_by_raw_query"`
	} `json:"data"`
}

func (timeline *searchTimeline) parseTweets() ([]*Tweet, string) {
	tweets := make([]*Tweet, 0)
	cursor := ""
	for _, instruction := range timeline.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions {
		if instruction.Type == "TimelineAddEntries" || instruction.Type == "TimelineReplaceEntry" {
			if instruction.Entry.Content.CursorType == "Bottom" {
				cursor = instruction.Entry.Content.Value
				continue
			}
			for _, entry := range instruction.Entries {
				if entry.Content.ItemContent.TweetDisplayType == "Tweet" {
					var legacy *legacyTweet = &entry.Content.ItemContent.TweetResults.Result.Legacy
					var user *legacyUser = &entry.Content.ItemContent.TweetResults.Result.Core.UserResults.Result.Legacy
					if entry.Content.ItemContent.TweetResults.Result.Typename == "TweetWithVisibilityResults" {
						legacy = &entry.Content.ItemContent.TweetResults.Result.Tweet.Legacy
						user = &entry.Content.ItemContent.TweetResults.Result.Tweet.Core.UserResults.Result.Legacy
					}
					if tweet := parseLegacyTweet(user, legacy); tweet != nil {
						var views = entry.Content.ItemContent.TweetResults.Result.Views.Count
						if entry.Content.ItemContent.TweetResults.Result.Typename == "TweetWithVisibilityResults" {
							views = entry.Content.ItemContent.TweetResults.Result.Tweet.Views.Count
						}
						if tweet.Views == 0 && views != "" {
							tweet.Views, _ = strconv.Atoi(views)
						}
						tweets = append(tweets, tweet)
					}
				} else if entry.Content.CursorType == "Bottom" {
					cursor = entry.Content.Value
				}
			}
		}
	}
	return tweets, cursor
}

func (timeline *searchTimeline) parseUsers() ([]*Profile, string) {
	profiles := make([]*Profile, 0)
	cursor := ""
	for _, instruction := range timeline.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions {
		if instruction.Type == "TimelineAddEntries" || instruction.Type == "TimelineReplaceEntry" {
			if instruction.Entry.Content.CursorType == "Bottom" {
				cursor = instruction.Entry.Content.Value
				continue
			}
			for _, entry := range instruction.Entries {
				if entry.Content.ItemContent.UserDisplayType == "User" {
					if profile := parseProfileV2(entry.Content.ItemContent.UserResults.Result); profile.Name != "" {
						if profile.UserID == "" {
							profile.UserID = entry.Content.ItemContent.UserResults.Result.RestID
						}
						profiles = append(profiles, &profile)
					}
				} else if entry.Content.CursorType == "Bottom" {
					cursor = entry.Content.Value
				}
			}
		}
	}
	return profiles, cursor
}

// SearchTweets returns channel with tweets for a given search query
func (s *Scraper) SearchTweets(ctx context.Context, query string, maxTweetsNbr int) <-chan *TweetResult {
	return getTweetTimeline(ctx, query, maxTweetsNbr, s.FetchSearchTweets)
}

// SearchProfiles returns channel with profiles for a given search query
func (s *Scraper) SearchProfiles(ctx context.Context, query string, maxProfilesNbr int) <-chan *ProfileResult {
	return getUserTimeline(ctx, query, maxProfilesNbr, s.FetchSearchProfiles)
}

// getSearchTimeline gets results for a given search query, via the Twitter frontend API
func (s *Scraper) getSearchTimeline(query string, maxNbr int, cursor string) (*searchTimeline, error) {
	if maxNbr > 50 {
		maxNbr = 50
	}

	req, err := s.newRequest("GET", searchURL)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"rawQuery":    query,
		"count":       maxNbr,
		"querySource": "typed_query",
		"product":     "Top",
	}

	features := map[string]interface{}{
		"rweb_video_screen_enabled":                                         true,
		"profile_label_improvements_pcf_label_in_post_enabled":              true,
		"responsive_web_profile_redirect_enabled":                           true,
		"rweb_tipjar_consumption_enabled":                                   true,
		"verified_phone_label_enabled":                                      false,
		"creator_subscriptions_tweet_preview_api_enabled":                   true,
		"responsive_web_graphql_timeline_navigation_enabled":                true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled": false,
		"premium_content_api_read_enabled":                                  false,
		"communities_web_enable_tweet_community_results_fetch":              true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                         true,
		"responsive_web_grok_analyze_button_fetch_trends_enabled":           false,
		"responsive_web_grok_analyze_post_followups_enabled":                false,
		"responsive_web_jetfuel_frame":                                      false,
		"responsive_web_grok_share_attachment_enabled":                      true,
		"articles_preview_enabled":                                          true,
		"responsive_web_edit_tweet_api_enabled":                             true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":        true,
		"view_counts_everywhere_api_enabled":                                true,
		"longform_notetweets_consumption_enabled":                           true,
		"responsive_web_twitter_article_tweet_consumption_enabled":          false,
		"tweet_awards_web_tipping_enabled":                                  false,
		"responsive_web_grok_show_grok_translated_post":                     true,
		"responsive_web_grok_analysis_button_from_backend":                  true,
		"creator_subscriptions_quote_tweet_preview_enabled":                 false,
		"freedom_of_speech_not_reach_fetch_enabled":                         true,
		"standardized_nudges_misinfo":                                       true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_grok_image_annotation_enabled":                            true,
		"responsive_web_grok_imagine_annotation_enabled":                          true,
		"responsive_web_grok_community_note_auto_translation_is_enabled":          true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	fieldToggles := map[string]interface{}{
		"withPayments":                true,
		"withAuxiliaryUserLabels":     false,
		"withArticleRichContentState": true,
		"withArticlePlainText":        false,
		"withGrokAnalyze":             false,
		"withDisallowedReplyControls": false,
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}
	switch s.searchMode {
	case SearchLatest:
		variables["product"] = "Latest"
	case SearchPhotos:
		variables["product"] = "Photos"
	case SearchVideos:
		variables["product"] = "Videos"
	case SearchUsers:
		variables["product"] = "People"
	}

	q := url.Values{}
	q.Set("variables", mapToJSONString(variables))
	q.Set("features", mapToJSONString(features))
	q.Set("fieldToggles", mapToJSONString(fieldToggles))
	req.URL.RawQuery = q.Encode()

	var timeline searchTimeline
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, err
	}
	return &timeline, nil
}

// FetchSearchTweets gets tweets for a given search query, via the Twitter frontend API
func (s *Scraper) FetchSearchTweets(query string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	timeline, err := s.getSearchTimeline(query, maxTweetsNbr, cursor)
	if err != nil {
		return nil, "", err
	}
	tweets, nextCursor := timeline.parseTweets()
	return tweets, nextCursor, nil
}

// FetchSearchProfiles gets users for a given search query, via the Twitter frontend API
func (s *Scraper) FetchSearchProfiles(query string, maxProfilesNbr int, cursor string) ([]*Profile, string, error) {
	timeline, err := s.getSearchTimeline(query, maxProfilesNbr, cursor)
	if err != nil {
		return nil, "", err
	}
	users, nextCursor := timeline.parseUsers()
	return users, nextCursor, nil
}
