package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

type SpeechRequest struct {
	Text      string `json:"text"`
	Voice     string `json:"voice"`
	Speed     int    `json:"speed"`
	Pitch     int    `json:"pitch"`
	Amplitude int    `json:"amplitude"`
}

type Job struct {
	Request  SpeechRequest
	Response chan string
	Error    chan error
}

var jobQueue chan Job

func InitWorkerPool(workerCount int) {
	jobQueue = make(chan Job, 10)

	for i := 0; i < workerCount; i++ {
		go worker()
	}
}

func worker() {
	for job := range jobQueue {
		filename, err := processSpeech(job.Request)
		if err != nil {
			job.Error <- err
		} else {
			job.Response <- filename
		}
	}
}

func processSpeech(rq SpeechRequest) (string, error) {
	tempFile := fmt.Sprintf("temp_%d.wav", os.Getpid())

	cmd := exec.Command("espeak-ng",
		"-v", rq.Voice,
		"-s", fmt.Sprintf("%d", rq.Speed),
		"-p", fmt.Sprintf("%d", rq.Pitch),
		"-a", fmt.Sprintf("%d", rq.Amplitude),
		"-w", tempFile,
		rq.Text)

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to synthesize speech: %w", err)
	}

	return tempFile, nil
}

func ToSpeech(w http.ResponseWriter, r *http.Request) {
	var rq SpeechRequest
	err := json.NewDecoder(r.Body).Decode(&rq)
	if err != nil {
		http.Error(w, "Couldn't decode request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if rq.Text == "" {
		http.Error(w, "Text field is required", http.StatusBadRequest)
		return
	}

	responseChan := make(chan string)
	errorChan := make(chan error)

	job := Job{Request: rq, Response: responseChan, Error: errorChan}

	jobQueue <- job

	select {
	case file := <-responseChan:
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Content-Disposition", "attachment; filename=speech.wav")
		http.ServeFile(w, r, file)
		_ = os.Remove(file)
	case err := <-errorChan:
		http.Error(w, "Failed to synthesize speech: "+err.Error(), http.StatusInternalServerError)
	}
}
