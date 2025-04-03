package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SpeechRequest struct {
	Text      string `json:"text"`
	Voice     string `json:"voice"`
	Speed     int    `json:"speed"`
	Pitch     int    `json:"pitch"`
	Amplitude int    `json:"amplitude"`
}

type Job struct {
	ID        string
	Request   SpeechRequest
	FilePath  string
	Status    string
	Error     string
	CreatedAt time.Time
}

var (
	jobQueue  = make(chan Job, 10)
	jobs      = make(map[string]*Job)
	jobsMutex = sync.Mutex{}
)

func InitWorkerPool(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go worker()
	}
}

func worker() {
	for job := range jobQueue {
		jobsMutex.Lock()
		job.Status = "processing"
		jobsMutex.Unlock()

		filename, err := processSpeech(job.Request)
		jobsMutex.Lock()
		if err != nil {
			job.Status = "failed"
			job.Error = err.Error()
		} else {
			job.Status = "completed"
			job.FilePath = filename
		}
		jobsMutex.Unlock()
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

	jobID := uuid.New().String()
	job := Job{
		ID:        jobID,
		Request:   rq,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	jobsMutex.Lock()
	jobs[jobID] = &job
	jobsMutex.Unlock()

	jobQueue <- job

	response := map[string]string{"job_id": jobID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func JobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "Missing job_id parameter", http.StatusBadRequest)
		return
	}

	jobsMutex.Lock()
	job, exists := jobs[jobID]
	jobsMutex.Unlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if job.Status == "completed" {
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", job.FilePath))
		http.ServeFile(w, r, job.FilePath)
		_ = os.Remove(job.FilePath)
	} else {
		response := map[string]string{"status": job.Status, "error": job.Error}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func SetParameter(w http.ResponseWriter, r *http.Request) {
	var newParams SpeechRequest
	if err := json.NewDecoder(r.Body).Decode(&newParams); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newParams)
}
