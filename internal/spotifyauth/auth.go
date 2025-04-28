package spotifyauth

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/oauth2"
)

var (
	ErrMissingConfig     = errors.New("missing OAuth configuration")
	ErrInvalidState      = errors.New("invalid OAuth state")
	ErrTokenExchange     = errors.New("token exchange failed")
)

type Authenticator struct {
	token  *oauth2.Token
	config *oauth2.Config
	state  string
}

// New creates a new Spotify authenticator
func New(clientID, clientSecret, redirectURL string, scopes []string) *Authenticator {
	return &Authenticator{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.spotify.com/authorize",
				TokenURL: "https://accounts.spotify.com/api/token",
			},
		},
		state: generateState(),
	}
}

// NewFromEnv creates authenticator from environment variables
func NewFromEnv() (*Authenticator, error) {
	clientID := "9a10dd80bdcf420ca6aa9a909bc5b833"
	clientSecret := "6ec9b48448234d3ab537216358c15a34"
	redirectURL := "http://localhost:8080/api/spotify-callback"

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, ErrMissingConfig
	}

	return New(clientID, clientSecret, redirectURL, []string{
		"user-read-currently-playing",
		"user-read-playback-state",
	}), nil
}

// AuthURL returns the URL to redirect users for authentication
func (a *Authenticator) AuthURL() string {
	return a.config.AuthCodeURL(a.state)
}

// Exchange converts an authorization code into a token
func (a *Authenticator) Exchange(ctx context.Context, state, code string) (*oauth2.Token, error) {
	if state != a.state {
		return nil, ErrInvalidState
	}

	token, err := a.config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.Join(ErrTokenExchange, err)
	}
	a.token = token
	return token, nil
}

func (a *Authenticator) GetAccessToken() (string, error) {
	return a.token.AccessToken, nil
}

func (a *Authenticator) GetToken() (*oauth2.Token){
	return a.token
}

func (a *Authenticator) SetToken(token *oauth2.Token){
	a.token = token
}

// Client creates an authenticated HTTP client
func (a *Authenticator) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return a.config.Client(ctx, token)
}

func generateState() string {
	// In production, use a cryptographically secure random string
	return "random-state-string"
}