package twitterscraper

type AccountSettings struct {
	ScreenName            string `json:"screen_name"`
	Protected             bool   `json:"protected"`
	DisplaySensitiveMedia bool   `json:"display_sensitive_media"`
	Language              string `json:"language"`
	CountryCode           string `json:"country_code"`

	DiscoverableByEmail                          bool   `json:"discoverable_by_email"`
	DiscoverableByMobilePhone                    bool   `json:"discoverable_by_mobile_phone"`
	PersonalizedTrends                           bool   `json:"personalized_trends"`
	AllowMediaTagging                            string `json:"allow_media_tagging"`
	AllowContributorRequest                      string `json:"allow_contributor_request"`
	AllowAdsPersonalization                      bool   `json:"allow_ads_personalization"`
	AllowLoggedOutDevicePersonalization          bool   `json:"allow_logged_out_device_personalization"`
	AllowLocationHistoryPersonalization          bool   `json:"allow_location_history_personalization"`
	AllowSharingDataForThirdPartyPersonalization bool   `json:"allow_sharing_data_for_third_party_personalization"`
	AllowDmsFrom                                 string `json:"allow_dms_from"`
	AllowDmGroupsFrom                            string `json:"allow_dm_groups_from"`
	AddressBookLiveSyncEnabled                   bool   `json:"address_book_live_sync_enabled"`
	UniversalQualityFilteringEnabled             string `json:"universal_quality_filtering_enabled"`
	DmReceiptSetting                             string `json:"dm_receipt_setting"`
	AllowAuthenticatedPeriscopeRequests          bool   `json:"allow_authenticated_periscope_requests"`
	ProtectPasswordReset                         bool   `json:"protect_password_reset"`
	RequirePasswordLogin                         bool   `json:"require_password_login"`
	RequiresLoginVerification                    bool   `json:"requires_login_verification"`
	DmQualityFilter                              string `json:"dm_quality_filter"`
	AutoplayDisabled                             bool   `json:"autoplay_disabled"`
}

type Account struct {
	UserID         string `json:"user_id"`
	Name           string `json:"name"`
	ScreenName     string `json:"screen_name"`
	AvatarImageURL string `json:"avatar_image_url"`
	IsSuspended    bool   `json:"is_suspended"`
	IsVerified     bool   `json:"is_verified"`
	IsProtected    bool   `json:"is_protected"`
	IsAuthValid    bool   `json:"is_auth_valid"`
}

type AccountList struct {
	Users []Account `json:"users"`
}

func (s *Scraper) GetAccountSettings() (AccountSettings, error) {
	var settings AccountSettings
	req, err := s.newRequest("GET", "https://api.x.com/1.1/account/settings.json")
	if err != nil {
		return settings, err
	}

	err = s.RequestAPI(req, &settings)
	return settings, err
}

func (s *Scraper) GetAccountList() ([]Account, error) {
	var list AccountList
	req, err := s.newRequest("GET", "https://api.x.com/1.1/account/multi/list.json")
	if err != nil {
		return list.Users, err
	}

	err = s.RequestAPI(req, &list)
	return list.Users, err
}
