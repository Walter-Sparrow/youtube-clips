package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/rs/cors"
)

const youtubeUrl = "https://www.youtube.com/watch?v="
const clipsDir = "./clips"

type ClipRequest struct {
	VideoId  string `json:"videoId"`
	Start    string `json:"start"`
	Duration string `json:"duration"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /clip", func(w http.ResponseWriter, r *http.Request) {
		var req ClipRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = clip(req.VideoId, req.Start, req.Duration)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/clips/", http.StripPrefix("/clips/", http.FileServer(http.Dir(clipsDir))))
	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":8080", handler)
}

func clip(videoId string, start, duration string) error {
	var out, err bytes.Buffer
	youtubeDl := exec.Command("yt-dlp", "-g", youtubeUrl+videoId, "--skip-download")
	youtubeDl.Stdout = &out
	youtubeDl.Stderr = &err

	youtubeDl.Run()
	if err.Len() > 0 {
		log.Fatalf("Failed to run youtube-dl: %s", err.String())
		return fmt.Errorf("Failed to run youtube-dl: %s", err.String())
	}

	urls := strings.Split(out.String(), "\n")
	videoUrl := urls[0]
	audioUrl := urls[1]
	log.Printf("videoUrl: %s", videoUrl)
	log.Printf("audioUrl: %s", audioUrl)

	ffmpeg := exec.Command("ffmpeg",
		"-y",
		"-i", videoUrl,
		"-i", audioUrl,
		"-ss", start,
		"-t", duration,
		"-c", "copy",
		clipsDir+"/"+videoId+".mp4",
		"-loglevel", "error",
	)
	ffmpeg.Stderr = &err

	ffmpeg.Run()
	if err.Len() > 0 {
		log.Fatalf("Failed to run ffmpeg: %s", err.String())
		return fmt.Errorf("Failed to run ffmpeg: %s", err.String())
	}
	return nil
}
