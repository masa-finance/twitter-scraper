package twitterscraper

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// GetTweets returns channel with tweets for a given user.
func (s *Scraper) GetTweets(ctx context.Context, user string, maxTweetsNbr int) <-chan *TweetResult {
	return getTweetTimeline(ctx, user, maxTweetsNbr, s.FetchTweets)
}

// GetTweetsAndReplies returns channel with tweets and replies for a given user.
func (s *Scraper) GetTweetsAndReplies(ctx context.Context, user string, maxTweetsNbr int) <-chan *TweetResult {
	return getTweetTimeline(ctx, user, maxTweetsNbr, s.FetchTweetsAndReplies)
}

// FetchTweets gets tweets for a given user, via the Twitter frontend API.
func (s *Scraper) FetchTweets(user string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	userID, err := s.GetUserIDByScreenName(user)
	if err != nil {
		return nil, "", err
	}

	return s.FetchTweetsByUserID(userID, maxTweetsNbr, cursor)
}

// FetchTweetsAndReplies gets tweets and replies for a given user, via the Twitter frontend API.
func (s *Scraper) FetchTweetsAndReplies(user string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	userID, err := s.GetUserIDByScreenName(user)
	if err != nil {
		return nil, "", err
	}

	return s.FetchTweetsAndRepliesByUserID(userID, maxTweetsNbr, cursor)
}

// FetchTweetsAndRepliesByUserID gets tweets and replies for a given userID, via the Twitter frontend GraphQL API.
func (s *Scraper) FetchTweetsAndRepliesByUserID(userID string, maxReplysNbr int, cursor string) ([]*Tweet, string, error) {
	if maxReplysNbr > 200 {
		maxReplysNbr = 200
	}

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/bt4TKuFz4T7Ckk-VvQVSow/UserTweetsAndReplies")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"userId":                                 userID,
		"count":                                  maxReplysNbr,
		"includePromotedContent":                 false,
		"withQuickPromoteEligibilityTweetFields": false,
		"withVoice":                              true,
		"withV2Timeline":                         true,
	}

	features := map[string]interface{}{
		"rweb_tipjar_consumption_enabled":                                         true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"articles_preview_enabled":                                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"creator_subscriptions_quote_tweet_preview_enabled":                       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"rweb_video_timestamps_enabled":                                           true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	req.URL.RawQuery = query.Encode()

	var timeline timelineV2
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, "", err
	}

	tweets, nextCursor := timeline.parseTweets()
	return tweets, nextCursor, nil
}

// FetchTweetsByUserID gets tweets for a given userID, via the Twitter frontend GraphQL API.
func (s *Scraper) FetchTweetsByUserID(userID string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	if maxTweetsNbr > 200 {
		maxTweetsNbr = 200
	}

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/UGi7tjRPr-d_U3bCPIko5Q/UserTweets")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"userId":                                 userID,
		"count":                                  maxTweetsNbr,
		"includePromotedContent":                 false,
		"withQuickPromoteEligibilityTweetFields": false,
		"withVoice":                              true,
		"withV2Timeline":                         true,
	}
	features := map[string]interface{}{
		"rweb_lists_timeline_redesign_enabled":                              true,
		"responsive_web_graphql_exclude_directive_enabled":                  true,
		"verified_phone_label_enabled":                                      false,
		"creator_subscriptions_tweet_preview_api_enabled":                   true,
		"responsive_web_graphql_timeline_navigation_enabled":                true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled": false,
		"tweetypie_unmention_optimization_enabled":                          true,
		"vibe_api_enabled":                                                        true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": false,
		"interactive_text_enabled":                                                true,
		"responsive_web_text_conversations_enabled":                               false,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                false,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	req.URL.RawQuery = query.Encode()

	var timeline timelineV2
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, "", err
	}

	tweets, nextCursor := timeline.parseTweets()
	return tweets, nextCursor, nil
}

// FetchTweetsByUserIDLegacy gets tweets for a given userID, via the Twitter frontend legacy API.
func (s *Scraper) FetchTweetsByUserIDLegacy(userID string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	if maxTweetsNbr > 200 {
		maxTweetsNbr = 200
	}

	req, err := s.newRequest("GET", "https://api.twitter.com/2/timeline/profile/"+userID+".json")
	if err != nil {
		return nil, "", err
	}

	q := req.URL.Query()
	q.Add("count", strconv.Itoa(maxTweetsNbr))
	q.Add("userId", userID)
	if cursor != "" {
		q.Add("cursor", cursor)
	}
	req.URL.RawQuery = q.Encode()

	var timeline timelineV1
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, "", err
	}

	tweets, nextCursor := timeline.parseTweets()
	return tweets, nextCursor, nil
}

// GetTweet get a single tweet by ID.
func (s *Scraper) GetTweet(id string) (*Tweet, error) {
	if s.isLogged {
		req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/VWFGPVAGkZMGRKGe3GFFnA/TweetDetail")
		if err != nil {
			return nil, err
		}

		variables := map[string]interface{}{
			"focalTweetId":                           id,
			"with_rux_injections":                    false,
			"includePromotedContent":                 true,
			"withCommunity":                          true,
			"withQuickPromoteEligibilityTweetFields": true,
			"withBirdwatchNotes":                     true,
			"withVoice":                              true,
			"withV2Timeline":                         true,
		}

		features := map[string]interface{}{
			"rweb_lists_timeline_redesign_enabled":                                    true,
			"responsive_web_graphql_exclude_directive_enabled":                        true,
			"verified_phone_label_enabled":                                            false,
			"creator_subscriptions_tweet_preview_api_enabled":                         true,
			"responsive_web_graphql_timeline_navigation_enabled":                      true,
			"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
			"tweetypie_unmention_optimization_enabled":                                true,
			"responsive_web_edit_tweet_api_enabled":                                   true,
			"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
			"view_counts_everywhere_api_enabled":                                      true,
			"longform_notetweets_consumption_enabled":                                 true,
			"tweet_awards_web_tipping_enabled":                                        false,
			"freedom_of_speech_not_reach_fetch_enabled":                               true,
			"standardized_nudges_misinfo":                                             true,
			"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": false,
			"longform_notetweets_rich_text_read_enabled":                              true,
			"longform_notetweets_inline_media_enabled":                                true,
			"responsive_web_enhance_cards_enabled":                                    false,
		}

		query := url.Values{}
		query.Set("variables", mapToJSONString(variables))
		query.Set("features", mapToJSONString(features))
		req.URL.RawQuery = query.Encode()

		var conversation threadedConversation

		err = s.RequestAPI(req, &conversation)

		if err != nil {
			return nil, err
		}

		tweets, _ := conversation.parse(id)
		for _, tweet := range tweets {
			if tweet.ID == id {
				return tweet, nil
			}
		}
	} else {
		req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/xBtHv5-Xsk268T5ng_OGNg/TweetResultByRestId")
		if err != nil {
			return nil, err
		}
		variables := map[string]interface{}{
			"tweetId":                id,
			"withCommunity":          false,
			"includePromotedContent": false,
			"withVoice":              false,
		}

		features := map[string]interface{}{
			"creator_subscriptions_tweet_preview_api_enabled":                         true,
			"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
			"tweetypie_unmention_optimization_enabled":                                true,
			"responsive_web_edit_tweet_api_enabled":                                   true,
			"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
			"view_counts_everywhere_api_enabled":                                      true,
			"longform_notetweets_consumption_enabled":                                 true,
			"responsive_web_twitter_article_tweet_consumption_enabled":                true,
			"tweet_awards_web_tipping_enabled":                                        false,
			"freedom_of_speech_not_reach_fetch_enabled":                               true,
			"standardized_nudges_misinfo":                                             true,
			"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
			"rweb_video_timestamps_enabled":                                           true,
			"longform_notetweets_rich_text_read_enabled":                              true,
			"longform_notetweets_inline_media_enabled":                                true,
			"responsive_web_graphql_exclude_directive_enabled":                        true,
			"verified_phone_label_enabled":                                            false,
			"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
			"responsive_web_graphql_timeline_navigation_enabled":                      true,
			"responsive_web_enhance_cards_enabled":                                    false,
		}

		fieldToggles := map[string]interface{}{"withArticleRichContentState": true}

		query := url.Values{}
		query.Set("variables", mapToJSONString(variables))
		query.Set("features", mapToJSONString(features))
		query.Set("fieldToggles", mapToJSONString(fieldToggles))
		req.URL.RawQuery = query.Encode()

		var result tweetResult

		err = s.RequestAPI(req, &result)

		if err != nil {
			return nil, err
		}

		tweet := result.parse()
		return tweet, nil
	}
	return nil, fmt.Errorf("tweet with ID %s not found", id)
}

type homeEntry struct {
	EntryId   string `json:"entryId"`
	SortIndex string `json:"sortIndex"`
	Content   struct {
		EntryType   string `json:"entryType"`
		ItemContent struct {
			ItemType     string `json:"itemType"`
			TweetResults struct {
				Result result `json:"result"`
			} `json:"tweet_results"`
		} `json:"itemContent"`
		Cursor     string `json:"value"`
		CursorType string `json:"cursorType"`
	} `json:"content"`
}

// timeline v2 JSON object
type homeTimeline struct {
	Data struct {
		Home struct {
			HomeTimeline struct {
				Instructions []struct {
					Entries []homeEntry `json:"entries"`
					Type    string      `json:"type"`
				} `json:"instructions"`
				Metadata struct {
					SribeConfig []struct {
						Page string `json:"page"`
					} `json:"scribe_config"`
				} `json:"metadata"`
			} `json:"home_timeline_urt"`
		} `json:"home"`
	} `json:"data"`
}

func (timeline *homeTimeline) parseTweets() ([]*Tweet, string) {
	var cursor string
	var tweets []*Tweet
	for _, instruction := range timeline.Data.Home.HomeTimeline.Instructions {
		for _, entry := range instruction.Entries {
			if entry.Content.CursorType == "Bottom" {
				cursor = entry.Content.Cursor
			} else if entry.Content.ItemContent.TweetResults.Result.Typename == "Tweet" {
				if tweet := entry.Content.ItemContent.TweetResults.Result.parse(); tweet != nil {
					tweets = append(tweets, tweet)
				}
			}
		}
	}
	return tweets, cursor
}

// GetHomeTweets returns channel with tweets from home timeline
func (s *Scraper) GetHomeTweets(ctx context.Context, maxTweetsNbr int) <-chan *TweetResult {
	return getTweetTimeline(ctx, "", maxTweetsNbr, s.fetchHomeTweets)
}

func (s *Scraper) FetchHomeTweets(maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	return s.fetchHomeTweets("", maxTweetsNbr, cursor)
}

// FetchHomeTweets gets tweets from home timline, via the Twitter frontend API.
func (s *Scraper) fetchHomeTweets(_ string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	if maxTweetsNbr > 200 {
		maxTweetsNbr = 200
	}

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/9EwYy8pLBOSFlEoSP2STiQ/HomeLatestTimeline")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"count":                                  maxTweetsNbr,
		"includePromotedContent":                 true,
		"withQuickPromoteEligibilityTweetFields": true,
		"requestContext":                         "launch",
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	features := map[string]interface{}{
		"rweb_tipjar_consumption_enabled":                                         true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"articles_preview_enabled":                                                true,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"creator_subscriptions_quote_tweet_preview_enabled":                       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"rweb_video_timestamps_enabled":                                           true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	req.Header.Set("content-type", "application/json")

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	req.URL.RawQuery = query.Encode()

	var timeline homeTimeline
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, "", err
	}

	tweets, nextCursor := timeline.parseTweets()
	return tweets, nextCursor, nil
}

// GetForYouTweets returns channel with tweets from for you timeline
func (s *Scraper) GetForYouTweets(ctx context.Context, maxTweetsNbr int) <-chan *TweetResult {
	return getTweetTimeline(ctx, "", maxTweetsNbr, s.fetchForYouTweets)
}

func (s *Scraper) FetchForYouTweets(maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	return s.fetchForYouTweets("", maxTweetsNbr, cursor)
}

// FetchForYouTweets gets tweets from for you timline, via the Twitter frontend API.
func (s *Scraper) fetchForYouTweets(_ string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	if maxTweetsNbr > 200 {
		maxTweetsNbr = 200
	}

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/1u0Wlkw6Ru1NwBUD-pDiww/HomeTimeline")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"count":                  maxTweetsNbr,
		"includePromotedContent": true,
		"latestControlAvailable": true,
		"requestContext":         "launch",
		"withCommunity":          true,
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	features := map[string]interface{}{
		"rweb_tipjar_consumption_enabled":                                         true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"articles_preview_enabled":                                                true,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"creator_subscriptions_quote_tweet_preview_enabled":                       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"rweb_video_timestamps_enabled":                                           true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	req.Header.Set("content-type", "application/json")

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	req.URL.RawQuery = query.Encode()

	var timeline homeTimeline
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, "", err
	}

	tweets, nextCursor := timeline.parseTweets()
	return tweets, nextCursor, nil
}
