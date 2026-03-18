# Changelog

## Unreleased

- Rotated all internal GraphQL query IDs to match **[d60/twikit](https://github.com/d60/twikit)** `main` (`client/gql.py`): `SearchTimeline`, `UserTweets`, `UserTweetsAndReplies`, `TweetDetail`, `TweetResultByRestId`, `Retweeters`, `UserByScreenName`, `UserByRestId`. X serves multiple bundle versions; if an endpoint 404s, try refreshing IDs from twikit or DevTools.

## v0.0.13

01.10.2024

- Added methods `GetTweetsAndReplies`, `FetchTweetsAndReplies`, `FetchTweetsAndRepliesByUserID` thanks to @thewh1teagle
- Added methods `GetAccountSettings` and `GetAccountList` thanks to @thewh1teagle
- Added methods `GetUserAgent` and `SetUserAgent`

## v0.0.12

09.08.2024

- Added method `GetProfileByID`

## v0.0.11

05.08.2024

- Added method `GetTweetRetweeters`

## v0.0.10

01.08.2024

- Added method `GetTweetReplies`

## v0.0.9

24.07.2024

- Added method `GetSpace`
- Added methods `LikeTweet`, `UnlikeTweet`

## v0.0.8

09.07.2024

- Added methods `CreateTweet`, `DeleteTweet`
- Added methods `CreateRetweet`, `DeleteRetweet`
- Added methods `GetHomeTweets`, `GetForYouTweets`

## v0.0.7

26.04.2024

- Fixed nsfw `GetTweet`, `FetchTweets`, `FetchSearchTweets`
- Added HSLSURL property to video

## v0.0.6

09.03.2024

- Added method `UploadMedia`
- Added type `AuthToken`
- Medias can now be attached to scheduled tweets
- Fixed error caused by weird status codes returned by twitter

## v0.0.5

08.03.2024

- Fixed `GetTweet` using `TweetResultByRestId` endpoint for anon users
- Fixed added accidentaly removed IsPrivate property to user

## v0.0.4

23.02.2024

- Added methods `FetchScheduledTweets`, `DeleteScheduledTweets`, `CreateScheduledTweets`

## v0.0.3

21.02.2024

- Added methods `GetBookmarks`, `FetchBookmarks`
- Added methods `FetchFollowing`, `FetchFollowers`

## v0.0.2

20.02.2024

- Added methods `GetMediaTweets`, `FetchMediaTweets`, `FetchMediaTweetsByUserID`
- Fixed anon socks5 proxy

## v0.0.1

28.01.2024

- `LoginOpenAccount` now returns `OpenAccount` object
- Added method `WithOpenAccount` to login with previously created open accounts

## v0.0.0

28.01.2024

- Forked repo
