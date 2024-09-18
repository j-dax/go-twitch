package auth

import (
	"helix/dotenv"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExchangeCode(t *testing.T) {
	// Mock the twitch api server
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "test_access_token",
			"token_type": "bearer",
			"scope": "user:read:email",
			"expires_in": 3600
		}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Override tokenUrl
	var twitchConfig dotenv.TwitchConfig
	tokenAuthUri := server.URL + "/oauth2/token"
	exchangeResponse, err := exchangeCode(twitchConfig, "test_code", "http://test_redirect", tokenAuthUri)
	if err != nil {
		t.Fatalf("exchangeCode() returned an error: %v", err)
	}
	if exchangeResponse.AccessToken != "test_access_token" {
		t.Errorf("Expected access token 'test_access_token', got %s", exchangeResponse.AccessToken)
	}
}

func TestValidateAccess(t *testing.T) {
	// Mock the twitch server
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer valid_token" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"client_id": "test_client_id",
			"login": "test_login",
			"user_id": "test_user_id",
			"scope": "user:read:email"
		}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	apiUri := server.URL + "/oauth2/validate"

	accessResponse, err := validateAccess("valid_token", apiUri)
	if err != nil {
		t.Fatalf("validateAccess() returned an error: %v", err)
	}
	if accessResponse == nil {
		t.Error("validateAccess() returned a nil response object")
	}
	if accessResponse.UserId != "test_login" {
		t.Errorf("Expected login 'test_login', got %s", accessResponse.UserId)
	}
}
