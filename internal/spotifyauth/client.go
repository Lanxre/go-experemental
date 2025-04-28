package spotifyauth

import (
	"context"
	"fmt"
	"net/http"
	"golang.org/x/oauth2"
	"github.com/imroc/req/v3"
	"github.com/lanxre/go-experemental/internal/models"
)

type SpotifyClient struct {
	http    *req.Client
	token   *oauth2.Token
	auth    *Authenticator
	ctx     context.Context
}

// NewClient creates a new authenticated Spotify client
func NewClient(auth *Authenticator) *SpotifyClient {
	return &SpotifyClient{
		http:  req.C(),
		token: auth.token,
		auth:  auth,
		ctx:   context.Background(),
	}
}

// WithContext sets the context for the client
func (c *SpotifyClient) WithContext(ctx context.Context) *SpotifyClient {
	c.ctx = ctx
	return c
}

func (s *SpotifyClient) SetAccessToken(token *oauth2.Token) {
	s.token = token
}


// GetCurrentTrack gets the user's currently playing track
func (c *SpotifyClient) GetCurrentTrack() (*models.CurrentlyPlaying, error) {
	var track models.CurrentlyPlaying

	resp, err := c.http.R().
		SetContext(c.ctx).
		SetBearerAuthToken(c.token.AccessToken).
		SetSuccessResult(&track).
		Get("https://api.spotify.com/v1/me/player/currently-playing")

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil // No content
	}

	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &track, nil
}

// RefreshToken refreshes the access token if expired
func (c *SpotifyClient) RefreshToken() error {
	token, err := c.auth.Exchange(c.ctx, c.auth.state, "")
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	c.token = token
	return nil
}