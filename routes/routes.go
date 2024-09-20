package routes

import (
	"fmt"
	"helix/auth"
	"helix/dotenv"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var store NonceStore

// Create a random string of length `n`
func RandString(n int) string {
	randString := make([]byte, n)
	lower := "abcdefghijklmnopqrstuvwxyz"
	nums := "1234567890"
	chars := []byte(lower + strings.ToUpper(lower) + nums)

	i := 0
	for i < n {
		n := rand.IntN(len(chars))
		randString[i] = chars[n]
		i++
	}

	return string(randString)
}

type NonceStore struct {
	nonce string
}

// User initiates authorizing this application by visiting this route
func handleAuth(logger *log.Logger, nonceStore *NonceStore) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Println(r.URL.String())
			serverConfig := dotenv.GetServerConfig()
			twitchConfig := dotenv.GetTwitchConfig()
			redirectUri := fmt.Sprintf("http://%s:%s/oauth2/authorize", serverConfig.Host, serverConfig.Port)
			nonceStore.nonce = RandString(32)
			uri := url.URL{
				Scheme: "https",
				Host:   "id.twitch.tv",
				Path:   "/oauth2/authorize",
				RawQuery: url.Values{
					"client_id":     {twitchConfig.AppId},
					"redirect_uri":  {redirectUri},
					"response_type": {"code"},
					"scope":         {"user:read:chat user:write:chat"},
					// the nonce verifies that communication between twitch and this application are not being tampered with
					"state": {nonceStore.nonce},
				}.Encode(),
			}
			http.Redirect(w, r, uri.String(), http.StatusSeeOther)
		},
	)
}

// This route is called by twitch directly to pass an authorization code.
func handleAuthCallback(logger *log.Logger, nonceStore *NonceStore) http.Handler {
	serverConfig := dotenv.GetServerConfig()
	redirectUri := fmt.Sprintf("http://%s:%s/oauth2/authorize", serverConfig.Host, serverConfig.Port)
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Println(r.URL.String())
			prevNonce := nonceStore.nonce
			nonceStore.nonce = ""
			nonce := r.URL.Query().Get("state")
			if nonce != prevNonce {
				http.Error(w, "Nonce state does not match. Try again.", http.StatusBadRequest)
				return
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				http.Error(w, "Missing authorization code", http.StatusBadRequest)
				return
			}

			// exchange auth code for access and refresh token
			exchangeResponse, err := auth.ExchangeCode(dotenv.GetTwitchConfig(), code, redirectUri)
			if err != nil {
				logger.Println("Failed to exchange auth code: ", err.Error())
				http.Error(w, "Failed to exchange auth code for token. Try again.", http.StatusInternalServerError)
				return
			}

			os.Setenv("USER_ACCESS_TOKEN", exchangeResponse.AccessToken)
			os.Setenv("USER_REFRESH_TOKEN", exchangeResponse.RefreshToken)
			dotenv.Save(".env", []string{
				"APP_ID",
				"APP_SECRET",
				"USER_ACCESS_TOKEN",
				"USER_REFRESH_TOKEN",
			})

			logger.Println("Updated token")
		},
	)
}

func AddRoutes(
	mux *http.ServeMux,
	logger *log.Logger,
) {
	mux.Handle("/login", handleAuth(logger, &store))
	mux.Handle("/oauth2/authorize", handleAuthCallback(logger, &store))
	mux.Handle("/", http.NotFoundHandler())
}
