package twitterscraper

import (
	"context"
	"fmt"
	"net/url"
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

	// Live x.com UserTweetsAndReplies (Mar 2026).
	req, err := s.newRequest("GET", "https://x.com/i/api/graphql/zedqO5hg41Ox6UeAKsWWzA/UserTweetsAndReplies")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"userId":                 userID,
		"count":                  maxReplysNbr,
		"includePromotedContent": true,
		"withCommunity":          true,
		"withVoice":              true,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(graphqlWebClientFeatures()))
	query.Set("fieldToggles", mapToJSONString(map[string]interface{}{"withArticlePlainText": false}))
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

	// Hash + payload from live x.com UserTweets (Mar 2026).
	req, err := s.newRequest("GET", "https://x.com/i/api/graphql/O0epvwaQPUx-bT9YlqlL6w/UserTweets")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"userId":                                 userID,
		"count":                                  maxTweetsNbr,
		"includePromotedContent":                 true,
		"withQuickPromoteEligibilityTweetFields": true,
		"withVoice":                              true,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(graphqlWebClientFeatures()))
	query.Set("fieldToggles", mapToJSONString(map[string]interface{}{"withArticlePlainText": false}))
	req.URL.RawQuery = query.Encode()

	var timeline timelineV2
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
		req, err := s.newRequest("GET", "https://x.com/i/api/graphql/xIYgDwjboktoFeXe_fgacw/TweetDetail")
		if err != nil {
			return nil, err
		}

		variables := map[string]interface{}{
			"focalTweetId":                           id,
			"referrer":                               "tweet",
			"with_rux_injections":                    false,
			"rankingMode":                            "Relevance",
			"includePromotedContent":                 true,
			"withCommunity":                          true,
			"withQuickPromoteEligibilityTweetFields": true,
			"withBirdwatchNotes":                     true,
			"withVoice":                              true,
		}

		fieldToggles := map[string]interface{}{
			"withArticleRichContentState": true,
			"withArticlePlainText":        false,
			"withArticleSummaryText":      true,
			"withArticleVoiceOver":        true,
			"withGrokAnalyze":             false,
			"withDisallowedReplyControls": false,
		}

		query := url.Values{}
		query.Set("variables", mapToJSONString(variables))
		query.Set("features", mapToJSONString(graphqlWebClientFeatures()))
		query.Set("fieldToggles", mapToJSONString(fieldToggles))
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
		// Use TweetResultByRestId for guest/unauthenticated requests
		req, err := s.newRequest("GET", "https://api.x.com/graphql/zy39CwTyYhU-_0LP7dljjg/TweetResultByRestId")
		if err != nil {
			return nil, err
		}

		variables := map[string]interface{}{
			"tweetId":                id,
			"withCommunity":          false,
			"includePromotedContent": false,
			"withVoice":              false,
		}

		query := url.Values{}
		query.Set("variables", mapToJSONString(variables))
		query.Set("features", mapToJSONString(map[string]interface{}{
			"creator_subscriptions_tweet_preview_api_enabled":                         true,
			"premium_content_api_read_enabled":                                        false,
			"communities_web_enable_tweet_community_results_fetch":                    true,
			"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
			"responsive_web_grok_analyze_button_fetch_trends_enabled":                 false,
			"responsive_web_grok_analyze_post_followups_enabled":                      false,
			"responsive_web_jetfuel_frame":                                            true,
			"responsive_web_grok_share_attachment_enabled":                            true,
			"responsive_web_grok_annotations_enabled":                                 true,
			"articles_preview_enabled":                                                true,
			"responsive_web_edit_tweet_api_enabled":                                   true,
			"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
			"view_counts_everywhere_api_enabled":                                      true,
			"longform_notetweets_consumption_enabled":                                 true,
			"responsive_web_twitter_article_tweet_consumption_enabled":                true,
			"tweet_awards_web_tipping_enabled":                                        false,
			"content_disclosure_indicator_enabled":                                    true,
			"content_disclosure_ai_generated_indicator_enabled":                       true,
			"responsive_web_grok_show_grok_translated_post":                           false,
			"responsive_web_grok_analysis_button_from_backend":                        true,
			"post_ctas_fetch_enabled":                                                 true,
			"freedom_of_speech_not_reach_fetch_enabled":                               true,
			"standardized_nudges_misinfo":                                             true,
			"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
			"longform_notetweets_rich_text_read_enabled":                              true,
			"longform_notetweets_inline_media_enabled":                                false,
			"profile_label_improvements_pcf_label_in_post_enabled":                    true,
			"responsive_web_profile_redirect_enabled":                                 false,
			"rweb_tipjar_consumption_enabled":                                         false,
			"verified_phone_label_enabled":                                            false,
			"responsive_web_grok_image_annotation_enabled":                            true,
			"responsive_web_grok_imagine_annotation_enabled":                          true,
			"responsive_web_grok_community_note_auto_translation_is_enabled":          false,
			"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
			"responsive_web_graphql_timeline_navigation_enabled":                      true,
			"responsive_web_enhance_cards_enabled":                                    false,
		}))
		query.Set("fieldToggles", mapToJSONString(map[string]interface{}{
			"withArticleRichContentState": true,
			"withArticlePlainText":        false,
			"withArticleSummaryText":      true,
			"withArticleVoiceOver":        true,
			"withGrokAnalyze":             false,
			"withDisallowedReplyControls": false,
		}))
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
