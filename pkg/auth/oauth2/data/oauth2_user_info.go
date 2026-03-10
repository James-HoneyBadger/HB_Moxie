package data

import "github.com/James-HoneyBadger/HB_Moxie/pkg/core"

// OAuth2UserInfo represents the user info retrieved from OAuth 2.0 provider
type OAuth2UserInfo struct {
	UserName       string
	Email          string
	NickName       string
	LanguageCode   string
	CurrencyCode   string
	FirstDayOfWeek core.WeekDay
}
