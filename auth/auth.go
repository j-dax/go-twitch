// A collection of Twitch utilities to lookup and authenticate
package auth

import (
	"encoding/json"
	"fmt"
	"helix/dotenv"
	"io"
	"net/http"
	"net/url"
)

type AccessJsonResponse struct {
	ClientId  string   `json:"client_id"`
	ExpiresIn int      `json:"expires_in"`
	Login     string   `json:"login"`
	Scopes    []string `json:"scopes"`
	UserId    string   `json:"user_id"`
}

// Test if the given access token is valid. Twitch encourages checking this every hour.
func ValidateAccess(accessToken string) (*AccessJsonResponse, error) {
	return validateAccess(accessToken, "https://id.twitch.tv/oauth2/validate")
}

func validateAccess(accessToken, validateUri string) (*AccessJsonResponse, error) {
	// access token should be nonempty
	if accessToken == "" {
		return nil, fmt.Errorf("Access token is not set")
	}
	// validate with the api
	req, err := http.NewRequest("GET", validateUri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to validate token: %s", resp.Status)
	}

	var accessResponse AccessJsonResponse
	err = json.Unmarshal(body, &accessResponse)
	if err != nil {
		return nil, err
	}

	return &accessResponse, nil
}

type ExchangeJsonResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	Scope        []string `json:"scope"`
	ExpiresIn    int      `json:"expires_in"`
}

func ExchangeCode(twitchConfig dotenv.TwitchConfig, code, redirectUri string) (*ExchangeJsonResponse, error) {
	return exchangeCode(twitchConfig, code, redirectUri, "https://id.twitch.tv/oauth2/token")
}

// Exchange the authorization code for an access and refresh token
func exchangeCode(twitchConfig dotenv.TwitchConfig, code, redirectUri, tokenAuthUri string) (*ExchangeJsonResponse, error) {
	// send POST request to twitch's token endpoint to get tokens
	resp, err := http.PostForm(tokenAuthUri, url.Values{
		"client_id":     {twitchConfig.AppId},
		"client_secret": {twitchConfig.AppSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectUri},
	})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to exchange auth code %s:", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonResponse ExchangeJsonResponse
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}
	return &jsonResponse, nil
}
