package twitterscraper

import (
	"net/url"
	"strings"
)

// GetTweetRetweeters returns users who retweeted a specific tweet
func (s *Scraper) GetTweetRetweeters(tweetId string, maxUsersNbr int, cursor string) ([]*Profile, string, error) {
	if maxUsersNbr > 200 {
		maxUsersNbr = 200
	}

	req, err := s.newRequest("GET", "https://x.com/i/api/graphql/uhTjAvG7nm0lyrfujroWUw/Retweeters")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"tweetId":                tweetId,
		"count":                  maxUsersNbr,
		"enableRanking":          true,
		"includePromotedContent": true,
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(graphqlWebClientFeatures()))
	req.URL.RawQuery = query.Encode()

	var timeline retweetersTimelineV2
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, "", err
	}

	users, nextCursor := timeline.parseUsers()

	if strings.HasPrefix(nextCursor, "0|") {
		nextCursor = ""
	}

	return users, nextCursor, nil
}
