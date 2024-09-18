package dotenv

import "os"

type TwitchConfig struct {
	AccessToken  string
	AppId        string
	AppSecret    string
	AuthToken    string
	Nonce        string
	RefreshToken string
}

func GetTwitchConfig() TwitchConfig {
	return TwitchConfig{
		AccessToken:  os.Getenv("USER_ACCESS_TOKEN"),
		AppId:        os.Getenv("APP_ID"),
		AppSecret:    os.Getenv("APP_SECRET"),
		Nonce:        os.Getenv("NONCE"),
		RefreshToken: os.Getenv("USER_REFRESH_TOKEN"),
	}
}

type ServerConfig struct {
	Host string
	Port string
}

func GetServerConfig() ServerConfig {
	return ServerConfig{
		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
	}
}

func DefaultServerConfig() {
	if _, found := os.LookupEnv("HOST"); !found {
		os.Setenv("HOST", "localhost")
	}
	if _, found := os.LookupEnv("PORT"); !found {
		os.Setenv("PORT", "7582")
	}
}
