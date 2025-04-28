package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"time"
	"github.com/joho/godotenv"
	"github.com/gorilla/websocket"
	"github.com/lanxre/go-experemental/internal/spotifyauth"
	"github.com/lanxre/go-experemental/internal/models"
	"github.com/lanxre/go-experemental/internal/utils"
)

const (
	port               = ":8080"
	defaultPingPeriod  = 30 * time.Second
	playingDelay       = 1 * time.Second
	notPlayingDelay    = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

type App struct {
	spotifyAuth   *spotifyauth.Authenticator
	spotifyClient *spotifyauth.SpotifyClient
	clients       map[*websocket.Conn]bool
}

func main() {
	
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found: %v", err)
	}

	
	auth, err := spotifyauth.NewFromEnv()

	if err != nil {
		log.Fatalf("Failed to init spotify auth: %v", err)
	}

	app := &App{
		spotifyAuth: auth,
		spotifyClient: spotifyauth.NewClient(auth),
		clients: make(map[*websocket.Conn]bool),
	}

	http.HandleFunc("/", app.handleRoot)
	http.HandleFunc("/api/spotify-callback", app.handleSpotifyCallback)
	http.HandleFunc("/api/spotify-current-playing", app.handleCurrentPlaying)
	http.HandleFunc("/api/ws/spotify-current-playing", app.handleWsCurrentPlaying)

	fmt.Printf("Server starting on: [http://localhost%s]\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func (app *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, app.spotifyAuth.AuthURL(), http.StatusFound)
}

func (app *App) handleSpotifyCallback(w http.ResponseWriter, r *http.Request) {

	state := r.FormValue("state")
	code := r.FormValue("code")

	if state == "" || code == "" {
		response := models.SpotifyTokenResponse{
			Status:       "error",
			Error:        "invalid_request",
			ErrorMessage: "Missing state or code parameter",
			Timestamp: time.Now(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	token, err := app.spotifyAuth.Exchange(r.Context(), state, code)
	
	if err != nil {
		response := models.SpotifyTokenResponse{
			Status:       "error",
			Error:        "token_exchange_failed",
			ErrorMessage: err.Error(),
			Timestamp: time.Now(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}


	app.spotifyAuth.SetToken(token)
	response := models.SpotifyTokenResponse{
		Status:      "success",
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		ExpiresIn:   int(time.Until(token.Expiry).Seconds()),
		Scope:       fmt.Sprint(token.Extra("scope")),
		Timestamp:   time.Now(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (app *App) handleCurrentPlaying(w http.ResponseWriter, r *http.Request){

	token := app.spotifyAuth.GetToken()

	if token == nil {
		app.sendErrorResponse(w, http.StatusUnauthorized, "no_token", "No access token available")
		return
		
	}

	app.spotifyClient.SetAccessToken(token)
	track, err := app.spotifyClient.GetCurrentTrack()

	if err != nil {
        app.sendErrorResponse(w, http.StatusInternalServerError, "track_fetch_failed", err.Error())
        return
    }
    
    if track == nil || !track.IsPlaying || track.Item.Name == "" {
        app.sendErrorResponse(w, http.StatusOK, "no_track_playing", "No track is currently playing")
        return
    }
	var trackResponse = &models.TrackUpdate{
		ID:      track.Item.ID,
		Name:    track.Item.Name,
		Artists: utils.GetArtistNames(track),
		Album:   track.Item.Album.Name,
		Playing: track.IsPlaying,
		Time:    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trackResponse)
}

func (app *App) handleWsCurrentPlaying (w http.ResponseWriter, r *http.Request) {

	if !websocket.IsWebSocketUpgrade(r) {
		http.Error(w, "Not a websocket request", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	defer conn.Close()

	app.spotifyClient.SetAccessToken(app.spotifyAuth.GetToken())

	for {
		track, err := app.spotifyClient.GetCurrentTrack()

		if err != nil {
			conn.WriteJSON(map[string]interface{}{
				"track":   nil,
			})
		}

		utils.SendTrackUpdate(conn, track)
		utils.OptimizeUpdateTrack(track)
	}
}

func (app *App) sendErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	response := models.SpotifyTokenResponse{
		Status:       "error",
		Error:        errorCode,
		ErrorMessage: message,
		Timestamp:    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
