package twitterscraper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

const homeTimelineURL = "https://x.com/i/api/graphql/L8Lb9oomccM012S7fQ-QKA/HomeTimeline"

// HomeTimelineGraphQLURL is the Home (Following) feed endpoint (Mar 2026).
const HomeTimelineGraphQLURL = homeTimelineURL

type homeTimelineInstruction struct {
	Type    string  `json:"type"`
	Entries []entry `json:"entries"`
}

func parseHomeInstructions(instructions []homeTimelineInstruction) ([]*Tweet, string) {
	var cursor string
	var tweets []*Tweet
	for _, instruction := range instructions {
		for _, ent := range instruction.Entries {
			if ent.Content.CursorType == "Bottom" {
				cursor = ent.Content.Value
				continue
			}
			tn := ent.Content.ItemContent.TweetResults.Result.Typename
			if tn == "Tweet" || tn == "TweetWithVisibilityResults" {
				if tw := ent.Content.ItemContent.TweetResults.Result.parse(); tw != nil {
					tweets = append(tweets, tw)
				}
			}
			for _, it := range ent.Content.Items {
				tn2 := it.Item.ItemContent.TweetResults.Result.Typename
				if tn2 == "Tweet" || tn2 == "TweetWithVisibilityResults" {
					if tw := it.Item.ItemContent.TweetResults.Result.parse(); tw != nil {
						tweets = append(tweets, tw)
					}
				}
			}
		}
	}
	return tweets, cursor
}

func decodeHomeTimelineInstructions(raw json.RawMessage) []homeTimelineInstruction {
	if len(raw) == 0 {
		return nil
	}
	var direct struct {
		Instructions []homeTimelineInstruction `json:"instructions"`
	}
	if err := json.Unmarshal(raw, &direct); err == nil && len(direct.Instructions) > 0 {
		return direct.Instructions
	}
	var nested struct {
		Timeline struct {
			Instructions []homeTimelineInstruction `json:"instructions"`
		} `json:"timeline"`
	}
	if err := json.Unmarshal(raw, &nested); err == nil && len(nested.Timeline.Instructions) > 0 {
		return nested.Timeline.Instructions
	}
	return nil
}

// FetchHomeTweets returns tweets from the Home (Following) timeline. Requires login.
func (s *Scraper) FetchHomeTweets(count int, cursor string) ([]*Tweet, string, error) {
	if count < 1 {
		count = 20
	}
	if count > 100 {
		count = 100
	}

	req, err := http.NewRequest("GET", homeTimelineURL, nil)
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"count":                  count,
		"includePromotedContent": true,
		"withCommunity":          true,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	q := url.Values{}
	q.Set("variables", mapToJSONString(variables))
	q.Set("features", mapToJSONString(graphqlWebClientFeatures()))
	req.URL.RawQuery = q.Encode()

	var wrap struct {
		Data struct {
			Home struct {
				HomeTimelineUrt json.RawMessage `json:"home_timeline_urt"`
			} `json:"home"`
		} `json:"data"`
	}
	if err := s.RequestAPI(req, &wrap); err != nil {
		return nil, "", err
	}

	inst := decodeHomeTimelineInstructions(wrap.Data.Home.HomeTimelineUrt)
	if len(inst) == 0 {
		return []*Tweet{}, "", nil
	}
	tweets, next := parseHomeInstructions(inst)
	return tweets, next, nil
}

// GetHomeTweets streams home timeline tweets up to maxTweetsNbr. Requires login.
func (s *Scraper) GetHomeTweets(ctx context.Context, maxTweetsNbr int) <-chan *TweetResult {
	return getTweetTimeline(ctx, "", maxTweetsNbr, func(_ string, max int, c string) ([]*Tweet, string, error) {
		page := max
		if page > 100 {
			page = 100
		}
		return s.FetchHomeTweets(page, c)
	})
}
