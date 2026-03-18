package twitterscraper

import "net/url"

type ThreadCursor struct {
	FocalTweetID string
	ThreadID     string
	Cursor       string
	CursorType   string
}

func (s *Scraper) GetTweetReplies(id string, cursor string) ([]*Tweet, []*ThreadCursor, error) {
	req, err := s.newRequest("GET", "https://x.com/i/api/graphql/xIYgDwjboktoFeXe_fgacw/TweetDetail")
	if err != nil {
		return nil, nil, err
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

	if cursor != "" {
		variables["cursor"] = cursor
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

	var threads threadedConversation

	err = s.RequestAPI(req, &threads)
	if err != nil {
		return nil, nil, err
	}

	tweets, cursors := threads.parse(id)

	return tweets, cursors, nil
}
