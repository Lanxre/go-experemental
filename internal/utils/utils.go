package utils

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lanxre/go-experemental/internal/models"
)

func SendTrackUpdate(conn *websocket.Conn, track *models.CurrentlyPlaying) {
	update := map[string]interface{}{
		"event":   "track_change",
		"track":   track.Item.Name,
		"artists": GetArtistNames(track),
		"album":   track.Item.Album.Name,
		"time":    time.Now().Unix(),
	}

	if err := conn.WriteJSON(update); err != nil {
		log.Println("Send update error:", err)
	}
}

func GetArtistNames(track *models.CurrentlyPlaying) []string {
	names := make([]string, len(track.Item.Artists))
	for i, artist := range track.Item.Artists {
		names[i] = artist.Name
	}
	return names
}

func OptimizeUpdateTrack(track *models.CurrentlyPlaying){
	delay := 10 * time.Second
	if !track.IsPlaying {
		delay = 10 * time.Second
	}
	time.Sleep(delay)
}