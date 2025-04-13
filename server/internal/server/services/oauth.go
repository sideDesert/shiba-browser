package services

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var OauthConfig = make(map[string]*oauth2.Config)

func createGoogleOAuthConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Scopes: []string{
			"openid",
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}
}
